package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsPackageJSON(t *testing.T) {
	rootPath := t.TempDir()

	packageJSONPath := filepath.Join(rootPath, "package.json")

	err := os.WriteFile(
		packageJSONPath,
		[]byte(`{"name":"test-project"}`),
		0644,
	)

	if err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	facts, err := Scan(rootPath)

	if err != nil {
		t.Fatalf("Scan() returned an error: %v", err)
	}

	if facts.RootPath != rootPath {
		t.Errorf(
			"RootPath = %q, want %q",
			facts.RootPath,
			rootPath,
		)
	}

	if facts.PackageJSON != packageJSONPath {
		t.Errorf(
			"PackageJSON = %q, want %q",
			facts.PackageJSON,
			packageJSONPath,
		)
	}
}

func TestScanHandlesMissingPackageJSON(t *testing.T) {
	rootPath := t.TempDir()

	facts, err := Scan(rootPath)

	if err != nil {
		t.Fatalf("Scan() returned an error: %v", err)
	}

	if facts.PackageJSON != "" {
		t.Errorf(
			"PackageJSON = %q, want empty string",
			facts.PackageJSON,
		)
	}
}

func TestScanRejectsMissingRoot(t *testing.T) {
	rootPath := filepath.Join(
		t.TempDir(),
		"repository-does-not-exist",
	)

	_, err := Scan(rootPath)

	if err == nil {
		t.Fatal("Scan() expected an error, got nil")
	}
}

func TestScanRejectsFileAsRoot(t *testing.T) {
	tempDir := t.TempDir()

	filePath := filepath.Join(tempDir, "not-a-directory")

	err := os.WriteFile(
		filePath,
		[]byte("test"),
		0644,
	)

	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err = Scan(filePath)

	if err == nil {
		t.Fatal("Scan() expected an error, got nil")
	}
}

func TestScanFindsKnownRepositoryFiles(t *testing.T) {
	rootPath := t.TempDir()

	prismaDir := filepath.Join(rootPath, "prisma")

	if err := os.Mkdir(prismaDir, 0755); err != nil {
		t.Fatalf("failed to create prisma directory: %v", err)
	}

	files := []string{
		"package.json",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		".nvmrc",
		".node-version",
		"Dockerfile",
		"docker-compose.yaml",
		".env.example",
		"README.md",
		"prisma/schema.prisma",
	}

	for _, relativePath := range files {
		fullPath := filepath.Join(rootPath, relativePath)

		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf(
				"failed to create test file %s: %v",
				relativePath,
				err,
			)
		}
	}

	facts, err := Scan(rootPath)

	if err != nil {
		t.Fatalf("Scan() returned an error: %v", err)
	}

	expected := map[string]string{
		"PackageJSON":   filepath.Join(rootPath, "package.json"),
		"PackageLock":   filepath.Join(rootPath, "package-lock.json"),
		"PnpmLock":      filepath.Join(rootPath, "pnpm-lock.yaml"),
		"YarnLock":      filepath.Join(rootPath, "yarn.lock"),
		"Nvmrc":         filepath.Join(rootPath, ".nvmrc"),
		"NodeVersion":   filepath.Join(rootPath, ".node-version"),
		"Dockerfile":    filepath.Join(rootPath, "Dockerfile"),
		"DockerCompose": filepath.Join(rootPath, "docker-compose.yaml"),
		"EnvExample":    filepath.Join(rootPath, ".env.example"),
		"Readme":        filepath.Join(rootPath, "README.md"),
		"PrismaSchema":  filepath.Join(rootPath, "prisma/schema.prisma"),
	}

	actual := map[string]string{
		"PackageJSON":   facts.PackageJSON,
		"PackageLock":   facts.PackageLock,
		"PnpmLock":      facts.PnpmLock,
		"YarnLock":      facts.YarnLock,
		"Nvmrc":         facts.Nvmrc,
		"NodeVersion":   facts.NodeVersion,
		"Dockerfile":    facts.Dockerfile,
		"DockerCompose": facts.DockerCompose,
		"EnvExample":    facts.EnvExample,
		"Readme":        facts.Readme,
		"PrismaSchema":  facts.PrismaSchema,
	}

	for name, expectedPath := range expected {
		actualPath := actual[name]

		if actualPath != expectedPath {
			t.Errorf(
				"%s = %q, want %q",
				name,
				actualPath,
				expectedPath,
			)
		}
	}
}

