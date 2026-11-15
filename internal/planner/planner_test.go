package planner

import (
	"strings"
	"testing"

	"repojet/internal/facts"
)

func TestPlannerPackageManagersAndLockfiles(t *testing.T) {
	tests := []struct {
		name            string
		pm              string
		lockfile        string
		scripts         map[string]string
		nodeVersion     string
		expectedInstall string
		expectedDev     string
		expectedBuild   string
		expectedStart   string
		expectedTest    string
		hasStep2        bool
		step2Command    string
	}{
		{
			name:            "npm with lockfile",
			pm:              "npm",
			lockfile:        "package-lock.json",
			scripts:         map[string]string{"dev": "next dev", "build": "next build", "start": "next start", "test": "jest"},
			nodeVersion:     "20",
			expectedInstall: "npm ci",
			expectedDev:     "npm run dev",
			expectedBuild:   "npm run build",
			expectedStart:   "npm start",
			expectedTest:    "npm test",
			hasStep2:        false, // npm skips step 2
		},
		{
			name:            "npm without lockfile",
			pm:              "npm",
			lockfile:        "",
			scripts:         map[string]string{"dev": "vite"},
			expectedInstall: "npm install",
			expectedDev:     "npm run dev",
			hasStep2:        false,
		},
		{
			name:            "yarn with lockfile",
			pm:              "yarn",
			lockfile:        "yarn.lock",
			scripts:         map[string]string{"dev": "vite", "build": "vite build", "start": "node index.js", "test": "mocha"},
			expectedInstall: "yarn install --frozen-lockfile",
			expectedDev:     "yarn dev",
			expectedBuild:   "yarn build",
			expectedStart:   "yarn start",
			expectedTest:    "yarn test",
			hasStep2:        true,
			step2Command:    "corepack enable",
		},
		{
			name:            "yarn without lockfile",
			pm:              "yarn",
			lockfile:        "",
			scripts:         map[string]string{"dev": "vite"},
			expectedInstall: "yarn install",
			expectedDev:     "yarn dev",
			hasStep2:        true,
			step2Command:    "corepack enable",
		},
		{
			name:            "pnpm with lockfile",
			pm:              "pnpm",
			lockfile:        "pnpm-lock.yaml",
			scripts:         map[string]string{"dev": "next dev", "build": "next build", "start": "next start", "test": "vitest"},
			nodeVersion:     ">=18.0.0",
			expectedInstall: "pnpm install --frozen-lockfile",
			expectedDev:     "pnpm dev",
			expectedBuild:   "pnpm build",
			expectedStart:   "pnpm start",
			expectedTest:    "pnpm test",
			hasStep2:        true,
			step2Command:    "corepack enable",
		},
		{
			name:            "pnpm without lockfile",
			pm:              "pnpm",
			lockfile:        "",
			scripts:         map[string]string{"dev": "next dev"},
			expectedInstall: "pnpm install",
			expectedDev:     "pnpm dev",
			hasStep2:        true,
			step2Command:    "corepack enable",
		},
		{
			name:            "bun with lockfile",
			pm:              "bun",
			lockfile:        "bun.lockb",
			scripts:         map[string]string{"dev": "bun run dev.ts", "build": "bun build", "start": "bun run server.ts", "test": "bun test"},
			expectedInstall: "bun install --frozen-lockfile",
			expectedDev:     "bun run dev",
			expectedBuild:   "bun run build",
			expectedStart:   "bun run start",
			expectedTest:    "bun run test",
			hasStep2:        true,
			step2Command:    "curl -fsSL https://bun.sh/install | bash",
		},
		{
			name:            "bun without lockfile",
			pm:              "bun",
			lockfile:        "",
			scripts:         map[string]string{"dev": "bun run dev.ts"},
			expectedInstall: "bun install",
			expectedDev:     "bun run dev",
			hasStep2:        true,
			step2Command:    "curl -fsSL https://bun.sh/install | bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := facts.New()
			f.PackageManager = facts.Fact{Value: tt.pm}
			f.Lockfile = tt.lockfile
			f.NodeVersion = facts.Fact{Value: tt.nodeVersion}
			f.Scripts = tt.scripts

			plan, err := Generate(f)
			if err != nil {
				t.Fatalf("Generate() returned error: %v", err)
			}

			if plan.Commands.Install != tt.expectedInstall {
				t.Errorf("Commands.Install = %q, want %q", plan.Commands.Install, tt.expectedInstall)
			}
			if tt.expectedDev != "" && plan.Commands.Dev != tt.expectedDev {
				t.Errorf("Commands.Dev = %q, want %q", plan.Commands.Dev, tt.expectedDev)
			}
			if tt.expectedBuild != "" && plan.Commands.Build != tt.expectedBuild {
				t.Errorf("Commands.Build = %q, want %q", plan.Commands.Build, tt.expectedBuild)
			}
			if tt.expectedStart != "" && plan.Commands.Start != tt.expectedStart {
				t.Errorf("Commands.Start = %q, want %q", plan.Commands.Start, tt.expectedStart)
			}
			if tt.expectedTest != "" && plan.Commands.Test != tt.expectedTest {
				t.Errorf("Commands.Test = %q, want %q", plan.Commands.Test, tt.expectedTest)
			}

			// Verify Step ordering
			for i, step := range plan.Steps {
				if step.Order != i+1 {
					t.Errorf("Step[%d].Order = %d, want %d", i, step.Order, i+1)
				}
			}

			// Verify Step 1 (Node)
			if len(plan.Steps) == 0 {
				t.Fatal("expected at least 1 step")
			}
			if tt.nodeVersion != "" {
				if !strings.Contains(plan.Steps[0].Title, tt.nodeVersion) {
					t.Errorf("Step 1 Title %q does not contain nodeVersion %q", plan.Steps[0].Title, tt.nodeVersion)
				}
				if !strings.Contains(plan.Steps[0].Command, tt.nodeVersion) {
					t.Errorf("Step 1 Command %q does not contain nodeVersion %q", plan.Steps[0].Command, tt.nodeVersion)
				}
			}

			// Verify Step 2 (Package manager step)
			if tt.hasStep2 {
				if len(plan.Steps) < 3 {
					t.Fatalf("expected at least 3 steps, got %d", len(plan.Steps))
				}
				if plan.Steps[1].Command != tt.step2Command {
					t.Errorf("Step 2 command = %q, want %q", plan.Steps[1].Command, tt.step2Command)
				}
			} else {
				// npm should go straight from install node to install dependencies
				if plan.Steps[1].Title != "Install dependencies" {
					t.Errorf("npm step 2 title = %q, want Install dependencies", plan.Steps[1].Title)
				}
			}
		})
	}
}

