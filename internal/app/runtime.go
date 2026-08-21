package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) SetShareDir(path string) error {
	root, err := canonicalDirectory(path)
	if err != nil {
		return err
	}
	if pathsOverlap(root, a.config.DataDir) {
		return fmt.Errorf("share directory and application data directory must not overlap")
	}
	tempDir := filepath.Join(root, ".wifi-share-tmp")
	if err := os.MkdirAll(tempDir, 0o700); err != nil {
		return fmt.Errorf("create upload temporary directory: %w", err)
	}
	_ = cleanupStaleUploads(tempDir, time.Now())
	a.rootM.Lock()
	a.root = root
	a.tempDir = tempDir
	a.rootM.Unlock()
	return nil
}
