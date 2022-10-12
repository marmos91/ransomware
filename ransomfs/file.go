package ransomfs

import (
	"os"
)

func WriteStringToFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
