package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetShareDirChangesRoot(t *testing.T) {
	instance := newTestApp(t)
	next := filepath.Join(t.TempDir(), "next")
	if err := os.Mkdir(next, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := instance.SetShareDir(next); err != nil {
		t.Fatal(err)
	}
	path, _, err := instance.resolve(rootID)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := filepath.Abs(next)
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}
