package app

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const rootID = "root"

const (
	maxUploadRequestBytes = 256 << 20
	maxUploadFileBytes    = 128 << 20
	maxUploadFiles        = 16
	maxShareBytes         = 10 << 30
	minFreeDiskBytes      = 1 << 30
	staleUploadAge        = 24 * time.Hour
)

type Config struct {
	Address  string
	ShareDir string
	DataDir  string
	WebDir   string
	Password string
}

type App struct {
	config              Config
	root                string
	tempDir             string
	rootM               sync.RWMutex
	db                  *sql.DB
	server              *http.Server
	handler             http.Handler
	sessions            map[string]time.Time
	sessionM            sync.RWMutex
	trashFn             func(string) error
	uploads             chan struct{}
	uploadM             sync.Mutex
	quotaM              sync.Mutex
	reservedUploadBytes int64
}

type fileDTO struct {
	ID       string    `json:"id"`
	ParentID string    `json:"parentId,omitempty"`
	Name     string    `json:"name"`
	Kind     string    `json:"kind"`
	MIME     string    `json:"mime"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

func New(config Config) (*App, error) {
	if strings.TrimSpace(config.Password) == "" {
		return nil, errors.New("password is required; set it in config.local.json")
	}
	if err := validateListenAddress(config.Address); err != nil {
		return nil, err
	}
	root, err := canonicalDirectory(config.ShareDir)
	if err != nil {
		return nil, fmt.Errorf("create share directory: %w", err)
	}
	data, err := canonicalDirectory(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	if pathsOverlap(root, data) {
		return nil, errors.New("share directory and application data directory must not overlap")
	}
	config.ShareDir, config.DataDir = root, data
	tempDir := filepath.Join(root, ".wifi-share-tmp")
	if err := os.MkdirAll(tempDir, 0o700); err != nil {
		return nil, fmt.Errorf("create upload temporary directory: %w", err)
	}
	_ = cleanupStaleUploads(tempDir, time.Now())

	db, err := sql.Open("sqlite", filepath.Join(data, "wifi-share.db"))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`
		PRAGMA journal_mode=WAL;
		CREATE TABLE IF NOT EXISTS operations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			item_id TEXT NOT NULL,
			details TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL
		);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize sqlite: %w", err)
	}

	app := &App{
		config: config, root: root, tempDir: tempDir, db: db,
		sessions: make(map[string]time.Time),
		trashFn:  moveToSystemTrash,
		uploads:  make(chan struct{}, 2),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.health)
	mux.HandleFunc("GET /api/auth/status", app.authStatus)
	mux.HandleFunc("POST /api/auth/login", app.login)
	mux.Handle("DELETE /api/auth/session", app.requireAuth(http.HandlerFunc(app.logout)))
	mux.HandleFunc("GET /api/files", app.listFiles)
	mux.Handle("POST /api/files/{id}/upload", app.requireAuth(http.HandlerFunc(app.upload)))
	mux.Handle("POST /api/files/{id}/folders", app.requireAuth(http.HandlerFunc(app.createFolder)))
	mux.HandleFunc("GET /api/files/{id}/content", app.getContent)
	mux.Handle("PUT /api/files/{id}/content", app.requireAuth(http.HandlerFunc(app.putContent)))
	mux.Handle("PATCH /api/files/{id}", app.requireAuth(http.HandlerFunc(app.rename)))
	mux.Handle("DELETE /api/files/{id}", app.requireAuth(http.HandlerFunc(app.remove)))
	mux.Handle("/", spaHandler(config.WebDir))

	app.handler = securityHeaders(requestLog(mux))
	app.server = &http.Server{
		Addr:              config.Address,
		Handler:           app.handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       90 * time.Second,
	}
	return app, nil
}

func (a *App) Close() error { return a.db.Close() }

func (a *App) Handler() http.Handler { return a.handler }

func (a *App) ListenAndServe() error {
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "address": a.config.Address})
}

func validateListenAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}
	if host == "" {
		return errors.New("wildcard listen address is unsafe; bind to localhost or a private LAN address")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("listen port must be between 1 and 65535")
	}
	ip := net.ParseIP(host)
	if !strings.EqualFold(host, "localhost") && (ip == nil || (!ip.IsLoopback() && !ip.IsPrivate())) {
		return errors.New("listen address must be localhost or a private LAN address")
	}
	return nil
}

