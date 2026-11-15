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
	ExpectedFramework string `json:"expected_framework,omitempty"`
	ExpectedPM        string `json:"expected_package_manager,omitempty"`
	IsNonNode         bool   `json:"is_non_node"`
}

type BenchmarkResult struct {
	Target              BenchmarkTarget `json:"target"`
	Success             bool            `json:"success"`
	Error               string          `json:"error,omitempty"`
	ErrorType           string          `json:"error_type,omitempty"`
	DetectedFramework   string          `json:"detected_framework,omitempty"`
	DetectedPM          string          `json:"detected_package_manager,omitempty"`
	DetectedNodeVersion string          `json:"detected_node_version,omitempty"`
	DurationMs          int64           `json:"duration_ms"`
	FrameworkMatched    bool            `json:"framework_matched"`
	PMMatched           bool            `json:"pm_matched"`
	CleanFailure        bool            `json:"clean_failure"`
}

type BenchmarkSummary struct {
	TotalRepos            int               `json:"total_repos"`
	NodeRepos             int               `json:"node_repos"`
	NonNodeRepos          int               `json:"non_node_repos"`
	SuccessfulAnalyses    int               `json:"successful_analyses"`
	FrameworkMatches      int               `json:"framework_matches"`
	PMMatches             int               `json:"pm_matches"`
	CleanFailures         int               `json:"clean_failures"`
	FrameworkAccuracyPct  float64           `json:"framework_accuracy_pct"`
	PMAccuracyPct         float64           `json:"pm_accuracy_pct"`
	CleanFailureRatePct   float64           `json:"clean_failure_rate_pct"`
	AverageDurationMs     int64             `json:"average_duration_ms"`
	Results               []BenchmarkResult `json:"results"`
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
	var totalDurationMs int64

	for i, target := range targets {
		fmt.Printf("[%2d/%2d] Analyzing %s ... ", i+1, len(targets), target.URL)
		res := benchmarkOne(target, time.Duration(*timeoutSec)*time.Second)
		results = append(results, res)
		totalDurationMs += res.DurationMs

		if res.Success {
			fmt.Printf("OK (%dms) -> %s / %s\n", res.DurationMs, res.DetectedFramework, res.DetectedPM)
		} else if res.CleanFailure {
			fmt.Printf("CLEAN FAIL (%dms) -> %s\n", res.DurationMs, res.ErrorType)
		} else {
			fmt.Printf("FAIL (%dms) -> %s\n", res.DurationMs, res.Error)
		}
	}

	summary := computeSummary(results, totalDurationMs)

	// Save results.json
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

	analysis, err := cmd.AnalyzeLocal(workspacePath, target.URL, cmd.DefaultAnalyzers())
	if err != nil {
		res.Error = err.Error()
		if strings.Contains(err.Error(), "unsupported project type") {
			res.ErrorType = "unsupported_project_type"
			if target.IsNonNode {
				res.CleanFailure = true
			}
		} else if strings.Contains(err.Error(), "no package.json") {
			res.ErrorType = "no_package_json"
			if target.IsNonNode {
				res.CleanFailure = true
			}
		} else {
			res.ErrorType = "analysis_error"
		}
		return res
	}

	// Successful analysis
	res.Success = true
	res.DetectedFramework = analysis.Facts.Framework.Value
	res.DetectedPM = analysis.Facts.PackageManager.Value
	res.DetectedNodeVersion = analysis.Facts.NodeVersion.Value

	if target.ExpectedFramework != "" {
		res.FrameworkMatched = strings.EqualFold(target.ExpectedFramework, res.DetectedFramework)
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
		target := BenchmarkTarget{URL: urlStr}

		if len(parts) > 1 {
			comment := strings.TrimSpace(parts[1])
			if strings.HasPrefix(comment, "expected:") {
				exp := strings.TrimSpace(strings.TrimPrefix(comment, "expected:"))
				if strings.EqualFold(exp, "non-node") {
					target.IsNonNode = true
				} else {
					expParts := strings.Split(exp, ",")
					if len(expParts) >= 1 {
						target.ExpectedFramework = strings.TrimSpace(expParts[0])
					}
					if len(expParts) >= 2 {
						target.ExpectedPM = strings.TrimSpace(expParts[1])
					}
				}
			}
		}

		targets = append(targets, target)
	}

	return targets, scanner.Err()
}

