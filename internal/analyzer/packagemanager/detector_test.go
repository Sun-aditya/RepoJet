package packagemanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManagerFromPackageJSON(t *testing.T) {
	rootPath := t.TempDir()

	packageJSONPath := filepath.Join(rootPath, "package.json")

	content := `{
		"name": "test-project",
		"packageManager": "pnpm@10.12.1"
	}`

	if err := os.WriteFile(
		packageJSONPath,
		[]byte(content),
		0644,
	); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	result, err := Detect(packageJSONPath)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if result.Name != "pnpm" {
		t.Errorf("Name = %q, want %q", result.Name, "pnpm")
	}

	if result.Version != "10.12.1" {
		t.Errorf(
			"Version = %q, want %q",
			result.Version,
			"10.12.1",
		)
	}

	if result.Source != "package.json" {
		t.Errorf(
			"Source = %q, want %q",
			result.Source,
			"package.json",
		)
	}
}

func TestDetectHandlesMissingPackageManagerField(t *testing.T) {
	rootPath := t.TempDir()

	packageJSONPath := filepath.Join(rootPath, "package.json")

	content := `{
		"name": "test-project"
	}`

	if err := os.WriteFile(
		packageJSONPath,
		[]byte(content),
		0644,
	); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	result, err := Detect(packageJSONPath)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result != nil {
		t.Errorf("Detect() result = %#v, want nil", result)
	}
}

func TestDetectHandlesMissingPackageJSONPath(t *testing.T) {
	result, err := Detect("")

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result != nil {
		t.Errorf("Detect() result = %#v, want nil", result)
	}
}

func TestDetectRejectsInvalidPackageJSON(t *testing.T) {
	rootPath := t.TempDir()

	packageJSONPath := filepath.Join(rootPath, "package.json")

	if err := os.WriteFile(
		packageJSONPath,
		[]byte(`{"packageManager":`),
		0644,
	); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	_, err := Detect(packageJSONPath)

	if err == nil {
		t.Fatal("Detect() expected an error, got nil")
	}
}