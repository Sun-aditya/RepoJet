package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MaxFileSize defines the maximum allowed file size to read or process (1 MB).
const MaxFileSize = 1024 * 1024

type RepositoryFacts struct {
	RootPath string

	PackageJSON string

	PackageLock string
	PnpmLock    string
	YarnLock    string
	BunLock     string

	Nvmrc       string
	NodeVersion string

	Dockerfile    string
	DockerCompose string

	EnvExample string
	Readme     string

	PrismaSchema string

	ViteConfig string
	NextConfig string

	Files []string
}

// ReadFile safely reads a file up to MaxFileSize and refuses to read symlinks.
func ReadFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect file %s: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing to read symlink: %s", path)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", path)
	}

	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf(
			"file %s exceeds maximum allowed size of %d bytes",
			path,
			MaxFileSize,
		)
	}

	return os.ReadFile(path)
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
		Files:    make([]string, 0),
	}

	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Never follow or read symlinks.
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		// Cap file size: ignore files exceeding 1 MB.
		if fileInfo.Size() > MaxFileSize {
			return nil
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Normalize to forward slashes for cross-platform comparison
		cleanRel := filepath.ToSlash(relPath)
		baseName := d.Name()

		if isImportantFile(cleanRel, baseName) {
			facts.Files = append(facts.Files, cleanRel)
			recordFile(facts, cleanRel, baseName, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk repository: %w", err)
	}

	return facts, nil
}

func isImportantFile(relPath, baseName string) bool {
	if relPath == "package.json" {
		return true
	}

	switch baseName {
	case "package.json",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"bun.lockb",
		"bun.lock",
		".nvmrc",
		".node-version",
		".env.example",
		"Dockerfile",
		"docker-compose.yml",
		"docker-compose.yaml":
		return true
	}

	if isViteConfig(baseName) || isNextConfig(baseName) {
		return true
	}

	if relPath == "prisma/schema.prisma" {
		return true
	}

	if isReadme(baseName) {
		return true
	}

	return false
}

func isViteConfig(name string) bool {
	switch name {
	case "vite.config.js", "vite.config.ts", "vite.config.mjs", "vite.config.cjs":
		return true
	default:
		return false
	}
}

func isNextConfig(name string) bool {
	switch name {
	case "next.config.js", "next.config.mjs", "next.config.ts", "next.config.cjs":
		return true
	default:
		return false
	}
}

func isReadme(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "readme.md", "readme", "readme.txt", "readme.markdown":
		return true
	default:
		return false
	}
}

func recordFile(facts *RepositoryFacts, relPath, baseName, fullPath string) {
	switch baseName {
	case "package.json":
		if relPath == "package.json" || facts.PackageJSON == "" {
			facts.PackageJSON = fullPath
		}
	case "package-lock.json":
		if relPath == "package-lock.json" || facts.PackageLock == "" {
			facts.PackageLock = fullPath
		}
	case "pnpm-lock.yaml":
		if relPath == "pnpm-lock.yaml" || facts.PnpmLock == "" {
			facts.PnpmLock = fullPath
		}
	case "yarn.lock":
		if relPath == "yarn.lock" || facts.YarnLock == "" {
			facts.YarnLock = fullPath
		}
	case "bun.lockb", "bun.lock":
		if facts.BunLock == "" || relPath == "bun.lockb" {
			facts.BunLock = fullPath
		}
	case ".nvmrc":
		if relPath == ".nvmrc" || facts.Nvmrc == "" {
			facts.Nvmrc = fullPath
		}
	case ".node-version":
		if relPath == ".node-version" || facts.NodeVersion == "" {
			facts.NodeVersion = fullPath
		}
	case "Dockerfile":
		if relPath == "Dockerfile" || facts.Dockerfile == "" {
			facts.Dockerfile = fullPath
		}
	case "docker-compose.yml", "docker-compose.yaml":
		if relPath == "docker-compose.yml" || facts.DockerCompose == "" {
			facts.DockerCompose = fullPath
		}
	case ".env.example":
		if relPath == ".env.example" || facts.EnvExample == "" {
			facts.EnvExample = fullPath
		}
	}

	if relPath == "prisma/schema.prisma" {
		facts.PrismaSchema = fullPath
	}

	if isViteConfig(baseName) && (filepath.Dir(relPath) == "." || facts.ViteConfig == "") {
		facts.ViteConfig = fullPath
	}

	if isNextConfig(baseName) && (filepath.Dir(relPath) == "." || facts.NextConfig == "") {
		facts.NextConfig = fullPath
	}

	if isReadme(baseName) && (filepath.Dir(relPath) == "." || facts.Readme == "") {
		facts.Readme = fullPath
	}
}
