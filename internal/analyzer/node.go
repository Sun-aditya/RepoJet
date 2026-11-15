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
	Main            string            `json:"main"`
	Exports         any               `json:"exports"`
	Bin             any               `json:"bin"`
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

	// Project Type Detection (application, library, or unknown)
	a.detectProjectType(manifest, result)

	// Workspaces check (monorepo root)
	isMonorepo := manifest.Workspaces != nil
	if isMonorepo {
		result.AddWarning("workspaces detected in package.json (monorepo root detected; framework detection at root may reflect monorepo tooling or doc sites)")
		subprojects := a.discoverSubprojects(repoFacts)
		if len(subprojects) > 0 {
			result.AddWarning(fmt.Sprintf("%d workspace subproject(s) discovered: %s", len(subprojects), strings.Join(subprojects, ", ")))
		}
	}

	// Package Manager & Lockfile Detection
	a.detectPackageManagerAndLockfile(manifest, repoFacts, result)

	// Node Version Detection
	a.detectNodeVersion(manifest, repoFacts, result)

	// Framework Detection (Primary and Secondary frameworks)
	a.detectFrameworks(manifest, repoFacts, isMonorepo, result)

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

func (a *NodeAnalyzer) detectProjectType(manifest packageJSONManifest, result *facts.Facts) {
	hasDevOrStart := false
	if manifest.Scripts != nil {
		if strings.TrimSpace(manifest.Scripts["dev"]) != "" || strings.TrimSpace(manifest.Scripts["start"]) != "" {
			hasDevOrStart = true
		}
	}

	hasEntrypoint := strings.TrimSpace(manifest.Main) != "" || manifest.Exports != nil || manifest.Bin != nil

	if hasEntrypoint && !hasDevOrStart {
		result.Type = facts.Fact{
			Value:      "library",
			Evidence:   "package.json (main/exports/bin, no dev/start script)",
			Confidence: facts.ConfidenceHigh,
		}
	} else if hasDevOrStart {
		result.Type = facts.Fact{
			Value:      "application",
			Evidence:   "package.json (dev/start scripts)",
			Confidence: facts.ConfidenceHigh,
		}
	} else {
		result.Type = facts.Fact{
			Value:      "unknown",
			Evidence:   "package.json",
			Confidence: facts.ConfidenceLow,
		}
	}
}

func (a *NodeAnalyzer) discoverSubprojects(repoFacts *scanner.RepositoryFacts) []string {
	var subs []string
	for _, f := range repoFacts.Files {
		clean := filepath.ToSlash(f)
		if strings.HasSuffix(clean, "package.json") && clean != "package.json" {
			subDir := filepath.Dir(clean)
			subs = append(subs, subDir)
		}
	}
	return subs
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
		result.PackageManagerVersion = pmResult.Version

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

	if strings.HasPrefix(strings.ToLower(v1), "lts") || strings.HasPrefix(strings.ToLower(v2), "lts") {
		return false
	}
	if strings.ContainsAny(v1, "><=^~*| -") || strings.ContainsAny(v2, "><=^~*| -") {
		return false
	}

	major1 := strings.Split(v1, ".")[0]
	major2 := strings.Split(v2, ".")[0]
	return major1 != major2
}

type frameworkCandidate struct {
	name       string
	evidence   string
	confidence facts.Confidence
	priority   int // lower = higher priority
}

