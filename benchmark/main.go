package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"repojet/cmd"
	gitmanager "repojet/internal/git"
	"repojet/internal/repository"
	"repojet/internal/workspace"
)

type BenchmarkTarget struct {
	URL               string `json:"url"`
	Set               string `json:"set"`      // "tuning" or "held-out"
	Category          string `json:"category"` // "app", "library", "monorepo", "no-manifest", "non-node"
	ExpectedFramework string `json:"expected_framework,omitempty"`
	ExpectedPM        string `json:"expected_package_manager,omitempty"`
	ExpectedType      string `json:"expected_type,omitempty"`
}

type BenchmarkResult struct {
	Target              BenchmarkTarget `json:"target"`
	Cloned              bool            `json:"cloned"`
	Success             bool            `json:"success"`
	CleanFailure        bool            `json:"clean_failure"`
	Error               string          `json:"error,omitempty"`
	ErrorType           string          `json:"error_type,omitempty"`
	DetectedType        string          `json:"detected_type,omitempty"`
	DetectedFramework   string          `json:"detected_framework,omitempty"`
	SecondaryFrameworks []string        `json:"secondary_frameworks,omitempty"`
	DetectedPM          string          `json:"detected_package_manager,omitempty"`
	DetectedNodeVersion string          `json:"detected_node_version,omitempty"`
	DurationMs          int64           `json:"duration_ms"`
	FrameworkMatched    bool            `json:"framework_matched"`
	PMMatched           bool            `json:"pm_matched"`
	TypeMatched         bool            `json:"type_matched"`
}

type SetMetrics struct {
	TotalRepos          int     `json:"total_repos"`
	CloneSuccess        int     `json:"clone_success"`
	CloneSuccessPct     float64 `json:"clone_success_pct"`
	AnalysisSuccess     int     `json:"analysis_success"`
	AnalysisSuccessPct   float64 `json:"analysis_success_pct"` // of cloned
	FrameworkMatches    int     `json:"framework_matches"`
	FrameworkAccuracy   float64 `json:"framework_accuracy_pct"` // of analyzed
	PMMatches           int     `json:"pm_matches"`
	PMAccuracy          float64 `json:"pm_accuracy_pct"` // of analyzed
	TypeMatches         int     `json:"type_matches"`
	TypeAccuracy        float64 `json:"type_accuracy_pct"` // of analyzed
	CleanFailures       int     `json:"clean_failures"`
	CleanFailureRatePct float64 `json:"clean_failure_rate_pct"`

	// Category breakdowns for framework accuracy
	AppRepos            int     `json:"app_repos"`
	AppFrameworkMatches int     `json:"app_framework_matches"`
	AppAccuracyPct      float64 `json:"app_accuracy_pct"`

	LibRepos            int     `json:"lib_repos"`
	LibFrameworkMatches int     `json:"lib_framework_matches"`
	LibAccuracyPct      float64 `json:"lib_accuracy_pct"`

	MonoRepos            int     `json:"monorepos"`
	MonoFrameworkMatches int     `json:"monorepo_framework_matches"`
	MonoAccuracyPct      float64 `json:"monorepo_accuracy_pct"`
}

type BenchmarkSummary struct {
	Overall SetMetrics        `json:"overall"`
	Tuning  SetMetrics        `json:"tuning"`
	HeldOut SetMetrics        `json:"held_out"`
	Results []BenchmarkResult `json:"results"`
}

