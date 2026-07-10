package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

type RepositoryFacts struct {
	RootPath string

	PackageJSON string

	PackageLock string
	PnpmLock    string
	YarnLock    string

	Nvmrc       string
	NodeVersion string

	Dockerfile    string
	DockerCompose string

	EnvExample string
	Readme     string

	PrismaSchema string
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

	files := []struct {
		relativePath string
		destination  *string
	}{
		{"package.json", &facts.PackageJSON},

		{"package-lock.json", &facts.PackageLock},
		{"pnpm-lock.yaml", &facts.PnpmLock},
		{"yarn.lock", &facts.YarnLock},

		{".nvmrc", &facts.Nvmrc},
		{".node-version", &facts.NodeVersion},

		{"Dockerfile", &facts.Dockerfile},

		{".env.example", &facts.EnvExample},
		{"README.md", &facts.Readme},

		{"prisma/schema.prisma", &facts.PrismaSchema},
	}

	for _, file := range files {
		path, err := findFile(rootPath, file.relativePath)

		if err != nil {
			return nil, err
		}

		*file.destination = path
	}

	dockerComposePath, err := findFirstFile(
		rootPath,
		"docker-compose.yml",
		"docker-compose.yaml",
	)

	if err != nil {
		return nil, err
	}

	facts.DockerCompose = dockerComposePath

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

func findFirstFile(rootPath string, relativePaths ...string) (string, error) {
	for _, relativePath := range relativePaths {
		path, err := findFile(rootPath, relativePath)

		if err != nil {
			return "", err
		}

		if path != "" {
			return path, nil
		}
	}

	return "", nil
}