package packagemanager

import (
	"os"
	"path/filepath"
	"testing"

	"repojet/internal/scanner"
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

	facts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	result, err := Detect(facts)

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

	facts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	result, err := Detect(facts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result != nil {
		t.Errorf("Detect() result = %#v, want nil", result)
	}
}

func TestDetectHandlesMissingPackageJSONPath(t *testing.T) {
	facts := &scanner.RepositoryFacts{}

	result, err := Detect(facts)

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

	facts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	_, err := Detect(facts)

	if err == nil {
		t.Fatal("Detect() expected an error, got nil")
	}
}

func TestParsePackageManager(t *testing.T) {
	tests := []struct {
		name            string
		value           string
		expectedName    string
		expectedVersion string
		wantError       bool
	}{
		{
			name:            "pnpm with version",
			value:           "pnpm@10.12.1",
			expectedName:    "pnpm",
			expectedVersion: "10.12.1",
			wantError:       false,
		},
		{
			name:            "npm without version",
			value:           "npm",
			expectedName:    "npm",
			expectedVersion: "",
			wantError:       false,
		},
		{
			name:            "yarn with whitespace",
			value:           "  yarn@4.9.2  ",
			expectedName:    "yarn",
			expectedVersion: "4.9.2",
			wantError:       false,
		},
		{
			name:      "empty manager name",
			value:     "@10.0.0",
			wantError: true,
		},
		{
			name:      "empty version",
			value:     "pnpm@",
			wantError: true,
		},
		{
			name:      "unsupported manager",
			value:     "bun@1.2.0",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, version, err := parsePackageManager(test.value)

			gotError := err != nil

			if gotError != test.wantError {
				t.Fatalf(
					"parsePackageManager(%q) error = %v, wantError = %v",
					test.value,
					err,
					test.wantError,
				)
			}

			if test.wantError {
				return
			}

			if name != test.expectedName {
				t.Errorf(
					"name = %q, want %q",
					name,
					test.expectedName,
				)
			}

			if version != test.expectedVersion {
				t.Errorf(
					"version = %q, want %q",
					version,
					test.expectedVersion,
				)
			}
		})
	}
}

func TestDetectRejectsNilRepositoryFacts(t *testing.T) {
	_, err := Detect(nil)

	if err == nil {
		t.Fatal("Detect() expected an error, got nil")
	}
}


func TestCollectLockfileEvidence(t *testing.T) {
	tests := []struct {
		name          string
		facts         *scanner.RepositoryFacts
		expectedCount int
	}{
		{
			name:          "no lockfiles",
			facts:         &scanner.RepositoryFacts{},
			expectedCount: 0,
		},
		{
			name: "pnpm lockfile",
			facts: &scanner.RepositoryFacts{
				PnpmLock: "/repo/pnpm-lock.yaml",
			},
			expectedCount: 1,
		},
		{
			name: "multiple lockfiles",
			facts: &scanner.RepositoryFacts{
				PnpmLock:    "/repo/pnpm-lock.yaml",
				YarnLock:    "/repo/yarn.lock",
				PackageLock: "/repo/package-lock.json",
			},
			expectedCount: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			evidence := collectLockfileEvidence(test.facts)

			if len(evidence) != test.expectedCount {
				t.Errorf(
					"len(evidence) = %d, want %d",
					len(evidence),
					test.expectedCount,
				)
			}
		})
	}
}