func main() {
	reposFile := flag.String("repos", "benchmark/repos.txt", "Path to repos.txt")
	timeoutSec := flag.Int("timeout", 60, "Timeout per repo in seconds")
	flag.Parse()

	targets, err := parseReposFile(*reposFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading repos file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting RepoJet benchmark with %d repositories...\n\n", len(targets))

	var results []BenchmarkResult

	for i, target := range targets {
		fmt.Printf("[%2d/%2d] [%s/%s] Analyzing %s ... ", i+1, len(targets), target.Set, target.Category, target.URL)
		res := benchmarkOne(target, time.Duration(*timeoutSec)*time.Second)
		results = append(results, res)

		if res.Success {
			fmt.Printf("OK (%dms) -> %s (%s) / %s\n", res.DurationMs, res.DetectedFramework, res.DetectedType, res.DetectedPM)
		} else if res.CleanFailure {
			fmt.Printf("CLEAN FAIL (%dms) -> %s\n", res.DurationMs, res.ErrorType)
		} else {
			fmt.Printf("FAIL (%dms) -> %s\n", res.DurationMs, res.Error)
		}
	}

	summary := computeSummary(results)

	// Save results.json (ignored by git)
	resultsJSON, _ := json.MarshalIndent(summary, "", "  ")
	_ = os.WriteFile("benchmark/results.json", resultsJSON, 0644)

	// Save RESULTS.md
	markdown := generateMarkdown(summary)
	if err := os.WriteFile("benchmark/RESULTS.md", []byte(markdown), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing benchmark/RESULTS.md: %v\n", err)
	}

	// Print summary
	fmt.Println("\n" + markdown)
}

func benchmarkOne(target BenchmarkTarget, timeout time.Duration) BenchmarkResult {
	start := time.Now()
	res := BenchmarkResult{
		Target: target,
	}

	defer func() {
		res.DurationMs = time.Since(start).Milliseconds()
	}()

	if err := repository.ValidateURL(target.URL); err != nil {
		res.Error = err.Error()
		res.ErrorType = "invalid_url"
		return res
	}

	if err := gitmanager.CheckInstalled(); err != nil {
		res.Error = err.Error()
		res.ErrorType = "git_missing"
		return res
	}

	workspacePath, err := workspace.Create()
	if err != nil {
		res.Error = err.Error()
		res.ErrorType = "workspace_create_error"
		return res
	}
	defer workspace.Remove(workspacePath)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := gitmanager.CloneContext(ctx, target.URL, workspacePath); err != nil {
		res.Error = err.Error()
		res.ErrorType = "clone_error"
		return res
	}

	res.Cloned = true

	analysis, err := cmd.AnalyzeLocal(workspacePath, target.URL, cmd.DefaultAnalyzers())
	if err != nil {
		res.Error = err.Error()
		if strings.Contains(err.Error(), "unsupported project type") {
			res.ErrorType = "unsupported_project_type"
			if target.Category == "non-node" || target.Category == "no-manifest" {
				res.CleanFailure = true
			}
		} else if strings.Contains(err.Error(), "no package.json") {
			res.ErrorType = "no_package_json"
			if target.Category == "non-node" || target.Category == "no-manifest" {
				res.CleanFailure = true
			}
		} else {
			res.ErrorType = "analysis_error"
		}
		return res
	}

	// Successful analysis
	res.Success = true
	res.DetectedType = analysis.Facts.Type.Value
	res.DetectedFramework = analysis.Facts.Framework.Value
	res.DetectedPM = analysis.Facts.PackageManager.Value
	res.DetectedNodeVersion = analysis.Facts.NodeVersion.Value

	for _, sf := range analysis.Facts.SecondaryFrameworks {
		res.SecondaryFrameworks = append(res.SecondaryFrameworks, sf.Value)
	}

	if target.ExpectedType != "" {
		res.TypeMatched = strings.EqualFold(target.ExpectedType, res.DetectedType)
	}

	if target.ExpectedFramework != "" {
		if strings.EqualFold(target.ExpectedFramework, res.DetectedFramework) {
			res.FrameworkMatched = true
		} else {
			// Check if secondary frameworks contain the expected framework
			for _, sf := range res.SecondaryFrameworks {
				if strings.EqualFold(target.ExpectedFramework, sf) {
					res.FrameworkMatched = true
					break
				}
			}
		}
	}

	if target.ExpectedPM != "" {
		res.PMMatched = strings.EqualFold(target.ExpectedPM, res.DetectedPM)
	}

	return res
}

func parseReposFile(path string) ([]BenchmarkTarget, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []BenchmarkTarget
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "#", 2)
		urlStr := strings.TrimSpace(parts[0])
		target := BenchmarkTarget{
			URL: urlStr,
			Set: "tuning",
		}

		if len(parts) > 1 {
			tagsStr := strings.TrimSpace(parts[1])
			tags := strings.Split(tagsStr, ",")
			for _, tag := range tags {
				kv := strings.SplitN(strings.TrimSpace(tag), ":", 2)
				if len(kv) != 2 {
					continue
				}
				key := strings.TrimSpace(kv[0])
				val := strings.TrimSpace(kv[1])

				switch key {
				case "set":
					target.Set = val
				case "category":
					target.Category = val
				case "expected_framework":
					target.ExpectedFramework = val
				case "expected_pm":
					target.ExpectedPM = val
				case "expected_type":
					target.ExpectedType = val
				}
			}
		}

		targets = append(targets, target)
	}

	return targets, scanner.Err()
}

