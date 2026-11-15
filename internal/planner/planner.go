package planner

import (
	"fmt"
	"strings"

	"repojet/internal/facts"
)

type Step struct {
	Order   int    `json:"order"`
	Title   string `json:"title"`
	Command string `json:"command"`
}

type Commands struct {
	Install string `json:"install"`
	Dev     string `json:"dev"`
	Build   string `json:"build"`
	Start   string `json:"start"`
	Test    string `json:"test"`
}

type SetupPlan struct {
	Steps    []Step   `json:"steps"`
	Commands Commands `json:"commands"`
	Warnings []string `json:"warnings"`
}

// Generate creates an ordered setup plan and execution commands from the given facts.
// It is pure and does not perform any system execution.
func Generate(repoFacts *facts.Facts) (*SetupPlan, error) {
	if repoFacts == nil {
		return nil, fmt.Errorf("facts cannot be nil")
	}

	pm := strings.ToLower(strings.TrimSpace(repoFacts.PackageManager.Value))
	if pm == "" {
		pm = "npm"
	}

	warnings := make([]string, len(repoFacts.Warnings))
	copy(warnings, repoFacts.Warnings)

	// Determine install command
	installCmd, installWarning := resolveInstallCommand(pm, repoFacts.PackageManagerVersion, repoFacts.Lockfile)
	if installWarning != "" {
		warnings = append(warnings, installWarning)
	}

	commands := Commands{
		Install: installCmd,
	}

	// Dev command priority: dev, then start. If neither exists, leave empty and warn.
	hasDevScript := hasScript(repoFacts.Scripts, "dev")
	hasStartScript := hasScript(repoFacts.Scripts, "start")

	if hasDevScript {
		commands.Dev = scriptCommand(pm, "dev")
	} else if hasStartScript {
		commands.Dev = scriptCommand(pm, "start")
	}

	if hasScript(repoFacts.Scripts, "build") {
		commands.Build = scriptCommand(pm, "build")
	}
	if hasStartScript {
		commands.Start = scriptCommand(pm, "start")
	}
	if hasScript(repoFacts.Scripts, "test") {
		commands.Test = scriptCommand(pm, "test")
	}

	if commands.Dev == "" {
		warnings = append(warnings, "no dev or start script found in package.json")
	}

	// Build ordered setup steps
	steps := make([]Step, 0)
	stepOrder := 1

	// Step 1: Install Node.js
	nodeTitle, nodeCmd := resolveNodeInstallCommand(repoFacts.NodeVersion.Value)
	steps = append(steps, Step{
		Order:   stepOrder,
		Title:   nodeTitle,
		Command: nodeCmd,
	})
	stepOrder++

	// Step 2: Install / Enable package manager
	switch pm {
	case "pnpm", "yarn":
		steps = append(steps, Step{
			Order:   stepOrder,
			Title:   fmt.Sprintf("Enable %s", pm),
			Command: "corepack enable",
		})
		stepOrder++
	case "bun":
		steps = append(steps, Step{
			Order:   stepOrder,
			Title:   "Install bun",
			Command: "npm install -g bun",
		})
		stepOrder++
	case "npm":
		// npm is bundled with Node.js; skip step
	}

	// Step 3: Install dependencies
	steps = append(steps, Step{
		Order:   stepOrder,
		Title:   "Install dependencies",
		Command: commands.Install,
	})
	stepOrder++

	// Step 4: Start dev server
	if commands.Dev != "" {
		steps = append(steps, Step{
			Order:   stepOrder,
			Title:   "Start development server",
			Command: commands.Dev,
		})
		stepOrder++
	}

	return &SetupPlan{
		Steps:    steps,
		Commands: commands,
		Warnings: warnings,
	}, nil
}

func hasScript(scripts map[string]string, name string) bool {
	if scripts == nil {
		return false
	}
	val, ok := scripts[name]
	return ok && strings.TrimSpace(val) != ""
}

func resolveNodeInstallCommand(version string) (string, string) {
	v := strings.TrimSpace(version)
	if v == "" {
		return "Install Node.js", "nvm install --lts"
	}

	cleanV := strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	title := fmt.Sprintf("Install Node.js (%s)", v)

	lower := strings.ToLower(cleanV)
	if strings.HasPrefix(lower, "lts") || lower == "latest" || lower == "node" {
		return title, "nvm install --lts"
	}

	// If range or wildcard (>=, >, <=, <, ^, ~, x, *)
	if strings.ContainsAny(cleanV, "><=^~*xX| -") {
		major := extractMinimumMajor(cleanV)
		if major != "" {
			return title, fmt.Sprintf("nvm install %s", major)
		}
		return title, "nvm install --lts"
	}

	// Exact version or major (e.g. 20, 18.19.0)
	return title, fmt.Sprintf("nvm install %s", cleanV)
}

func extractMinimumMajor(v string) string {
	var digits []rune
	foundStart := false
	for _, r := range v {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
			foundStart = true
		} else if foundStart {
			break
		}
	}
	if len(digits) > 0 {
		return string(digits)
	}
	return ""
}

func lockfileMatchesManager(pm, lockfile string) bool {
	switch pm {
	case "npm":
		return lockfile == "package-lock.json"
	case "yarn":
		return lockfile == "yarn.lock"
	case "pnpm":
		return lockfile == "pnpm-lock.yaml"
	case "bun":
		return lockfile == "bun.lockb" || lockfile == "bun.lock"
	default:
		return false
	}
}

func resolveInstallCommand(pm, pmVersion, lockfile string) (command string, warning string) {
	hasMatchingLock := lockfile != "" && lockfileMatchesManager(pm, lockfile)

	if lockfile != "" && !hasMatchingLock {
		warning = fmt.Sprintf("lockfile %s does not match package manager %s; using plain install", lockfile, pm)
	}

	switch pm {
	case "yarn":
		if hasMatchingLock {
			if isYarnBerryOrLater(pmVersion) {
				return "yarn install --immutable", warning
			}
			return "yarn install --frozen-lockfile", warning
		}
		return "yarn install", warning
	case "pnpm":
		if hasMatchingLock {
			return "pnpm install --frozen-lockfile", warning
		}
		return "pnpm install", warning
	case "bun":
		if hasMatchingLock {
			return "bun install --frozen-lockfile", warning
		}
		return "bun install", warning
	case "npm":
		fallthrough
	default:
		if hasMatchingLock {
			return "npm ci", warning
		}
		return "npm install", warning
	}
}

func isYarnBerryOrLater(version string) bool {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	if v == "" {
		return false
	}
	parts := strings.Split(v, ".")
	major := parts[0]
	// major >= 2
	if major != "0" && major != "1" && major != "" {
		return true
	}
	return false
}

func scriptCommand(pm, scriptName string) string {
	switch pm {
	case "npm":
		if scriptName == "start" {
			return "npm start"
		}
		if scriptName == "test" {
			return "npm test"
		}
		return "npm run " + scriptName
	case "yarn":
		return "yarn " + scriptName
	case "pnpm":
		return "pnpm " + scriptName
	case "bun":
		return "bun run " + scriptName
	default:
		if scriptName == "start" {
			return "npm start"
		}
		if scriptName == "test" {
			return "npm test"
		}
		return "npm run " + scriptName
	}
}