func (a *NodeAnalyzer) detectFrameworks(
	manifest packageJSONManifest,
	repoFacts *scanner.RepositoryFacts,
	isMonorepo bool,
	result *facts.Facts,
) {
	var candidates []frameworkCandidate

	// Check script contents for build/dev runners (highest confidence)
	scriptContains := func(token string) string {
		if manifest.Scripts == nil {
			return ""
		}
		for _, sName := range []string{"dev", "start", "build"} {
			cmd := manifest.Scripts[sName]
			words := strings.Fields(cmd)
			for _, w := range words {
				if w == token || strings.HasPrefix(w, token) {
					return fmt.Sprintf("package.json (scripts: %s)", sName)
				}
			}
		}
		return ""
	}

	depEvidence := func(depName string) string {
		if manifest.Dependencies != nil {
			if _, ok := manifest.Dependencies[depName]; ok {
				return fmt.Sprintf("package.json (dependencies: %s)", depName)
			}
		}
		if manifest.DevDependencies != nil {
			if _, ok := manifest.DevDependencies[depName]; ok {
				return fmt.Sprintf("package.json (devDependencies: %s)", depName)
			}
		}
		return ""
	}

	// 1. Next.js
	if ev := scriptContains("next"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Next.js", evidence: ev, confidence: facts.ConfidenceHigh, priority: 1})
	} else if repoFacts.NextConfig != "" {
		candidates = append(candidates, frameworkCandidate{name: "Next.js", evidence: filepath.Base(repoFacts.NextConfig), confidence: facts.ConfidenceHigh, priority: 1})
	} else if ev := depEvidence("next"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Next.js", evidence: ev, confidence: facts.ConfidenceMedium, priority: 1})
	}

	// 2. Vite
	if ev := scriptContains("vite"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Vite", evidence: ev, confidence: facts.ConfidenceHigh, priority: 2})
	} else if repoFacts.ViteConfig != "" {
		candidates = append(candidates, frameworkCandidate{name: "Vite", evidence: filepath.Base(repoFacts.ViteConfig), confidence: facts.ConfidenceHigh, priority: 2})
	} else if ev := depEvidence("vite"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Vite", evidence: ev, confidence: facts.ConfidenceMedium, priority: 2})
	}

	// 3. React
	if ev := scriptContains("react-scripts"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "React", evidence: ev, confidence: facts.ConfidenceHigh, priority: 3})
	} else if ev := depEvidence("react"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "React", evidence: ev, confidence: facts.ConfidenceMedium, priority: 3})
	}

	// 4. Server Frameworks: Express, Fastify, Koa, Nest
	// IMPORTANT: Server frameworks count ONLY when in "dependencies", NOT "devDependencies"
	serverDep := func(depName string) string {
		if manifest.Dependencies != nil {
			if _, ok := manifest.Dependencies[depName]; ok {
				return fmt.Sprintf("package.json (dependencies: %s)", depName)
			}
		}
		return ""
	}

	if ev := serverDep("express"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Express", evidence: ev, confidence: facts.ConfidenceHigh, priority: 4})
	}
	if ev := serverDep("fastify"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Fastify", evidence: ev, confidence: facts.ConfidenceHigh, priority: 5})
	}
	if ev := serverDep("koa"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "Koa", evidence: ev, confidence: facts.ConfidenceHigh, priority: 6})
	}
	if ev := serverDep("@nestjs/core"); ev != "" {
		candidates = append(candidates, frameworkCandidate{name: "NestJS", evidence: ev, confidence: facts.ConfidenceHigh, priority: 7})
	}

	if len(candidates) == 0 {
		result.Framework = facts.Fact{
			Value:      "None",
			Evidence:   "none",
			Confidence: facts.ConfidenceLow,
		}
		return
	}

	// Sort candidates by priority
	bestIdx := 0
	for i := 1; i < len(candidates); i++ {
		if candidates[i].priority < candidates[bestIdx].priority {
			bestIdx = i
		}
	}

	primary := candidates[bestIdx]
	confidence := primary.confidence
	if isMonorepo {
		confidence = facts.ConfidenceLow
	}

	result.Framework = facts.Fact{
		Value:      primary.name,
		Evidence:   primary.evidence,
		Confidence: confidence,
	}

	// Populate secondary frameworks for any other unique candidates
	for i, c := range candidates {
		if i == bestIdx || c.name == primary.name {
			continue
		}
		// If primary is Next.js, React is naturally its underlying library; include as secondary if detected
		cConfidence := c.confidence
		if isMonorepo {
			cConfidence = facts.ConfidenceLow
		}
		result.SecondaryFrameworks = append(result.SecondaryFrameworks, facts.Fact{
			Value:      c.name,
			Evidence:   c.evidence,
			Confidence: cConfidence,
		})
	}
}