func computeMetrics(results []BenchmarkResult) SetMetrics {
	var m SetMetrics
	m.TotalRepos = len(results)

	nonNodeOrNoManifest := 0
	analyzedCount := 0

	for _, r := range results {
		if r.Cloned {
			m.CloneSuccess++
		}

		if r.Target.Category == "non-node" || r.Target.Category == "no-manifest" {
			nonNodeOrNoManifest++
			if r.CleanFailure {
				m.CleanFailures++
			}
			continue
		}

		if r.Success {
			m.AnalysisSuccess++
			analyzedCount++

			if r.FrameworkMatched {
				m.FrameworkMatches++
			}
			if r.PMMatched {
				m.PMMatches++
			}
			if r.TypeMatched {
				m.TypeMatches++
			}

			// Category tracking
			switch r.Target.Category {
			case "app":
				m.AppRepos++
				if r.FrameworkMatched {
					m.AppFrameworkMatches++
				}
			case "library":
				m.LibRepos++
				if r.FrameworkMatched {
					m.LibFrameworkMatches++
				}
			case "monorepo":
				m.MonoRepos++
				if r.FrameworkMatched {
					m.MonoFrameworkMatches++
				}
			}
		}
	}

	clonedNodeRepos := m.CloneSuccess - nonNodeOrNoManifest

	if m.TotalRepos > 0 {
		m.CloneSuccessPct = (float64(m.CloneSuccess) / float64(m.TotalRepos)) * 100.0
	}
	if clonedNodeRepos > 0 {
		m.AnalysisSuccessPct = (float64(m.AnalysisSuccess) / float64(clonedNodeRepos)) * 100.0
	}
	if analyzedCount > 0 {
		m.FrameworkAccuracy = (float64(m.FrameworkMatches) / float64(analyzedCount)) * 100.0
		m.PMAccuracy = (float64(m.PMMatches) / float64(analyzedCount)) * 100.0
		m.TypeAccuracy = (float64(m.TypeMatches) / float64(analyzedCount)) * 100.0
	}
	if nonNodeOrNoManifest > 0 {
		m.CleanFailureRatePct = (float64(m.CleanFailures) / float64(nonNodeOrNoManifest)) * 100.0
	}

	if m.AppRepos > 0 {
		m.AppAccuracyPct = (float64(m.AppFrameworkMatches) / float64(m.AppRepos)) * 100.0
	}
	if m.LibRepos > 0 {
		m.LibAccuracyPct = (float64(m.LibFrameworkMatches) / float64(m.LibRepos)) * 100.0
	}
	if m.MonoRepos > 0 {
		m.MonoAccuracyPct = (float64(m.MonoFrameworkMatches) / float64(m.MonoRepos)) * 100.0
	}

	return m
}

func computeSummary(results []BenchmarkResult) BenchmarkSummary {
	var tuningResults []BenchmarkResult
	var heldOutResults []BenchmarkResult

	for _, r := range results {
		if r.Target.Set == "held-out" {
			heldOutResults = append(heldOutResults, r)
		} else {
			tuningResults = append(tuningResults, r)
		}
	}

	return BenchmarkSummary{
		Overall: computeMetrics(results),
		Tuning:  computeMetrics(tuningResults),
		HeldOut: computeMetrics(heldOutResults),
		Results: results,
	}
}

