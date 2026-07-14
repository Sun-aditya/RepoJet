package packagemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"repojet/internal/scanner"
)

type Result struct {
	Name    string
	Version string
	Source  string
}

type packageJSON struct {
	PackageManager string `json:"packageManager"`
}

func Detect(facts *scanner.RepositoryFacts) (*Result, error) {
	if facts == nil {
		return nil, fmt.Errorf("repository facts cannot be nil")
	}

	if facts.PackageJSON != "" {
		result, err := detectFromPackageJSON(facts.PackageJSON)

		if err != nil {
			return nil, err
		}

		if result != nil {
			return result, nil
		}
	}

	evidence := collectLockfileEvidence(facts)

	switch len(evidence) {
	case 0:
		return nil, nil

	case 1:
		return &Result{
			Name:    evidence[0].Name,
			Version: "",
			Source:  evidence[0].Source,
		}, nil

	default:
		return nil, fmt.Errorf(
			"conflicting package manager lockfiles detected",
		)
	}
}

func detectFromPackageJSON(packageJSONPath string) (*Result, error) {
	content, err := os.ReadFile(packageJSONPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var manifest packageJSON

	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	if manifest.PackageManager == "" {
		return nil, nil
	}

	name, version, err := parsePackageManager(manifest.PackageManager)

	if err != nil {
		return nil, fmt.Errorf(
			"invalid packageManager field: %w",
			err,
		)
	}

	return &Result{
		Name:    name,
		Version: version,
		Source:  "package.json",
	}, nil
}

func parsePackageManager(value string) (string, string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", "", fmt.Errorf("package manager value is empty")
	}

	parts := strings.SplitN(value, "@", 2)

	name := parts[0]

	if name == "" {
		return "", "", fmt.Errorf("package manager name is empty")
	}

	switch name {
	case "npm", "pnpm", "yarn":
	default:
		return "", "", fmt.Errorf(
			"unsupported package manager %q",
			name,
		)
	}

	if len(parts) == 1 {
		return name, "", nil
	}

	version := parts[1]

	if version == "" {
		return "", "", fmt.Errorf(
			"package manager version is empty",
		)
	}

	return name, version, nil
}

type lockfileEvidence struct {
	Name   string
	Source string
}

func collectLockfileEvidence(
	facts *scanner.RepositoryFacts,
) []lockfileEvidence {
	var evidence []lockfileEvidence

	if facts.PnpmLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "pnpm",
			Source: "pnpm-lock.yaml",
		})
	}

	if facts.YarnLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "yarn",
			Source: "yarn.lock",
		})
	}

	if facts.PackageLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "npm",
			Source: "package-lock.json",
		})
	}

	return evidence
}