func TestPlannerDevCommandPriorityAndMissing(t *testing.T) {
	t.Run("dev script prioritized over start script", func(t *testing.T) {
		f := facts.New()
		f.PackageManager = facts.Fact{Value: "npm"}
		f.Scripts = map[string]string{
			"dev":   "next dev",
			"start": "next start",
		}

		plan, err := Generate(f)
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}

		if plan.Commands.Dev != "npm run dev" {
			t.Errorf("Commands.Dev = %q, want %q", plan.Commands.Dev, "npm run dev")
		}
	})

	t.Run("start script used when dev script is absent", func(t *testing.T) {
		f := facts.New()
		f.PackageManager = facts.Fact{Value: "npm"}
		f.Scripts = map[string]string{
			"start": "node server.js",
		}

		plan, err := Generate(f)
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}

		if plan.Commands.Dev != "npm start" {
			t.Errorf("Commands.Dev = %q, want %q", plan.Commands.Dev, "npm start")
		}
	})

	t.Run("neither dev nor start exists - empty dev and warning", func(t *testing.T) {
		f := facts.New()
		f.PackageManager = facts.Fact{Value: "pnpm"}
		f.Scripts = map[string]string{
			"build": "pnpm build",
			"test":  "vitest",
		}

		plan, err := Generate(f)
		if err != nil {
			t.Fatalf("Generate() returned error: %v", err)
		}

		if plan.Commands.Dev != "" {
			t.Errorf("Commands.Dev = %q, want empty string", plan.Commands.Dev)
		}

		foundWarning := false
		for _, w := range plan.Warnings {
			if strings.Contains(w, "no dev or start script") {
				foundWarning = true
				break
			}
		}
		if !foundWarning {
			t.Errorf("expected warning for missing dev/start scripts, got: %v", plan.Warnings)
		}

		// Verify Start development server step is not in steps
		for _, s := range plan.Steps {
			if s.Title == "Start development server" {
				t.Error("unexpected Start development server step when no dev/start script exists")
			}
		}
	})
}

func TestPlannerCarriesOverWarnings(t *testing.T) {
	f := facts.New()
	f.AddWarning("first warning from analyzer")
	f.AddWarning("second warning from analyzer")

	plan, err := Generate(f)
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if len(plan.Warnings) < 2 {
		t.Fatalf("expected at least 2 warnings, got %d", len(plan.Warnings))
	}
	if plan.Warnings[0] != "first warning from analyzer" || plan.Warnings[1] != "second warning from analyzer" {
		t.Errorf("unexpected warnings: %v", plan.Warnings)
	}
}

func TestPlannerRejectsNilFacts(t *testing.T) {
	_, err := Generate(nil)
	if err == nil {
		t.Fatal("expected error for nil facts, got nil")
	}
}
