package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"repojet/internal/facts"
)

func TestNodeAnalyzerDetect(t *testing.T) {
	analyzer := NewNodeAnalyzer()

	if analyzer.Name() != "Node.js" {
		t.Errorf("Name() = %q, want %q", analyzer.Name(), "Node.js")
	}

	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{
			name:     "valid repo with package.json",
			dir:      "../../testdata/nextjs-pnpm",
			expected: true,
		},
		{
			name:     "repo without package.json",
			dir:      "../../testdata/no-package-json",
			expected: false,
		},
		{
			name:     "non-existent dir",
			dir:      "../../testdata/does-not-exist",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.Detect(tt.dir)
			if got != tt.expected {
				t.Errorf("Detect(%q) = %v, want %v", tt.dir, got, tt.expected)
			}
		})
	}
}

func TestNodeAnalyzerFixtures(t *testing.T) {
	analyzer := NewNodeAnalyzer()

	tests := []struct {
		name                   string
		fixtureDir             string
		expectError            bool
		errorContains          string
		expectedRuntime        string
		expectedFramework      string
		expectedPackageManager string
		pmConfidence           facts.Confidence
		expectedNodeVersion    string
		expectedScripts        map[string]string
		expectedEnvFiles       []string
		warningContains        string
	}{
		{
			name:                   "Next.js with pnpm",
			fixtureDir:             "../../testdata/nextjs-pnpm",
			expectedRuntime:        "Node.js",
			expectedFramework:      "Next.js",
			expectedPackageManager: "pnpm",
			pmConfidence:           facts.ConfidenceHigh,
			expectedNodeVersion:    ">=18",
			expectedScripts: map[string]string{
				"dev":   "next dev",
				"build": "next build",
				"start": "next start",
				"test":  "jest",
			},
			expectedEnvFiles: []string{".env.example"},
		},
		{
			name:                   "Vite with npm",
			fixtureDir:             "../../testdata/vite-npm",
			expectedRuntime:        "Node.js",
			expectedFramework:      "Vite",
			expectedPackageManager: "npm",
			pmConfidence:           facts.ConfidenceHigh,
			expectedNodeVersion:    "20",
			expectedScripts: map[string]string{
				"dev":   "vite",
				"build": "vite build",
			},
		},
		{
			name:                   "Express with yarn",
			fixtureDir:             "../../testdata/express-yarn",
			expectedRuntime:        "Node.js",
			expectedFramework:      "Express",
			expectedPackageManager: "yarn",
			pmConfidence:           facts.ConfidenceHigh,
			expectedNodeVersion:    "18.19.0",
			expectedScripts: map[string]string{
				"start": "node index.js",
				"test":  "mocha",
			},
		},
		{
			name:                   "Bun project with packageManager field",
			fixtureDir:             "../../testdata/bun-project",
			expectedRuntime:        "Node.js",
			expectedFramework:      "None",
			expectedPackageManager: "bun",
			pmConfidence:           facts.ConfidenceHigh,
			expectedScripts: map[string]string{
				"dev":   "bun run dev.ts",
				"build": "bun build",
			},
		},
		{
			name:                   "Multiple conflicting lockfiles",
			fixtureDir:             "../../testdata/multiple-lockfiles",
			expectedRuntime:        "Node.js",
			expectedFramework:      "Express",
			expectedPackageManager: "pnpm", // best guess by priority
			pmConfidence:           facts.ConfidenceLow,
			expectedScripts: map[string]string{
				"start": "node server.js",
			},
			warningContains: "multiple lockfiles detected",
		},
		{
			name:          "No package.json",
			fixtureDir:    "../../testdata/no-package-json",
			expectError:   true,
			errorContains: "no package.json found",
		},
		{
			name:          "Invalid package.json",
			fixtureDir:    "../../testdata/invalid-package-json",
			expectError:   true,
			errorContains: "failed to parse package.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := analyzer.Analyze(tt.fixtureDir)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errorContains)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if res.Runtime.Value != tt.expectedRuntime {
				t.Errorf("Runtime = %q, want %q", res.Runtime.Value, tt.expectedRuntime)
			}
			if res.Framework.Value != tt.expectedFramework {
				t.Errorf("Framework = %q, want %q", res.Framework.Value, tt.expectedFramework)
			}
			if res.PackageManager.Value != tt.expectedPackageManager {
				t.Errorf("PackageManager = %q, want %q", res.PackageManager.Value, tt.expectedPackageManager)
			}
			if res.PackageManager.Confidence != tt.pmConfidence {
				t.Errorf("PackageManager.Confidence = %q, want %q", res.PackageManager.Confidence, tt.pmConfidence)
			}

			if tt.expectedNodeVersion != "" && res.NodeVersion.Value != tt.expectedNodeVersion {
				t.Errorf("NodeVersion = %q, want %q", res.NodeVersion.Value, tt.expectedNodeVersion)
			}

			for scriptName, expectedCmd := range tt.expectedScripts {
				if cmd, ok := res.Scripts[scriptName]; !ok || cmd != expectedCmd {
					t.Errorf("Scripts[%q] = %q, want %q", scriptName, cmd, expectedCmd)
				}
			}

			for _, expectedEnv := range tt.expectedEnvFiles {
				found := false
				for _, env := range res.EnvFiles {
					if env == expectedEnv {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected env file %q in EnvFiles: %v", expectedEnv, res.EnvFiles)
				}
			}

			if tt.warningContains != "" {
				foundWarning := false
				for _, w := range res.Warnings {
					if strings.Contains(w, tt.warningContains) {
						foundWarning = true
						break
					}
				}
				if !foundWarning {
					t.Errorf("expected warning containing %q, got: %v", tt.warningContains, res.Warnings)
				}
			}
		})
	}
}

