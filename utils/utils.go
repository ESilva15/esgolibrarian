package utils

import (
	"errors"
	"os"
	"path/filepath"
)

func IsDir(path string) bool {
	fi, _ := os.Stat(path)
	return fi.Mode().IsDir()
}

func PathExists(path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	return true, nil
}
