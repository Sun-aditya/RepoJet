package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestAnalyzeLocalFixtures(t *testing.T) {
	tests := []struct {
		name          string
		repoPath      string
		repoURL       string
		expectError   bool
		errorContains string
		runtime       string
		framework     string
		pm            string
		installCmd    string
		devCmd        string
	}{
		{
			name:        "Next.js + pnpm",
			repoPath:    "../testdata/nextjs-pnpm",
			repoURL:     "https://github.com/example/nextjs-pnpm",
			runtime:     "Node.js",
			framework:   "Next.js",
			pm:          "pnpm",
			installCmd:  "pnpm install --frozen-lockfile",
			devCmd:      "pnpm dev",
		},
		{
			name:        "Vite + npm",
			repoPath:    "../testdata/vite-npm",
			repoURL:     "https://github.com/example/vite-npm",
			runtime:     "Node.js",
			framework:   "Vite",
			pm:          "npm",
			installCmd:  "npm ci",
			devCmd:      "npm run dev",
		},
		{
			name:        "Express + yarn",
			repoPath:    "../testdata/express-yarn",
			repoURL:     "https://github.com/example/express-yarn",
			runtime:     "Node.js",
			framework:   "Express",
			pm:          "yarn",
			installCmd:  "yarn install --frozen-lockfile",
			devCmd:      "yarn start",
		},
		{
			name:        "Bun project",
			repoPath:    "../testdata/bun-project",
			repoURL:     "https://github.com/example/bun-project",
			runtime:     "Node.js",
			framework:   "None",
			pm:          "bun",
			installCmd:  "bun install --frozen-lockfile",
			devCmd:      "bun run dev",
		},
		{
			name:        "No dev or start scripts",
			repoPath:    "../testdata/no-dev-start",
			repoURL:     "https://github.com/example/no-dev-start",
			runtime:     "Node.js",
			framework:   "None",
			pm:          "npm",
			installCmd:  "npm install",
			devCmd:      "",
		},
		{
			name:          "No package.json (unsupported project type)",
			repoPath:      "../testdata/no-package-json",
			repoURL:       "https://github.com/example/no-package-json",
			expectError:   true,
			errorContains: "unsupported project type",
		},
		{
			name:          "Invalid package.json",
			repoPath:      "../testdata/invalid-package-json",
			repoURL:       "https://github.com/example/invalid-package-json",
			expectError:   true,
			errorContains: "failed to parse package.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AnalyzeLocal(tt.repoPath, tt.repoURL, nil)

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

			if result.Repository != tt.repoURL {
				t.Errorf("Repository = %q, want %q", result.Repository, tt.repoURL)
			}
			if result.Facts.Runtime.Value != tt.runtime {
				t.Errorf("Runtime = %q, want %q", result.Facts.Runtime.Value, tt.runtime)
			}
			if result.Facts.Framework.Value != tt.framework {
				t.Errorf("Framework = %q, want %q", result.Facts.Framework.Value, tt.framework)
			}
			if result.Facts.PackageManager.Value != tt.pm {
				t.Errorf("PackageManager = %q, want %q", result.Facts.PackageManager.Value, tt.pm)
			}
			if result.Plan.Commands.Install != tt.installCmd {
				t.Errorf("Commands.Install = %q, want %q", result.Plan.Commands.Install, tt.installCmd)
			}
			if result.Plan.Commands.Dev != tt.devCmd {
				t.Errorf("Commands.Dev = %q, want %q", result.Plan.Commands.Dev, tt.devCmd)
			}
		})
	}
}

func TestPrintReport(t *testing.T) {
	result, err := AnalyzeLocal("../testdata/nextjs-pnpm", "https://github.com/example/nextjs-pnpm", nil)
	if err != nil {
		t.Fatalf("AnalyzeLocal failed: %v", err)
	}

	var buf bytes.Buffer
	PrintReport(&buf, result)
	out := buf.String()

	expectedSections := []string{
		"Repository",
		"Project",
		"Commands",
		"Setup Plan",
		"Type:              application",
		"Runtime:           Node.js (package.json, high confidence)",
		"Framework:         Next.js (package.json (scripts: dev), high confidence)",
		"Secondary Fw:      React",
		"Package Manager:   pnpm (pnpm-lock.yaml, high confidence)",
		"Node Requirement:  >=18 (package.json (engines.node), high confidence)",
		"Install:   pnpm install --frozen-lockfile",
		"Dev:       pnpm dev",
		"1. Install Node.js (>=18)",
		"   nvm install 18",
		"2. Enable pnpm",
		"   corepack enable",
		"3. Install dependencies",
		"   pnpm install --frozen-lockfile",
		"4. Start development server",
		"   pnpm dev",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(out, expected) {
			t.Errorf("expected report to contain %q\nFull output:\n%s", expected, out)
		}
	}
}

func TestPrintJSONGolden(t *testing.T) {
	result, err := AnalyzeLocal("../testdata/nextjs-pnpm", "https://github.com/example/nextjs-pnpm", nil)
	if err != nil {
		t.Fatalf("AnalyzeLocal failed: %v", err)
	}

	var buf bytes.Buffer
	if err := PrintJSON(&buf, result); err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	goldenPath := filepath.Join("..", "testdata", "golden", "nextjs-pnpm.golden.json")

	if *update {
		if err := os.WriteFile(goldenPath, buf.Bytes(), 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		t.Log("updated golden file:", goldenPath)
	}

	goldenBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Normalize both to generic JSON maps to prevent formatting/newline differences
	var actualMap, expectedMap map[string]any
	if err := json.Unmarshal(buf.Bytes(), &actualMap); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal(goldenBytes, &expectedMap); err != nil {
		t.Fatalf("failed to unmarshal golden JSON: %v", err)
	}

	actualNormalized, _ := json.MarshalIndent(actualMap, "", "  ")
	expectedNormalized, _ := json.MarshalIndent(expectedMap, "", "  ")

	if string(actualNormalized) != string(expectedNormalized) {
		t.Errorf("JSON output does not match golden file %s.\nGot:\n%s\nWant:\n%s", goldenPath, string(actualNormalized), string(expectedNormalized))
	}
}