func TestScanFindsReadmeVariant(t *testing.T) {
	rootPath := t.TempDir()

	readmePath := filepath.Join(rootPath, "readme.md")

	if err := os.WriteFile(
		readmePath,
		[]byte("# Test Repository"),
		0644,
	); err != nil {
		t.Fatalf("failed to create readme file: %v", err)
	}

	facts, err := Scan(rootPath)

	if err != nil {
		t.Fatalf("Scan() returned an error: %v", err)
	}

	if facts.Readme != readmePath {
		t.Errorf(
			"Readme = %q, want %q",
			facts.Readme,
			readmePath,
		)
	}
}

func TestScanSkipsNodeModulesAndGit(t *testing.T) {
	rootPath := t.TempDir()

	nodeModules := filepath.Join(rootPath, "node_modules", "some-pkg")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, "package.json"), []byte(`{"name":"nested"}`), 0644); err != nil {
		t.Fatalf("failed to write nested package.json: %v", err)
	}

	gitDir := filepath.Join(rootPath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "package.json"), []byte(`{"name":"git-fake"}`), 0644); err != nil {
		t.Fatalf("failed to write git fake package.json: %v", err)
	}

	rootPackageJSON := filepath.Join(rootPath, "package.json")
	if err := os.WriteFile(rootPackageJSON, []byte(`{"name":"root"}`), 0644); err != nil {
		t.Fatalf("failed to write root package.json: %v", err)
	}

	facts, err := Scan(rootPath)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if facts.PackageJSON != rootPackageJSON {
		t.Errorf("PackageJSON = %q, want %q", facts.PackageJSON, rootPackageJSON)
	}

	for _, f := range facts.Files {
		if filepath.HasPrefix(f, "node_modules") || filepath.HasPrefix(f, ".git") {
			t.Errorf("expected to skip node_modules and .git, found: %s", f)
		}
	}
}

func TestScanSkipsSymlinks(t *testing.T) {
	rootPath := t.TempDir()

	targetFile := filepath.Join(rootPath, "actual-package.json")
	if err := os.WriteFile(targetFile, []byte(`{"name":"symlink-target"}`), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	symlinkPath := filepath.Join(rootPath, "package.json")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	facts, err := Scan(rootPath)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Symlink should be skipped, so facts.PackageJSON should not point to the symlink
	if facts.PackageJSON == symlinkPath {
		t.Errorf("expected symlink to be ignored, got PackageJSON = %q", facts.PackageJSON)
	}

	// ReadFile should also refuse to read symlinks
	_, err = ReadFile(symlinkPath)
	if err == nil {
		t.Error("ReadFile() expected error reading symlink, got nil")
	}
}

func TestScanCapFileSizeAndReadFile(t *testing.T) {
	rootPath := t.TempDir()

	largeFile := filepath.Join(rootPath, "package.json")
	// Create file slightly larger than 1 MB
	data := make([]byte, MaxFileSize+10)
	if err := os.WriteFile(largeFile, data, 0644); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	facts, err := Scan(rootPath)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Oversized file should be ignored
	if facts.PackageJSON != "" {
		t.Errorf("expected oversized package.json to be skipped, got %q", facts.PackageJSON)
	}

	// ReadFile should refuse file exceeding MaxFileSize
	_, err = ReadFile(largeFile)
	if err == nil {
		t.Error("ReadFile() expected error for oversized file, got nil")
	}
}

func TestScanFindsBunAndConfigs(t *testing.T) {
	rootPath := t.TempDir()

	files := []string{
		"package.json",
		"bun.lockb",
		"vite.config.ts",
		"next.config.mjs",
	}

	for _, f := range files {
		p := filepath.Join(rootPath, f)
		if err := os.WriteFile(p, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", f, err)
		}
	}

	facts, err := Scan(rootPath)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if facts.BunLock != filepath.Join(rootPath, "bun.lockb") {
		t.Errorf("BunLock = %q, want %q", facts.BunLock, filepath.Join(rootPath, "bun.lockb"))
	}
	if facts.ViteConfig != filepath.Join(rootPath, "vite.config.ts") {
		t.Errorf("ViteConfig = %q, want %q", facts.ViteConfig, filepath.Join(rootPath, "vite.config.ts"))
	}
	if facts.NextConfig != filepath.Join(rootPath, "next.config.mjs") {
		t.Errorf("NextConfig = %q, want %q", facts.NextConfig, filepath.Join(rootPath, "next.config.mjs"))
	}
}

func TestScanFindsBunLockTextVariant(t *testing.T) {
	rootPath := t.TempDir()

	bunLockPath := filepath.Join(rootPath, "bun.lock")
	if err := os.WriteFile(bunLockPath, []byte("# bun lockfile v1"), 0644); err != nil {
		t.Fatalf("failed to create bun.lock: %v", err)
	}

	facts, err := Scan(rootPath)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if facts.BunLock != bunLockPath {
		t.Errorf("BunLock = %q, want %q", facts.BunLock, bunLockPath)
	}
}