func (a *App) listFiles(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("parent")
	if id == "" {
		id = rootID
	}
	dir, rel, err := a.resolve(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	files := make([]fileDTO, 0, len(entries))
	for _, entry := range entries {
		if isReservedName(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		childRel := filepath.Join(rel, entry.Name())
		kind := "file"
		if entry.IsDir() {
			kind = "folder"
		}
		files = append(files, fileDTO{
			ID:       encodeID(childRel),
			ParentID: id,
			Name:     entry.Name(),
			Kind:     kind,
			MIME:     mimeFor(entry.Name(), entry.IsDir()),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Kind != files[j].Kind {
			return files[i].Kind == "folder"
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"parent": map[string]string{"id": id, "name": filepath.Base(dir)},
		"files":  files,
	})
}

func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	select {
	case a.uploads <- struct{}{}:
		defer func() { <-a.uploads }()
	default:
		writeError(w, http.StatusTooManyRequests, errors.New("too many uploads are already running"))
		return
	}
	dir, rel, err := a.resolve(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadRequestBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if errors.Is(err, http.ErrBodyReadAfterClose) {
			writeError(w, http.StatusBadRequest, errors.New("invalid multipart upload"))
		} else {
			writeError(w, http.StatusRequestEntityTooLarge, errors.New("upload request exceeds the configured limit"))
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 || len(headers) > maxUploadFiles {
		writeError(w, http.StatusBadRequest, errors.New("multipart field 'files' is required"))
		return
	}
	for _, header := range headers {
		if header.Size < 0 || header.Size > maxUploadFileBytes {
			writeError(w, http.StatusRequestEntityTooLarge, errors.New("an uploaded file exceeds the configured limit"))
			return
		}
	}
	reserved, err := a.reserveUploadSpace(headers)
	if err != nil {
		writeError(w, http.StatusInsufficientStorage, err)
		return
	}
	defer a.releaseUploadSpace(reserved)

	created := make([]fileDTO, 0, len(headers))
	for _, header := range headers {
		name, err := safeName(header.Filename)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		source, err := header.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		targetPath := filepath.Join(dir, name)
		if _, err := os.Lstat(targetPath); err == nil {
			source.Close()
			writeError(w, http.StatusConflict, errors.New("a file with this name already exists"))
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			source.Close()
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		target, err := os.CreateTemp(a.tempDir, "upload-*")
		if err != nil {
			source.Close()
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		tempPath := target.Name()
		_, copyErr := io.Copy(target, io.LimitReader(source, maxUploadFileBytes+1))
		if copyErr == nil {
			if info, err := target.Stat(); err != nil || info.Size() > maxUploadFileBytes {
				copyErr = errors.New("an uploaded file exceeds the configured limit")
			}
		}
		syncErr := target.Sync()
		closeErr := target.Close()
		source.Close()
		if copyErr != nil || syncErr != nil || closeErr != nil {
			_ = os.Remove(tempPath)
			if copyErr != nil && strings.Contains(copyErr.Error(), "configured limit") {
				writeError(w, http.StatusRequestEntityTooLarge, copyErr)
			} else {
				writeError(w, http.StatusInternalServerError, errors.Join(copyErr, syncErr, closeErr))
			}
			return
		}
		a.uploadM.Lock()
		_, destinationErr := os.Lstat(targetPath)
		if destinationErr == nil {
			a.uploadM.Unlock()
			_ = os.Remove(tempPath)
			writeError(w, http.StatusConflict, errors.New("a file with this name already exists"))
			return
		}
		if !errors.Is(destinationErr, os.ErrNotExist) {
			a.uploadM.Unlock()
			_ = os.Remove(tempPath)
			writeError(w, http.StatusInternalServerError, destinationErr)
			return
		}
		err = os.Rename(tempPath, targetPath)
		a.uploadM.Unlock()
		if err != nil {
			_ = os.Remove(tempPath)
			if errors.Is(err, os.ErrExist) {
				writeError(w, http.StatusConflict, errors.New("a file with this name already exists"))
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}
		info, _ := os.Stat(targetPath)
		item := fileDTO{
			ID: encodeID(filepath.Join(rel, name)), ParentID: r.PathValue("id"),
			Name: name, Kind: "file", MIME: mimeFor(name, false),
			Size: info.Size(), Modified: info.ModTime(),
		}
		created = append(created, item)
		a.audit("upload", item.ID, map[string]string{"name": name})
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *App) createFolder(w http.ResponseWriter, r *http.Request) {
	dir, rel, err := a.resolve(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name, err := safeName(input.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	id := encodeID(filepath.Join(rel, name))
	a.audit("create-folder", id, map[string]string{"name": name})
	writeJSON(w, http.StatusCreated, fileDTO{
		ID: id, ParentID: r.PathValue("id"), Name: name, Kind: "folder",
		MIME: "inode/directory", Modified: time.Now(),
	})
}

func (a *App) rename(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == rootID {
		writeError(w, http.StatusBadRequest, errors.New("the shared root cannot be renamed"))
		return
	}
	current, rel, err := a.resolve(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	name, err := safeName(input.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	next := filepath.Join(filepath.Dir(current), name)
	if err := os.Rename(current, next); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	newID := encodeID(filepath.Join(filepath.Dir(rel), name))
	a.audit("rename", newID, map[string]string{"from": filepath.Base(rel), "to": name})
	writeJSON(w, http.StatusOK, map[string]string{"id": newID, "name": name})
}

func (a *App) getContent(w http.ResponseWriter, r *http.Request) {
	path, _, err := a.resolve(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		writeError(w, http.StatusBadRequest, errors.New("content is only available for files"))
		return
	}
	contentType := mimeFor(info.Name(), false)
	w.Header().Set("Content-Type", contentType)
	disposition := "inline"
	if !safeInlineContentType(contentType) {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", disposition+`; filename*=UTF-8''`+url.PathEscape(info.Name()))
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (a *App) putContent(w http.ResponseWriter, r *http.Request) {
	path, _, err := a.resolve(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		writeError(w, http.StatusBadRequest, errors.New("only existing files can be edited"))
		return
	}
	if r.ContentLength < 0 || r.ContentLength > 5<<20 {
		writeError(w, http.StatusRequestEntityTooLarge, errors.New("editable content is limited to 5 MiB"))
		return
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".wifi-share-edit-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := io.Copy(temp, io.LimitReader(r.Body, (5<<20)+1)); err != nil {
		temp.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := temp.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := os.Rename(tempName, path); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.audit("edit", r.PathValue("id"), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) remove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == rootID {
		writeError(w, http.StatusBadRequest, errors.New("the shared root cannot be deleted"))
		return
	}
	path, _, err := a.resolve(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := a.trashFn(path); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.audit("recycle", id, map[string]string{"name": filepath.Base(path)})
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) resolve(id string) (string, string, error) {
	a.rootM.RLock()
	root := a.root
	a.rootM.RUnlock()
	rel, err := decodeID(id)
	if err != nil {
		return "", "", err
	}
	candidate := filepath.Clean(filepath.Join(root, rel))
	if !within(root, candidate) {
		return "", "", errors.New("path escapes the shared directory")
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if isReservedName(part) {
			return "", "", errors.New("path is reserved for application use")
		}
	}
	if evaluated, err := filepath.EvalSymlinks(candidate); err == nil && !within(root, evaluated) {
		return "", "", errors.New("symbolic link escapes the shared directory")
	}
	return candidate, rel, nil
}

func canonicalDirectory(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func pathsOverlap(first, second string) bool {
	return within(first, second) || within(second, first)
}

func cleanupStaleUploads(root string, now time.Time) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasPrefix(entry.Name(), "upload-") {
			return err
		}
		info, err := entry.Info()
		if err == nil && now.Sub(info.ModTime()) > staleUploadAge {
			return os.Remove(path)
		}
		return err
	})
}

func (a *App) reserveUploadSpace(headers []*multipart.FileHeader) (int64, error) {
	var requested int64
	for _, header := range headers {
		requested += header.Size
	}
	used, err := directorySize(a.root)
	if err != nil {
		return 0, err
	}
	free, err := availableDiskBytes(a.root)
	if err != nil {
		return 0, err
	}
	a.quotaM.Lock()
	defer a.quotaM.Unlock()
	if requested > maxShareBytes-used-a.reservedUploadBytes {
		return 0, errors.New("the shared-folder quota would be exceeded")
	}
	if requested > free-minFreeDiskBytes-a.reservedUploadBytes {
		return 0, errors.New("not enough free disk space for this upload")
	}
	a.reservedUploadBytes += requested
	return requested, nil
}

func (a *App) releaseUploadSpace(bytes int64) {
	a.quotaM.Lock()
	a.reservedUploadBytes -= bytes
	a.quotaM.Unlock()
}

func directorySize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(_ string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

func (a *App) audit(action, id string, details any) {
	payload, _ := json.Marshal(details)
	if details == nil {
		payload = []byte("{}")
	}
	if _, err := a.db.Exec(
		`INSERT INTO operations(action, item_id, details, created_at) VALUES(?, ?, ?, ?)`,
		action, id, string(payload), time.Now().UTC().Format(time.RFC3339Nano),
	); err != nil {
		log.Printf("write audit event: %v", err)
	}
}

func encodeID(rel string) string {
	if rel == "" || rel == "." {
		return rootID
	}
	return base64.RawURLEncoding.EncodeToString([]byte(filepath.ToSlash(rel)))
}

func decodeID(id string) (string, error) {
	if id == "" || id == rootID {
		return "", nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", errors.New("invalid file id")
	}
	rel := filepath.FromSlash(string(raw))
	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid file id")
	}
	return rel, nil
}

func within(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func safeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name ||
		strings.ContainsAny(name, `/\`) || isReservedName(name) {
		return "", errors.New("invalid file name")
	}
	return name, nil
}

func mimeFor(name string, directory bool) string {
	if directory {
		return "inode/directory"
	}
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); value != "" {
		return value
	}
	return "application/octet-stream"
}

func isReservedName(name string) bool {
	return strings.HasPrefix(name, ".wifi-share-")
}

func safeInlineContentType(value string) bool {
	contentType, _, err := mime.ParseMediaType(value)
	if err != nil || contentType == "image/svg+xml" {
		return false
	}
	return strings.HasPrefix(contentType, "audio/") ||
		strings.HasPrefix(contentType, "video/") ||
		strings.HasPrefix(contentType, "image/") ||
		contentType == "application/pdf" || contentType == "text/plain"
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; object-src 'none'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

func spaHandler(webDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
			return
		}
		path := filepath.Join(webDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		index := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(index); err != nil {
			writeError(w, http.StatusServiceUnavailable, errors.New("web application is not built; run npm run build in web/"))
			return
		}
		http.ServeFile(w, r, index)
	})
}
