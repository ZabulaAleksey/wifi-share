package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) SetShareDir(path string) error {
	root, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("open shared directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("shared path is not a directory")
	}
	a.rootM.Lock()
	a.root = root
	a.rootM.Unlock()
	return nil
}