func generateMarkdown(s BenchmarkSummary) string {
	var b strings.Builder

	b.WriteString("# RepoJet Benchmark Results (Step 7b)\n\n")
	b.WriteString(fmt.Sprintf("Evaluated on %d public GitHub repositories: **%d Tuning Set** and **%d Held-out Set**.\n\n",
		s.Overall.TotalRepos, s.Tuning.TotalRepos, s.HeldOut.TotalRepos))

	b.WriteString("## Benchmark Metrics Comparison\n\n")
	b.WriteString("| Metric | Tuning Set (30) | Held-out Set (10) | Overall (40) |\n")
	b.WriteString("| :--- | :--- | :--- | :--- |\n")
	b.WriteString(fmt.Sprintf("| **Clone Success** | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.CloneSuccess, s.Tuning.TotalRepos, s.Tuning.CloneSuccessPct,
		s.HeldOut.CloneSuccess, s.HeldOut.TotalRepos, s.HeldOut.CloneSuccessPct,
		s.Overall.CloneSuccess, s.Overall.TotalRepos, s.Overall.CloneSuccessPct))
	b.WriteString(fmt.Sprintf("| **Analysis Success (of Cloned)** | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.AnalysisSuccess, s.Tuning.CloneSuccess-4, s.Tuning.AnalysisSuccessPct,
		s.HeldOut.AnalysisSuccess, s.HeldOut.CloneSuccess-2, s.HeldOut.AnalysisSuccessPct,
		s.Overall.AnalysisSuccess, s.Overall.CloneSuccess-6, s.Overall.AnalysisSuccessPct))
	b.WriteString(fmt.Sprintf("| **Framework Accuracy (of Analyzed)** | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.FrameworkMatches, s.Tuning.AnalysisSuccess, s.Tuning.FrameworkAccuracy,
		s.HeldOut.FrameworkMatches, s.HeldOut.AnalysisSuccess, s.HeldOut.FrameworkAccuracy,
		s.Overall.FrameworkMatches, s.Overall.AnalysisSuccess, s.Overall.FrameworkAccuracy))
	b.WriteString(fmt.Sprintf("| &nbsp;&nbsp;↳ *Applications* | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.AppFrameworkMatches, s.Tuning.AppRepos, s.Tuning.AppAccuracyPct,
		s.HeldOut.AppFrameworkMatches, s.HeldOut.AppRepos, s.HeldOut.AppAccuracyPct,
		s.Overall.AppFrameworkMatches, s.Overall.AppRepos, s.Overall.AppAccuracyPct))
	b.WriteString(fmt.Sprintf("| &nbsp;&nbsp;↳ *Libraries* | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.LibFrameworkMatches, s.Tuning.LibRepos, s.Tuning.LibAccuracyPct,
		s.HeldOut.LibFrameworkMatches, s.HeldOut.LibRepos, s.HeldOut.LibAccuracyPct,
		s.Overall.LibFrameworkMatches, s.Overall.LibRepos, s.Overall.LibAccuracyPct))
	b.WriteString(fmt.Sprintf("| &nbsp;&nbsp;↳ *Monorepos* | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.MonoFrameworkMatches, s.Tuning.MonoRepos, s.Tuning.MonoAccuracyPct,
		s.HeldOut.MonoFrameworkMatches, s.HeldOut.MonoRepos, s.HeldOut.MonoAccuracyPct,
		s.Overall.MonoFrameworkMatches, s.Overall.MonoRepos, s.Overall.MonoAccuracyPct))
	b.WriteString(fmt.Sprintf("| **Package Manager Accuracy (of Analyzed)** | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.PMMatches, s.Tuning.AnalysisSuccess, s.Tuning.PMAccuracy,
		s.HeldOut.PMMatches, s.HeldOut.AnalysisSuccess, s.HeldOut.PMAccuracy,
		s.Overall.PMMatches, s.Overall.AnalysisSuccess, s.Overall.PMAccuracy))
	b.WriteString(fmt.Sprintf("| **Project Type Accuracy (of Analyzed)** | %d / %d (%.1f%%) | %d / %d (%.1f%%) | %d / %d (%.1f%%) |\n",
		s.Tuning.TypeMatches, s.Tuning.AnalysisSuccess, s.Tuning.TypeAccuracy,
		s.HeldOut.TypeMatches, s.HeldOut.AnalysisSuccess, s.HeldOut.TypeAccuracy,
		s.Overall.TypeMatches, s.Overall.AnalysisSuccess, s.Overall.TypeAccuracy))
	b.WriteString(fmt.Sprintf("| **Clean Failure (Non-Node / No-Manifest)** | %d / 4 (%.1f%%) | %d / 2 (%.1f%%) | %d / 6 (%.1f%%) |\n\n",
		s.Tuning.CleanFailures, s.Tuning.CleanFailureRatePct,
		s.HeldOut.CleanFailures, s.HeldOut.CleanFailureRatePct,
		s.Overall.CleanFailures, s.Overall.CleanFailureRatePct))

	b.WriteString("## Repository Details\n\n")
	b.WriteString("| Set | Category | Repository | Expected Fw | Detected Fw | Expected PM | Detected PM | Type | Status |\n")
	b.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")

	for _, r := range s.Results {
		repoName := filepath.Base(r.Target.URL)
		parts := strings.Split(strings.TrimPrefix(r.Target.URL, "https://github.com/"), "/")
		if len(parts) >= 2 {
			repoName = parts[0] + "/" + parts[1]
		}

		expectedFw := r.Target.ExpectedFramework
		if r.Target.Category == "non-node" || r.Target.Category == "no-manifest" {
			expectedFw = "-"
		} else if expectedFw == "" {
			expectedFw = "-"
		}

		expectedPM := r.Target.ExpectedPM
		if r.Target.Category == "non-node" || r.Target.Category == "no-manifest" {
			expectedPM = "-"
		} else if expectedPM == "" {
			expectedPM = "-"
		}

		detectedFw := r.DetectedFramework
		if len(r.SecondaryFrameworks) > 0 {
			detectedFw += " (+" + strings.Join(r.SecondaryFrameworks, ",") + ")"
		}
		if detectedFw == "" {
			detectedFw = "-"
		}

		detectedPM := r.DetectedPM
		if detectedPM == "" {
			detectedPM = "-"
		}

		typeStr := r.DetectedType
		if typeStr == "" {
			typeStr = "-"
		}

		status := "PASS"
		if r.Target.Category == "non-node" || r.Target.Category == "no-manifest" {
			if r.CleanFailure {
				status = "PASS (clean failure)"
			} else {
				status = "FAIL"
			}
		} else if !r.Success {
			status = "ERROR (" + r.ErrorType + ")"
		} else if !r.FrameworkMatched || !r.PMMatched {
			status = "MISMATCH"
		}

		b.WriteString(fmt.Sprintf(
			"| %s | %s | `%s` | %s | %s | %s | %s | %s | %s |\n",
			r.Target.Set,
			r.Target.Category,
			repoName,
			expectedFw,
			detectedFw,
			expectedPM,
			detectedPM,
			typeStr,
			status,
		))
	}

	// Discrepancies detail
	var discrepancies []string
	for _, r := range s.Results {
		if r.Target.Category == "non-node" || r.Target.Category == "no-manifest" {
			if !r.CleanFailure {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s** (%s): Expected clean failure, got error: %s", r.Target.URL, r.Target.Set, r.Error))
			}
		} else {
			if !r.Success {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s** (%s): Failed to analyze: %s", r.Target.URL, r.Target.Set, r.Error))
			} else if !r.FrameworkMatched || !r.PMMatched {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s** (%s): Expected framework %q / PM %q, detected framework %q / PM %q",
					r.Target.URL, r.Target.Set, r.Target.ExpectedFramework, r.Target.ExpectedPM, r.DetectedFramework, r.DetectedPM))
			}
		}
	}

	b.WriteString("\n## Failures & Discrepancies\n\n")
	if len(discrepancies) == 0 {
		b.WriteString("None! All repositories analyzed with 100% expected results.\n")
	} else {
		for _, d := range discrepancies {
			b.WriteString(d + "\n")
		}
	}

	return b.String()
}
