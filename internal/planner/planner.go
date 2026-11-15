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

	// Determine install command (frozen lockfile when lockfile is present)
	installCmd := installCommand(pm, repoFacts.Lockfile)

	// Determine script commands
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

	// Carry over existing warnings
	warnings := make([]string, len(repoFacts.Warnings))
	copy(warnings, repoFacts.Warnings)

	if commands.Dev == "" {
		warnings = append(warnings, "no dev or start script found in package.json")
	}

	// Build ordered setup steps
	steps := make([]Step, 0)
	stepOrder := 1

	// Step 1: Install Node.js
	nodeVer := strings.TrimSpace(repoFacts.NodeVersion.Value)
	nodeTitle := "Install Node.js"
	nodeCmd := "nvm install"
	if nodeVer != "" {
		nodeTitle = fmt.Sprintf("Install Node.js (%s)", nodeVer)
		nodeCmd = fmt.Sprintf("nvm install %s", nodeVer)
	}
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
			Command: "curl -fsSL https://bun.sh/install | bash",
		})
		stepOrder++
	case "npm":
		// npm is bundled with Node.js, so skip separate installation step
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

func installCommand(pm, lockfile string) string {
	hasLock := lockfile != ""
	switch pm {
	case "yarn":
		if hasLock {
			return "yarn install --frozen-lockfile"
		}
		return "yarn install"
	case "pnpm":
		if hasLock {
			return "pnpm install --frozen-lockfile"
		}
		return "pnpm install"
	case "bun":
		if hasLock {
			return "bun install --frozen-lockfile"
		}
		return "bun install"
	case "npm":
		fallthrough
	default:
		if hasLock {
			return "npm ci"
		}
		return "npm install"
	}
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
