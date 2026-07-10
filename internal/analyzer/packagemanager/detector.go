package packagemanager

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Result struct {
	Name    string
	Version string
	Source  string
}

type packageJSON struct {
	PackageManager string `json:"packageManager"`
}

func Detect(packageJSONPath string) (*Result, error) {
	if packageJSONPath == "" {
		return nil, nil
	}

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

	name, version := parsePackageManager(manifest.PackageManager)

	return &Result{
		Name:    name,
		Version: version,
		Source:  "package.json",
	}, nil
}

func parsePackageManager(value string) (string, string) {
	parts := strings.SplitN(value, "@", 2)

	name := parts[0]

	if len(parts) == 1 {
		return name, ""
	}

	return name, parts[1]
}