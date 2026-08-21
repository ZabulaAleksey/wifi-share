package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLocalConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.local.json")
	if err := os.WriteFile(path, []byte(`{
		"address": ":9090",
		"root": "D:\\Media",
		"data": "./private",
		"web": "./ui",
		"password": "secret"
	}`), 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := loadLocalConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Address != ":9090" || config.Root != `D:\Media` || config.Password != "secret" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestLoadLocalConfigRequiresPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.local.json")
	if err := os.WriteFile(path, []byte(`{"root":"D:\\Media"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadLocalConfig(path); err == nil {
		t.Fatal("expected an empty password to be rejected")
	}
}

func TestAccessURLsContainsLocalhost(t *testing.T) {
	urls := accessURLs("127.0.0.1:9090")
	if len(urls) == 0 || urls[0] != "http://localhost:9090" {
		t.Fatalf("unexpected URLs: %#v", urls)
	}
}

func TestAccessURLsForWildcardIncludesLocalhost(t *testing.T) {
	urls := accessURLs("0.0.0.0:9090")
	if len(urls) == 0 || urls[0] != "http://localhost:9090" {
		t.Fatalf("unexpected URLs: %#v", urls)
	}
}
