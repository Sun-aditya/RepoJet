package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

type RepositoryFacts struct {
	RootPath    string
	PackageJSON string
}

func Scan(rootPath string) (*RepositoryFacts, error) {
	info, err := os.Stat(rootPath)

	if err != nil {
		return nil, fmt.Errorf("failed to access repository root: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("repository root is not a directory")
	}

	facts := &RepositoryFacts{
		RootPath: rootPath,
	}

	packageJSONPath, err := findFile(rootPath, "package.json")

	if err != nil {
		return nil, err
	}

	facts.PackageJSON = packageJSONPath

	return facts, nil
}

func findFile(rootPath string, relativePath string) (string, error) {
	fullPath := filepath.Join(rootPath, relativePath)

	info, err := os.Stat(fullPath)

	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf(
			"failed to inspect file %s: %w",
			relativePath,
			err,
		)
	}

	if info.IsDir() {
		return "", nil
	}

	return fullPath, nil
}