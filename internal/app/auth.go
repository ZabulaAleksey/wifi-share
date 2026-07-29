package app

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
)

const sessionCookie = "wifi_share_session"

func (a *App) authStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": a.authorized(r)})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
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
	a.sessions[token] = expires
	a.sessionM.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", Expires: expires,
		HttpOnly: true, SameSite: http.SameSiteStrictMode,
	})
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
		next.ServeHTTP(w, r)
	})
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
