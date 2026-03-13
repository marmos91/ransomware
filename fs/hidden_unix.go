//go:build !windows

package fs

import (
	"path/filepath"
)

func IsHidden(path string) (bool, error) {
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return false, nil
	}
	return base[0] == '.', nil
}
