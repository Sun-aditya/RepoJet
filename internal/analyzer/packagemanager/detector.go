package packagemanager

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"repojet/internal/facts"
	"repojet/internal/scanner"
)

type Result struct {
	Name       string           `json:"name"`
	Version    string           `json:"version"`
	Source     string           `json:"source"`
	Confidence facts.Confidence `json:"confidence"`
	Conflicts  []string         `json:"conflicts,omitempty"`
}

type packageJSON struct {
	PackageManager string `json:"packageManager"`
}

func Detect(repoFacts *scanner.RepositoryFacts) (*Result, error) {
	if repoFacts == nil {
		return nil, fmt.Errorf("repository facts cannot be nil")
	}

	if repoFacts.PackageJSON != "" {
		result, err := detectFromPackageJSON(repoFacts.PackageJSON)
		if err != nil {
			return nil, err
		}

		if result != nil {
			return result, nil
		}
	}

	evidence := collectLockfileEvidence(repoFacts)

	switch len(evidence) {
	case 0:
		return nil, nil

	case 1:
		return &Result{
			Name:       evidence[0].Name,
			Version:    "",
			Source:     evidence[0].Source,
			Confidence: facts.ConfidenceHigh,
		}, nil

	default:
		conflicts := make([]string, len(evidence))
		for i, ev := range evidence {
			conflicts[i] = ev.Source
		}

		return &Result{
			Name:       evidence[0].Name,
			Version:    "",
			Source:     evidence[0].Source,
			Confidence: facts.ConfidenceLow,
			Conflicts:  conflicts,
		}, nil
	}
}

func detectFromPackageJSON(packageJSONPath string) (*Result, error) {
	content, err := scanner.ReadFile(packageJSONPath)
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
		Name:       name,
		Version:    version,
		Source:     "package.json",
		Confidence: facts.ConfidenceHigh,
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
	case "npm", "pnpm", "yarn", "bun":
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
	// Strip any "+hash" suffix (e.g., pnpm@8.15.0+sha512.cdd618...)
	if plusIdx := strings.Index(version, "+"); plusIdx != -1 {
		version = version[:plusIdx]
	}
	version = strings.TrimSpace(version)

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
	repoFacts *scanner.RepositoryFacts,
) []lockfileEvidence {
	var evidence []lockfileEvidence

	// Priority order: pnpm, yarn, bun, npm
	if repoFacts.PnpmLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "pnpm",
			Source: filepath.Base(repoFacts.PnpmLock),
		})
	}

	if repoFacts.YarnLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "yarn",
			Source: filepath.Base(repoFacts.YarnLock),
		})
	}

	if repoFacts.BunLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "bun",
			Source: filepath.Base(repoFacts.BunLock),
		})
	}

	if repoFacts.PackageLock != "" {
		evidence = append(evidence, lockfileEvidence{
			Name:   "npm",
			Source: filepath.Base(repoFacts.PackageLock),
		})
	}

	return evidence
}
