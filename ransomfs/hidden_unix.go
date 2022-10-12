//go:build !windows
// +build !windows

package ransomfs

const dotCharacter = 46

func IsHidden(path string) bool {
	return path[0] == dotCharacter
}
