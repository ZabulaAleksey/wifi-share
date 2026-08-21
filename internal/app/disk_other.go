//go:build !windows

package app

import "errors"

func availableDiskBytes(string) (int64, error) {
	return 0, errors.New("free disk space check is not implemented for this platform")
}