func TestNodeAnalyzerDisagreeingNodeVersions(t *testing.T) {
	tempDir := t.TempDir()

	pkgJSON := `{
		"name": "node-version-test",
		"engines": {
			"node": ">=20.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("18.17.0\n"), 0644); err != nil {
		t.Fatalf("failed to write .nvmrc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".node-version"), []byte("16.20.0\n"), 0644); err != nil {
		t.Fatalf("failed to write .node-version: %v", err)
	}

	analyzer := NewNodeAnalyzer()
	res, err := analyzer.Analyze(tempDir)
	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	// Should prioritize engines.node
	if res.NodeVersion.Value != ">=20.0.0" {
		t.Errorf("NodeVersion.Value = %q, want %q", res.NodeVersion.Value, ">=20.0.0")
	}
	if res.NodeVersion.Confidence != facts.ConfidenceHigh {
		t.Errorf("NodeVersion.Confidence = %q, want %q", res.NodeVersion.Confidence, facts.ConfidenceHigh)
	}

	// Should contain warnings about disagreement
	foundNvmrcWarning := false
	foundNodeVersionWarning := false
	for _, w := range res.Warnings {
		if strings.Contains(w, ".nvmrc") {
			foundNvmrcWarning = true
		}
		if strings.Contains(w, ".node-version") {
			foundNodeVersionWarning = true
		}
	}

	if !foundNvmrcWarning {
		t.Errorf("expected warning mentioning .nvmrc, got: %v", res.Warnings)
	}
	if !foundNodeVersionWarning {
		t.Errorf("expected warning mentioning .node-version, got: %v", res.Warnings)
	}
}

func TestNodeAnalyzerWorkspacesWarning(t *testing.T) {
	tempDir := t.TempDir()

	pkgJSON := `{
		"name": "monorepo-test",
		"workspaces": ["packages/*"]
	}`
	if err := os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	analyzer := NewNodeAnalyzer()
	res, err := analyzer.Analyze(tempDir)
	if err != nil {
		t.Fatalf("Analyze() returned error: %v", err)
	}

	foundWarning := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "workspaces detected") {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Errorf("expected workspaces warning, got warnings: %v", res.Warnings)
	}
}
