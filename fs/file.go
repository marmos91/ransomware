package fs

import (
	"os"
)

func WriteToFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

func WriteStringToFile(path string, content string) error {
	return WriteToFile(path, []byte(content))
}

func ReadStringFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func DeleteFileIfExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}

	return os.Remove(path)
}
