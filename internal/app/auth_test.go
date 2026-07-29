package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	base := t.TempDir()
	instance, err := New(Config{
		Address:  "127.0.0.1:0",
		ShareDir: filepath.Join(base, "shared"),
		DataDir:  filepath.Join(base, "data"),
		WebDir:   filepath.Join(base, "web"),
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = instance.Close() })
	return instance
}

func loginCookie(t *testing.T, instance *App) *http.Cookie {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"password":"correct horse battery staple"}`))
	response := httptest.NewRecorder()
	instance.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("login returned %d: %s", response.Code, response.Body.String())
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == sessionCookie {
			return cookie
		}
	}
	t.Fatal("login did not set a session cookie")
	return nil
}

func TestMutationRequiresAuthentication(t *testing.T) {
	instance := newTestApp(t)
	request := httptest.NewRequest(http.MethodPost, "/api/files/root/folders",
		bytes.NewBufferString(`{"name":"Private"}`))
	response := httptest.NewRecorder()

	instance.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	instance := newTestApp(t)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		bytes.NewBufferString(`{"password":"wrong"}`))
	response := httptest.NewRecorder()

	instance.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestAuthenticatedUserCanCreateFolder(t *testing.T) {
	instance := newTestApp(t)
	request := httptest.NewRequest(http.MethodPost, "/api/files/root/folders",
		bytes.NewBufferString(`{"name":"Private"}`))
	request.AddCookie(loginCookie(t, instance))
	response := httptest.NewRecorder()

	instance.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	if _, err := os.Stat(filepath.Join(instance.root, "Private")); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteUsesSystemTrashAdapter(t *testing.T) {
	instance := newTestApp(t)
	path := filepath.Join(instance.root, "remove-me.txt")
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	var recycled string
	instance.trashFn = func(path string) error {
		recycled = path
		return os.Remove(path)
	}
	request := httptest.NewRequest(http.MethodDelete,
		"/api/files/"+encodeID("remove-me.txt"), nil)
	request.AddCookie(loginCookie(t, instance))
	response := httptest.NewRecorder()

	instance.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", response.Code, response.Body.String())
	}
	if recycled != path {
		t.Fatalf("expected %q to be recycled, got %q", path, recycled)
	}
}
