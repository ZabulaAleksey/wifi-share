//go:build windows

package app

import "golang.org/x/sys/windows"

func availableDiskBytes(path string) (int64, error) {
	value, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var free uint64
	if err := windows.GetDiskFreeSpaceEx(value, &free, nil, nil); err != nil {
		return 0, err
	}
	return int64(free), nil
}
