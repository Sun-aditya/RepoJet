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