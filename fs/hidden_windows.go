//go:build windows

package fs

import (
	"path/filepath"
	"syscall"
)

func IsHidden(path string) (bool, error) {
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return false, nil
	}
	if base[0] == '.' {
		return true, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	// Prefix with `\\?\` to support long paths on Windows.
	// See: https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + absPath)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
