package app

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIDsRoundTrip(t *testing.T) {
	input := filepath.Join("Music", "Artist", "track.mp3")
	id := encodeID(input)
	output, err := decodeID(id)
	if err != nil {
		t.Fatal(err)
	}
	if output != input {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestDecodeIDRejectsTraversal(t *testing.T) {
	id := encodeID(filepath.Join("..", "secret.txt"))
	if _, err := decodeID(id); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

func TestSafeName(t *testing.T) {
	for _, name := range []string{"../secret", "folder/file", `folder\file`, "", "..", ".wifi-share-tmp"} {
		if _, err := safeName(name); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
	if _, err := safeName("holiday photo.jpg"); err != nil {
		t.Fatal(err)
	}
}

func TestReservedTemporaryDirectoryIsHiddenAndInaccessible(t *testing.T) {
	instance := newTestApp(t)
	if err := os.WriteFile(filepath.Join(instance.tempDir, "upload-partial"), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	list := httptest.NewRequest(http.MethodGet, "/api/files?parent=root", nil)
	listResponse := httptest.NewRecorder()
	instance.Handler().ServeHTTP(listResponse, list)
	if strings.Contains(listResponse.Body.String(), ".wifi-share-tmp") {
		t.Fatal("temporary directory must not be listed")
	}
	content := httptest.NewRequest(http.MethodGet, "/api/files/"+encodeID(".wifi-share-tmp/upload-partial")+"/content", nil)
	contentResponse := httptest.NewRecorder()
	instance.Handler().ServeHTTP(contentResponse, content)
	if contentResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected reserved path to be denied, got %d", contentResponse.Code)
	}
}

func TestNewRejectsOverlappingShareAndDataDirectories(t *testing.T) {
	base := t.TempDir()
	_, err := New(Config{
		Address: "127.0.0.1:8080", ShareDir: base, DataDir: filepath.Join(base, "data"),
		WebDir: filepath.Join(base, "web"), Password: "password",
	})
	if err == nil || !strings.Contains(err.Error(), "must not overlap") {
		t.Fatalf("expected overlapping directories to be rejected, got %v", err)
	}
}

func TestNewRejectsWildcardListenAddress(t *testing.T) {
	base := t.TempDir()
	_, err := New(Config{
		Address: ":8080", ShareDir: filepath.Join(base, "shared"), DataDir: filepath.Join(base, "data"),
		WebDir: filepath.Join(base, "web"), Password: "password",
	})
	if err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Fatalf("expected wildcard address to be rejected, got %v", err)
	}
}

func TestNewRejectsDynamicListenPort(t *testing.T) {
	base := t.TempDir()
	_, err := New(Config{
		Address: "127.0.0.1:0", ShareDir: filepath.Join(base, "shared"), DataDir: filepath.Join(base, "data"),
		WebDir: filepath.Join(base, "web"), Password: "password",
	})
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected dynamic port to be rejected, got %v", err)
	}
}

func TestNewAllowsLocalhostListenAddress(t *testing.T) {
	base := t.TempDir()
	instance, err := New(Config{
		Address: "localhost:8080", ShareDir: filepath.Join(base, "shared"), DataDir: filepath.Join(base, "data"),
		WebDir: filepath.Join(base, "web"), Password: "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer instance.Close()
}

func TestContentForcesDownloadForActiveContent(t *testing.T) {
	instance := newTestApp(t)
	name := "unsafe.svg"
	if err := os.WriteFile(filepath.Join(instance.root, name), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/files/"+encodeID(name)+"/content", nil)
	response := httptest.NewRecorder()
	instance.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.HasPrefix(response.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatalf("active content must be served as attachment, got %q", response.Header().Get("Content-Disposition"))
	}
	if response.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("expected a content security policy")
	}
}

func TestOnlySafeContentTypesAreInline(t *testing.T) {
	if safeInlineContentType("text/html; charset=utf-8") || safeInlineContentType("application/xhtml+xml") || safeInlineContentType("image/svg+xml") {
		t.Fatal("active content types must be downloaded")
	}
	if !safeInlineContentType("audio/mpeg") || !safeInlineContentType("text/plain; charset=utf-8") {
		t.Fatal("safe media and plain text should remain inline")
	}
}

func TestUploadRejectsTooManyFiles(t *testing.T) {
	instance := newTestApp(t)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for index := 0; index < maxUploadFiles+1; index++ {
		part, err := writer.CreateFormFile("files", "file.txt")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write([]byte("data"))
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/files/root/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Origin", "http://example.com")
	request.AddCookie(loginCookie(t, instance))
	response := httptest.NewRecorder()
	instance.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestConcurrentUploadsDoNotOverwriteExistingName(t *testing.T) {
	instance := newTestApp(t)
	cookie := loginCookie(t, instance)
	makeRequest := func(contents string) *http.Request {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("files", "same.txt")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write([]byte(contents))
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/files/root/upload", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Origin", "http://example.com")
		request.AddCookie(cookie)
		return request
	}
	responses := make(chan int, 2)
	var group sync.WaitGroup
	for _, contents := range []string{"first", "second"} {
		group.Add(1)
		go func(contents string) {
			defer group.Done()
			response := httptest.NewRecorder()
			instance.Handler().ServeHTTP(response, makeRequest(contents))
			responses <- response.Code
		}(contents)
	}
	group.Wait()
	close(responses)
	created, conflict := 0, 0
	for code := range responses {
		if code == http.StatusCreated {
			created++
		}
		if code == http.StatusConflict {
			conflict++
		}
	}
	if created != 1 || conflict != 1 {
		t.Fatalf("expected one created and one conflict, got %d created and %d conflict", created, conflict)
	}
}

func TestCleanupStaleUploads(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "upload-stale")
	if err := os.WriteFile(path, []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	then := time.Now().Add(-staleUploadAge - time.Hour)
	if err := os.Chtimes(path, then, then); err != nil {
		t.Fatal(err)
	}
	if err := cleanupStaleUploads(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected stale temporary upload to be removed, got %v", err)
	}
}
