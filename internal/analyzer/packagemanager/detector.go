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

	if facts.PackageJSON == "" {
		return nil, nil
	}

	content, err := os.ReadFile(facts.PackageJSON)
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
