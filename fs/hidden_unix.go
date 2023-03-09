//go:build !windows
// +build !windows

package fs

import (
	"path/filepath"
)

const dotCharacter = 46

func IsHidden(path string) (bool, error) {
	return filepath.Base(path)[0] == dotCharacter, nil
}
