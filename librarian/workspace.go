package librarian

import (
	"errors"
	"os"
)

func isDir(path string) bool {
	fi, _ := os.Stat(path)
	return fi.Mode().IsDir()
}

func pathExists(path string) (bool, error) {
	// Check that ffprobe exists - we can move this elsewhere
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	return true, nil
}
