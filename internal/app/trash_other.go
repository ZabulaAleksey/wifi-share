//go:build !windows

package app

import "fmt"

func moveToSystemTrash(string) error {
	return fmt.Errorf("system trash is currently implemented for Windows only")
}
