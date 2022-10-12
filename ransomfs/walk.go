package ransomfs

import (
	"io/fs"
	"path/filepath"

	"github.com/marmos91/ransomware/utils"
)

func WalkFilesWithExtFilter(path string, extensions []string, skipHidden bool, callback func(path string, info fs.FileInfo) error) error {
	return filepath.Walk(path, func(currentPath string, currentInfo fs.FileInfo, currentErr error) error {
		if currentErr != nil {
			return currentErr
		}

		if skipHidden && IsHidden(currentPath) {
			return currentErr
		}

		if !currentInfo.IsDir() && !utils.SliceContains(extensions, filepath.Ext(currentPath)) {
			currentErr = callback(currentPath, currentInfo)
		}

		return currentErr
	})
}
