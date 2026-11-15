package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"repojet/internal/analyzer/packagemanager"
	"repojet/internal/facts"
	"repojet/internal/scanner"
)

type NodeAnalyzer struct{}

func NewNodeAnalyzer() *NodeAnalyzer {
	return &NodeAnalyzer{}
}

func (a *NodeAnalyzer) Name() string {
	return "Node.js"
}

func (a *NodeAnalyzer) Detect(repoPath string) bool {
	packageJSONPath := filepath.Join(repoPath, "package.json")
	info, err := os.Stat(packageJSONPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

type packageJSONManifest struct {
	Name            string            `json:"name"`
	Scripts         map[string]string `json:"scripts"`
	Engines         struct {
		Node string `json:"node"`
	} `json:"engines"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Workspaces      any               `json:"workspaces"`
}

func (a *NodeAnalyzer) Analyze(repoPath string) (*facts.Facts, error) {
	repoFacts, err := scanner.Scan(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	if repoFacts.PackageJSON == "" {
		return nil, fmt.Errorf("no package.json found in repository")
	}

	content, err := scanner.ReadFile(repoFacts.PackageJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var manifest packageJSONManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	result := facts.New()

	// Runtime
	result.Runtime = facts.Fact{
		Value:      "Node.js",
		Evidence:   "package.json",
		Confidence: facts.ConfidenceHigh,
	}

	// Workspaces check (monorepos not supported yet)
	if manifest.Workspaces != nil {
		result.AddWarning("workspaces detected in package.json (monorepo not supported yet)")
	}

	// Package Manager & Lockfile Detection
	a.detectPackageManagerAndLockfile(manifest, repoFacts, result)

	// Node Version Detection
	a.detectNodeVersion(manifest, repoFacts, result)

	// Framework Detection (Priority: Next.js -> Vite -> Express -> React)
	a.detectFramework(manifest, repoFacts, result)

	// Scripts Extraction (extract dev, start, build, test only)
	if manifest.Scripts != nil {
		allowedScripts := []string{"dev", "start", "build", "test"}
		for _, name := range allowedScripts {
			if val, ok := manifest.Scripts[name]; ok {
				trimmed := strings.TrimSpace(val)
				if trimmed != "" {
					result.Scripts[name] = trimmed
				}
			}
		}
	}

	// Env Files
	if repoFacts.EnvExample != "" {
		result.EnvFiles = append(result.EnvFiles, filepath.Base(repoFacts.EnvExample))
	}

	return result, nil
}

func (a *NodeAnalyzer) detectPackageManagerAndLockfile(
	manifest packageJSONManifest,
	repoFacts *scanner.RepositoryFacts,
	result *facts.Facts,
) {
	var presentLockfiles []string
	if repoFacts.PnpmLock != "" {
		presentLockfiles = append(presentLockfiles, filepath.Base(repoFacts.PnpmLock))
	}
	if repoFacts.YarnLock != "" {
		presentLockfiles = append(presentLockfiles, filepath.Base(repoFacts.YarnLock))
	}
	if repoFacts.BunLock != "" {
		presentLockfiles = append(presentLockfiles, filepath.Base(repoFacts.BunLock))
	}
	if repoFacts.PackageLock != "" {
		presentLockfiles = append(presentLockfiles, filepath.Base(repoFacts.PackageLock))
	}

	pmResult, err := packagemanager.Detect(repoFacts)
	if err != nil {
		result.PackageManager = facts.Fact{
			Value:      "npm",
			Evidence:   "default fallback",
			Confidence: facts.ConfidenceLow,
		}
		result.AddWarning(fmt.Sprintf("package manager detection error: %v; defaulting to npm", err))
		return
	}

	if pmResult != nil {
		result.PackageManager = facts.Fact{
			Value:      pmResult.Name,
			Evidence:   pmResult.Source,
			Confidence: pmResult.Confidence,
		}

		if pmResult.Source == "package.json" {
			// packageManager field detected
			if len(presentLockfiles) > 0 {
				matchingLockfile := findMatchingLockfile(pmResult.Name, presentLockfiles)
				if matchingLockfile != "" {
					result.Lockfile = matchingLockfile
				} else {
					result.Lockfile = presentLockfiles[0]
					result.AddWarning(fmt.Sprintf(
						"packageManager field specifies %s but %s is present",
						pmResult.Name,
						strings.Join(presentLockfiles, ", "),
					))
				}
			}
		} else {
			// Detected from lockfile
			result.Lockfile = pmResult.Source
			if len(pmResult.Conflicts) > 1 {
				result.AddWarning(fmt.Sprintf(
					"multiple lockfiles detected (%s); chose %s",
					strings.Join(pmResult.Conflicts, ", "),
					pmResult.Name,
				))
			}
		}
	} else {
		result.PackageManager = facts.Fact{
			Value:      "npm",
			Evidence:   "default fallback",
			Confidence: facts.ConfidenceLow,
		}
		result.AddWarning("no package manager specified; defaulting to npm")
	}
}

func findMatchingLockfile(manager string, lockfiles []string) string {
	for _, lf := range lockfiles {
		switch manager {
		case "npm":
			if lf == "package-lock.json" {
				return lf
			}
		case "yarn":
			if lf == "yarn.lock" {
				return lf
			}
		case "pnpm":
			if lf == "pnpm-lock.yaml" {
				return lf
			}
		case "bun":
			if lf == "bun.lockb" || lf == "bun.lock" {
				return lf
			}
		}
	}
	return ""
}

func (a *NodeAnalyzer) detectNodeVersion(
	manifest packageJSONManifest,
	repoFacts *scanner.RepositoryFacts,
	result *facts.Facts,
) {
	enginesNode := strings.TrimSpace(manifest.Engines.Node)

	var nvmrcVersion string
	if repoFacts.Nvmrc != "" {
		if content, err := scanner.ReadFile(repoFacts.Nvmrc); err == nil {
			nvmrcVersion = strings.TrimSpace(string(content))
		}
	}

	var nodeVersionFile string
	if repoFacts.NodeVersion != "" {
		if content, err := scanner.ReadFile(repoFacts.NodeVersion); err == nil {
			nodeVersionFile = strings.TrimSpace(string(content))
		}
	}

	if enginesNode != "" {
		result.NodeVersion = facts.Fact{
			Value:      enginesNode,
			Evidence:   "package.json (engines.node)",
			Confidence: facts.ConfidenceHigh,
		}

		if isNodeVersionIncompatible(enginesNode, nvmrcVersion) {
			result.AddWarning(fmt.Sprintf(
				"conflicting Node versions: engines.node (%s) disagrees with .nvmrc (%s); using engines.node",
				enginesNode,
				nvmrcVersion,
			))
		}
		if isNodeVersionIncompatible(enginesNode, nodeVersionFile) {
			result.AddWarning(fmt.Sprintf(
				"conflicting Node versions: engines.node (%s) disagrees with .node-version (%s); using engines.node",
				enginesNode,
				nodeVersionFile,
			))
		}
	} else if nvmrcVersion != "" {
		result.NodeVersion = facts.Fact{
			Value:      nvmrcVersion,
			Evidence:   ".nvmrc",
			Confidence: facts.ConfidenceHigh,
		}

		if isNodeVersionIncompatible(nvmrcVersion, nodeVersionFile) {
			result.AddWarning(fmt.Sprintf(
				"conflicting Node versions: .nvmrc (%s) disagrees with .node-version (%s); using .nvmrc",
				nvmrcVersion,
				nodeVersionFile,
			))
		}
	} else if nodeVersionFile != "" {
		result.NodeVersion = facts.Fact{
			Value:      nodeVersionFile,
			Evidence:   ".node-version",
			Confidence: facts.ConfidenceHigh,
		}
	} else {
		result.NodeVersion = facts.Fact{
			Value:      "",
			Evidence:   "none",
			Confidence: facts.ConfidenceLow,
		}
	}
}

func isNodeVersionIncompatible(v1, v2 string) bool {
	v1 = strings.TrimSpace(v1)
	v2 = strings.TrimSpace(v2)
	v1 = strings.TrimPrefix(strings.TrimPrefix(v1, "v"), "V")
	v2 = strings.TrimPrefix(strings.TrimPrefix(v2, "v"), "V")
	if v1 == "" || v2 == "" || v1 == v2 {
		return false
	}

	// Skip comparison for ranges, aliases, or wildcard specifications
	if strings.HasPrefix(strings.ToLower(v1), "lts") || strings.HasPrefix(strings.ToLower(v2), "lts") {
		return false
	}
	if strings.ContainsAny(v1, "><=^~*| -") || strings.ContainsAny(v2, "><=^~*| -") {
		return false
	}

	// Both are simple version numbers: compare major version
	major1 := strings.Split(v1, ".")[0]
	major2 := strings.Split(v2, ".")[0]
	return major1 != major2
}

func (a *NodeAnalyzer) detectFramework(
	manifest packageJSONManifest,
	repoFacts *scanner.RepositoryFacts,
	result *facts.Facts,
) {
	depEvidence := func(name string) string {
		if manifest.Dependencies != nil {
			if _, ok := manifest.Dependencies[name]; ok {
				return fmt.Sprintf("package.json (dependencies: %s)", name)
			}
		}
		if manifest.DevDependencies != nil {
			if _, ok := manifest.DevDependencies[name]; ok {
				return fmt.Sprintf("package.json (devDependencies: %s)", name)
			}
		}
		return ""
	}

	// 1. Next.js
	if ev := depEvidence("next"); ev != "" {
		result.Framework = facts.Fact{
			Value:      "Next.js",
			Evidence:   ev,
			Confidence: facts.ConfidenceHigh,
		}
		return
	}
	if repoFacts.NextConfig != "" {
		result.Framework = facts.Fact{
			Value:      "Next.js",
			Evidence:   filepath.Base(repoFacts.NextConfig),
			Confidence: facts.ConfidenceHigh,
		}
		return
	}

	// 2. Vite
	if ev := depEvidence("vite"); ev != "" {
		result.Framework = facts.Fact{
			Value:      "Vite",
			Evidence:   ev,
			Confidence: facts.ConfidenceHigh,
		}
		return
	}
	if repoFacts.ViteConfig != "" {
		result.Framework = facts.Fact{
			Value:      "Vite",
			Evidence:   filepath.Base(repoFacts.ViteConfig),
			Confidence: facts.ConfidenceHigh,
		}
		return
	}

	// 3. Express
	if ev := depEvidence("express"); ev != "" {
		result.Framework = facts.Fact{
			Value:      "Express",
			Evidence:   ev,
			Confidence: facts.ConfidenceHigh,
		}
		return
	}

	// 4. React
	if ev := depEvidence("react"); ev != "" {
		result.Framework = facts.Fact{
			Value:      "React",
			Evidence:   ev,
			Confidence: facts.ConfidenceHigh,
		}
		return
	}

	// None detected
	result.Framework = facts.Fact{
		Value:      "None",
		Evidence:   "none",
		Confidence: facts.ConfidenceLow,
	}
}
