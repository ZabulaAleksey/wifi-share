//go:build windows

package app

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	fileOperationDelete         = 3
	fileOperationAllowUndo      = 0x0040
	fileOperationNoConfirmation = 0x0010
	fileOperationSilent         = 0x0004
)

type shellFileOperation struct {
	window        uintptr
	function      uint32
	from          *uint16
	to            *uint16
	flags         uint16
	aborted       int32
	nameMappings  uintptr
	progressTitle *uint16
}

func moveToSystemTrash(path string) error {
	from, err := syscall.UTF16FromString(path + "\x00")
	if err != nil {
		return err
	}
	operation := shellFileOperation{
		function: fileOperationDelete,
		from:     &from[0],
		flags:    fileOperationAllowUndo | fileOperationNoConfirmation | fileOperationSilent,
	}
	shell32 := syscall.NewLazyDLL("shell32.dll")
	procedure := shell32.NewProc("SHFileOperationW")
	result, _, callErr := procedure.Call(uintptr(unsafe.Pointer(&operation)))
	if result != 0 {
		return fmt.Errorf("SHFileOperationW failed with code %d: %v", result, callErr)
	}
	if operation.aborted != 0 {
		return fmt.Errorf("recycle operation was aborted")
	}
	return nil
}
