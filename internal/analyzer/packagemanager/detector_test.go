package packagemanager

import (
	"os"
	"path/filepath"
	"testing"

	"repojet/internal/facts"
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

	repoFacts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	result, err := Detect(repoFacts)

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

	if result.Confidence != facts.ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", result.Confidence, facts.ConfidenceHigh)
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

	repoFacts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	result, err := Detect(repoFacts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result != nil {
		t.Errorf("Detect() result = %#v, want nil", result)
	}
}

func TestDetectHandlesMissingPackageJSONPath(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{}

	result, err := Detect(repoFacts)

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

	repoFacts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
	}

	_, err := Detect(repoFacts)

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
			name:            "bun with version",
			value:           "bun@1.2.0",
			expectedName:    "bun",
			expectedVersion: "1.2.0",
			wantError:       false,
		},
		{
			name:            "pnpm with hash suffix",
			value:           "pnpm@8.15.0+sha512.cdd618b1ea3c2",
			expectedName:    "pnpm",
			expectedVersion: "8.15.0",
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
			value:     "cargo@1.2.0",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name, version, err := parsePackageManager(test.value)

			if test.wantError {
				if err == nil {
					t.Fatal("parsePackageManager() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parsePackageManager() returned error: %v", err)
			}

			if name != test.expectedName {
				t.Errorf("name = %q, want %q", name, test.expectedName)
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
	t.Run("no lockfiles", func(t *testing.T) {
		repoFacts := &scanner.RepositoryFacts{}

		evidence := collectLockfileEvidence(repoFacts)

		if len(evidence) != 0 {
			t.Errorf("len(evidence) = %d, want 0", len(evidence))
		}
	})

	t.Run("pnpm lockfile", func(t *testing.T) {
		repoFacts := &scanner.RepositoryFacts{
			PnpmLock: "/repo/pnpm-lock.yaml",
		}

		evidence := collectLockfileEvidence(repoFacts)

		if len(evidence) != 1 {
			t.Fatalf("len(evidence) = %d, want 1", len(evidence))
		}

		if evidence[0].Name != "pnpm" {
			t.Errorf("evidence[0].Name = %q, want %q", evidence[0].Name, "pnpm")
		}

		if evidence[0].Source != "pnpm-lock.yaml" {
			t.Errorf(
				"evidence[0].Source = %q, want %q",
				evidence[0].Source,
				"pnpm-lock.yaml",
			)
		}
	})

	t.Run("multiple lockfiles", func(t *testing.T) {
		repoFacts := &scanner.RepositoryFacts{
			PnpmLock:    "/repo/pnpm-lock.yaml",
			PackageLock: "/repo/package-lock.json",
			BunLock:     "/repo/bun.lockb",
		}

		evidence := collectLockfileEvidence(repoFacts)

		if len(evidence) != 3 {
			t.Fatalf("len(evidence) = %d, want 3", len(evidence))
		}

		// Check priority ordering: pnpm, bun, npm (yarn was absent)
		if evidence[0].Name != "pnpm" || evidence[1].Name != "bun" || evidence[2].Name != "npm" {
			t.Errorf("unexpected evidence order: %+v", evidence)
		}
	})
}

func TestDetectPrefersPackageManagerFieldOverLockfiles(t *testing.T) {
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

	repoFacts := &scanner.RepositoryFacts{
		PackageJSON: packageJSONPath,
		PackageLock: "/repo/package-lock.json",
	}

	result, err := Detect(repoFacts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if result.Name != "pnpm" {
		t.Errorf("Name = %q, want %q", result.Name, "pnpm")
	}

	if result.Source != "package.json" {
		t.Errorf(
			"Source = %q, want %q",
			result.Source,
			"package.json",
		)
	}
}

func TestDetectPackageManagerFromPnpmLockfile(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{
		PnpmLock: "/repo/pnpm-lock.yaml",
	}

	result, err := Detect(repoFacts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if result.Name != "pnpm" {
		t.Errorf("Name = %q, want %q", result.Name, "pnpm")
	}

	if result.Source != "pnpm-lock.yaml" {
		t.Errorf(
			"Source = %q, want %q",
			result.Source,
			"pnpm-lock.yaml",
		)
	}

	if result.Confidence != facts.ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", result.Confidence, facts.ConfidenceHigh)
	}
}

func TestDetectPackageManagerFromYarnLockfile(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{
		YarnLock: "/repo/yarn.lock",
	}

	result, err := Detect(repoFacts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if result.Name != "yarn" {
		t.Errorf("Name = %q, want %q", result.Name, "yarn")
	}

	if result.Source != "yarn.lock" {
		t.Errorf(
			"Source = %q, want %q",
			result.Source,
			"yarn.lock",
		)
	}
}

func TestDetectPackageManagerFromPackageLock(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{
		PackageLock: "/repo/package-lock.json",
	}

	result, err := Detect(repoFacts)

	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("Detect() returned nil result")
	}

	if result.Name != "npm" {
		t.Errorf("Name = %q, want %q", result.Name, "npm")
	}

	if result.Source != "package-lock.json" {
		t.Errorf(
			"Source = %q, want %q",
			result.Source,
			"package-lock.json",
		)
	}
}

func TestDetectPackageManagerFromBunLockfile(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{
		BunLock: "/repo/bun.lockb",
	}

	result, err := Detect(repoFacts)
	if err != nil {
		t.Fatalf("Detect() returned an error: %v", err)
	}
	if result == nil {
		t.Fatal("Detect() returned nil result")
	}
	if result.Name != "bun" {
		t.Errorf("Name = %q, want %q", result.Name, "bun")
	}
	if result.Source != "bun.lockb" {
		t.Errorf("Source = %q, want %q", result.Source, "bun.lockb")
	}
}

func TestDetectHandlesConflictingLockfiles(t *testing.T) {
	repoFacts := &scanner.RepositoryFacts{
		PnpmLock:    "/repo/pnpm-lock.yaml",
		PackageLock: "/repo/package-lock.json",
	}

	result, err := Detect(repoFacts)
	if err != nil {
		t.Fatalf("Detect() returned unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result for conflicting lockfiles, got nil")
	}

	// Priority order: pnpm comes before npm
	if result.Name != "pnpm" {
		t.Errorf("Name = %q, want best guess %q", result.Name, "pnpm")
	}
	if result.Confidence != facts.ConfidenceLow {
		t.Errorf("Confidence = %q, want %q", result.Confidence, facts.ConfidenceLow)
	}
	if len(result.Conflicts) != 2 {
		t.Errorf("Conflicts length = %d, want 2", len(result.Conflicts))
	}
}