func computeSummary(results []BenchmarkResult, totalDurationMs int64) BenchmarkSummary {
	s := BenchmarkSummary{
		TotalRepos: len(results),
		Results:    results,
	}

	if len(results) > 0 {
		s.AverageDurationMs = totalDurationMs / int64(len(results))
	}

	frameworkTotal := 0
	pmTotal := 0

	for _, r := range results {
		if r.Target.IsNonNode {
			s.NonNodeRepos++
			if r.CleanFailure {
				s.CleanFailures++
			}
		} else {
			s.NodeRepos++
			if r.Success {
				s.SuccessfulAnalyses++
			}
			if r.Target.ExpectedFramework != "" {
				frameworkTotal++
				if r.FrameworkMatched {
					s.FrameworkMatches++
				}
			}
			if r.Target.ExpectedPM != "" {
				pmTotal++
				if r.PMMatched {
					s.PMMatches++
				}
			}
		}
	}

	if frameworkTotal > 0 {
		s.FrameworkAccuracyPct = (float64(s.FrameworkMatches) / float64(frameworkTotal)) * 100.0
	}
	if pmTotal > 0 {
		s.PMAccuracyPct = (float64(s.PMMatches) / float64(pmTotal)) * 100.0
	}
	if s.NonNodeRepos > 0 {
		s.CleanFailureRatePct = (float64(s.CleanFailures) / float64(s.NonNodeRepos)) * 100.0
	}

	return s
}

func generateMarkdown(s BenchmarkSummary) string {
	var b strings.Builder

	b.WriteString("# RepoJet Benchmark Results\n\n")
	b.WriteString(fmt.Sprintf("Run on %d public GitHub repositories.\n\n", s.TotalRepos))

	b.WriteString("## Summary\n\n")
	b.WriteString("| Metric | Result |\n")
	b.WriteString("| :--- | :--- |\n")
	b.WriteString(fmt.Sprintf("| **Total Repositories** | %d |\n", s.TotalRepos))
	b.WriteString(fmt.Sprintf("| **Node.js Repositories** | %d |\n", s.NodeRepos))
	b.WriteString(fmt.Sprintf("| **Successful Node.js Analyses** | %d / %d (%.1f%%) |\n", s.SuccessfulAnalyses, s.NodeRepos, (float64(s.SuccessfulAnalyses)/float64(s.NodeRepos))*100.0))
	b.WriteString(fmt.Sprintf("| **Framework Detection Accuracy** | %d / %d (%.1f%%) |\n", s.FrameworkMatches, s.NodeRepos, s.FrameworkAccuracyPct))
	b.WriteString(fmt.Sprintf("| **Package Manager Accuracy** | %d / %d (%.1f%%) |\n", s.PMMatches, s.NodeRepos, s.PMAccuracyPct))
	b.WriteString(fmt.Sprintf("| **Non-Node Repositories** | %d |\n", s.NonNodeRepos))
	b.WriteString(fmt.Sprintf("| **Non-Node Clean Failure Rate** | %d / %d (%.1f%%) |\n", s.CleanFailures, s.NonNodeRepos, s.CleanFailureRatePct))
	b.WriteString(fmt.Sprintf("| **Average Analysis Duration** | %d ms |\n\n", s.AverageDurationMs))

	b.WriteString("## Repository Details\n\n")
	b.WriteString("| Repository | Expected Framework | Detected Framework | Expected PM | Detected PM | Duration | Status |\n")
	b.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")

	for _, r := range s.Results {
		repoName := filepath.Base(r.Target.URL)
		parts := strings.Split(strings.TrimPrefix(r.Target.URL, "https://github.com/"), "/")
		if len(parts) >= 2 {
			repoName = parts[0] + "/" + parts[1]
		}

		expectedFw := r.Target.ExpectedFramework
		if r.Target.IsNonNode {
			expectedFw = "Non-Node"
		} else if expectedFw == "" {
			expectedFw = "-"
		}

		expectedPM := r.Target.ExpectedPM
		if r.Target.IsNonNode {
			expectedPM = "Non-Node"
		} else if expectedPM == "" {
			expectedPM = "-"
		}

		detectedFw := r.DetectedFramework
		if detectedFw == "" {
			detectedFw = "-"
		}

		detectedPM := r.DetectedPM
		if detectedPM == "" {
			detectedPM = "-"
		}

		status := "PASS"
		if r.Target.IsNonNode {
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
			"| `%s` | %s | %s | %s | %s | %dms | %s |\n",
			repoName,
			expectedFw,
			detectedFw,
			expectedPM,
			detectedPM,
			r.DurationMs,
			status,
		))
	}

	// Failures / Mismatches list
	var discrepancies []string
	for _, r := range s.Results {
		if r.Target.IsNonNode {
			if !r.CleanFailure {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s**: Expected clean failure, but got: %s", r.Target.URL, r.Error))
			}
		} else {
			if !r.Success {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s**: Failed with %s: %s", r.Target.URL, r.ErrorType, r.Error))
			} else if !r.FrameworkMatched || !r.PMMatched {
				discrepancies = append(discrepancies, fmt.Sprintf("- **%s**: Expected %s / %s, detected %s / %s", r.Target.URL, r.Target.ExpectedFramework, r.Target.ExpectedPM, r.DetectedFramework, r.DetectedPM))
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
