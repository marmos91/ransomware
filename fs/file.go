package fs

import (
	"errors"
	"os"
)

func WriteToFileWithMode(path string, content []byte, mode os.FileMode) error {
	return os.WriteFile(path, content, mode)
}

func WriteStringToFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func ReadStringFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DeleteFileIfExists(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
