package workspace

import (
	"os"
)

func Create() (string, error) {
	path, err := os.MkdirTemp("", "repojet-*")

	if err != nil {
		return "", err
	}

	return path, nil
}

func Remove(path string) error {
	return os.RemoveAll(path)
}