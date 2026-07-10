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