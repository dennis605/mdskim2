package workspace

import (
	"os"
)

// StatFile wraps os.Stat.
func StatFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// ReadFile wraps os.ReadFile.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile wraps os.WriteFile.
func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
