package app

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const sessionCookie = "wifi_share_session"

const (
	maxLoginAttempts = 5
	maxLoginKeys     = 4096
	loginAttemptTTL  = 10 * time.Minute
)

type loginAttempts struct {
	count int
	until time.Time
	last  time.Time
}

func (a *App) authStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": a.authorized(r)})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.allowLogin(r) {
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, errors.New("too many login attempts; try again later"))
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	expected := sha256.Sum256([]byte(a.config.Password))
	provided := sha256.Sum256([]byte(input.Password))
	if subtle.ConstantTimeCompare(expected[:], provided[:]) != 1 {
		a.recordLoginFailure(r)
		a.audit("login-failed", "", nil)
		writeError(w, http.StatusUnauthorized, errors.New("invalid password"))
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	expires := time.Now().Add(24 * time.Hour)
	a.sessionM.Lock()
	a.dropExpiredSessionsLocked(time.Now())
	a.sessions[token] = expires
	a.sessionM.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", Expires: expires,
		HttpOnly: true, SameSite: http.SameSiteStrictMode,
		Secure: r.TLS != nil,
	})
	a.clearLoginFailures(r)
	a.audit("login", "", nil)
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		a.sessionM.Lock()
		delete(a.sessions, cookie.Value)
		a.sessionM.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authorized(r) {
			writeError(w, http.StatusUnauthorized, errors.New("full access requires authentication"))
			return
		}
		if !sameOriginMutation(r) {
			writeError(w, http.StatusForbidden, errors.New("cross-origin mutation is not allowed"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sameOriginMutation(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin != "" && origin != "null" && strings.EqualFold(origin, requestOrigin(r))
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func loginKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

var loginAttemptStore = struct {
	sync.Mutex
	items map[string]loginAttempts
}{items: make(map[string]loginAttempts)}

func (a *App) allowLogin(r *http.Request) bool {
	loginAttemptStore.Lock()
	defer loginAttemptStore.Unlock()
	now := time.Now()
	pruneLoginAttemptsLocked(now)
	key := loginKey(r)
	entry, exists := loginAttemptStore.items[key]
	if !exists && len(loginAttemptStore.items) >= maxLoginKeys {
		return false
	}
	return entry.until.IsZero() || time.Now().After(entry.until)
}

func (a *App) recordLoginFailure(r *http.Request) {
	loginAttemptStore.Lock()
	defer loginAttemptStore.Unlock()
	pruneLoginAttemptsLocked(time.Now())
	key := loginKey(r)
	entry := loginAttemptStore.items[key]
	entry.last = time.Now()
	entry.count++
	if entry.count >= maxLoginAttempts {
		entry.count = 0
		entry.until = time.Now().Add(time.Minute)
	}
	loginAttemptStore.items[key] = entry
}

func pruneLoginAttemptsLocked(now time.Time) {
	for key, entry := range loginAttemptStore.items {
		if now.Sub(entry.last) > loginAttemptTTL {
			delete(loginAttemptStore.items, key)
		}
	}
}

func (a *App) clearLoginFailures(r *http.Request) {
	loginAttemptStore.Lock()
	delete(loginAttemptStore.items, loginKey(r))
	loginAttemptStore.Unlock()
}

func (a *App) dropExpiredSessionsLocked(now time.Time) {
	for token, expires := range a.sessions {
		if now.After(expires) {
			delete(a.sessions, token)
		}
	}
}

func (a *App) authorized(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	a.sessionM.RLock()
	expires, ok := a.sessions[cookie.Value]
	a.sessionM.RUnlock()
	if !ok || time.Now().After(expires) {
		if ok {
			a.sessionM.Lock()
			delete(a.sessions, cookie.Value)
			a.sessionM.Unlock()
		}
		return false
	}
	return true
}
