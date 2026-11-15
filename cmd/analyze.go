package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"repojet/internal/analyzer"
	"repojet/internal/facts"
	gitmanager "repojet/internal/git"
	"repojet/internal/planner"
	"repojet/internal/repository"
	"repojet/internal/workspace"

	"github.com/spf13/cobra"
)

var jsonFlag bool

type AnalysisResult struct {
	Repository string             `json:"repository"`
	Facts      *facts.Facts       `json:"facts"`
	Plan       *planner.SetupPlan `json:"plan"`
}

// DefaultAnalyzers returns the registered runtime analyzers.
func DefaultAnalyzers() []analyzer.Analyzer {
	return []analyzer.Analyzer{
		analyzer.NewNodeAnalyzer(),
	}
}

// AnalyzeLocal analyzes an already-cloned or local repository path.
// This allows testing against local fixtures without requiring network or cloning.
func AnalyzeLocal(repoPath string, repoIdentifier string, analyzers []analyzer.Analyzer) (*AnalysisResult, error) {
	if len(analyzers) == 0 {
		analyzers = DefaultAnalyzers()
	}

	var selectedAnalyzer analyzer.Analyzer
	for _, a := range analyzers {
		if a.Detect(repoPath) {
			selectedAnalyzer = a
			break
		}
	}

	if selectedAnalyzer == nil {
		return nil, fmt.Errorf("unsupported project type: no matching analyzer found")
	}

	detectedFacts, err := selectedAnalyzer.Analyze(repoPath)
	if err != nil {
		return nil, err
	}

	plan, err := planner.Generate(detectedFacts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate setup plan: %w", err)
	}

	return &AnalysisResult{
		Repository: repoIdentifier,
		Facts:      detectedFacts,
		Plan:       plan,
	}, nil
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze <github-url>",
	Short: "Analyze a GitHub repository and generate a setup plan",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		repositoryURL := args[0]

		// Step 1: Validate repository URL.
		if err := repository.ValidateURL(repositoryURL); err != nil {
			return err
		}

		// Step 2: Verify Git is installed.
		if err := gitmanager.CheckInstalled(); err != nil {
			return err
		}

		// Step 3: Create temporary workspace.
		workspacePath, err := workspace.Create()
		if err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Step 4: Guarantee workspace cleanup on all exit paths.
		defer workspace.Remove(workspacePath)

		// Step 5: Shallow clone with timeout context.
		fmt.Fprintln(cmd.ErrOrStderr(), "Cloning repository...")

		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
		defer cancel()

		if err := gitmanager.CloneContext(ctx, repositoryURL, workspacePath); err != nil {
			return err
		}

		fmt.Fprintln(cmd.ErrOrStderr(), "Repository cloned successfully.")
		fmt.Fprintln(cmd.ErrOrStderr(), "Analyzing repository...")

		// Step 6: Run analysis and planning on the cloned workspace.
		result, err := AnalyzeLocal(workspacePath, repositoryURL, DefaultAnalyzers())
		if err != nil {
			return err
		}

		// Step 7: Print output.
		if jsonFlag {
			return PrintJSON(cmd.OutOrStdout(), result)
		}

		PrintReport(cmd.OutOrStdout(), result)
		return nil
	},
}

// PrintJSON outputs the analysis result as indented JSON.
func PrintJSON(w io.Writer, result *AnalysisResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// PrintReport outputs a human-readable summary of the repository analysis.
func PrintReport(w io.Writer, result *AnalysisResult) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Repository")
	fmt.Fprintln(w, "------------------------------")
	fmt.Fprintf(w, "URL: %s\n", result.Repository)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Project")
	fmt.Fprintln(w, "------------------------------")
	printFactLine(w, "Runtime", result.Facts.Runtime)
	printFactLine(w, "Framework", result.Facts.Framework)
	printFactLine(w, "Package Manager", result.Facts.PackageManager)
	printFactLine(w, "Node Requirement", result.Facts.NodeVersion)
	if result.Facts.Lockfile != "" {
		fmt.Fprintf(w, "%-18s %s\n", "Lockfile:", result.Facts.Lockfile)
	}
	if len(result.Facts.EnvFiles) > 0 {
		fmt.Fprintf(w, "%-18s %s\n", "Env Files:", strings.Join(result.Facts.EnvFiles, ", "))
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Commands")
	fmt.Fprintln(w, "------------------------------")
	printCommandStatus(w, "Install:", result.Plan.Commands.Install)
	printCommandStatus(w, "Dev:", result.Plan.Commands.Dev)
	printCommandStatus(w, "Build:", result.Plan.Commands.Build)
	printCommandStatus(w, "Start:", result.Plan.Commands.Start)
	printCommandStatus(w, "Test:", result.Plan.Commands.Test)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Setup Plan")
	fmt.Fprintln(w, "------------------------------")
	for _, step := range result.Plan.Steps {
		fmt.Fprintf(w, "%d. %s\n", step.Order, step.Title)
		fmt.Fprintf(w, "   %s\n", step.Command)
	}
	fmt.Fprintln(w)

	if len(result.Plan.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings")
		fmt.Fprintln(w, "------------------------------")
		for _, warning := range result.Plan.Warnings {
			fmt.Fprintf(w, "- %s\n", warning)
		}
		fmt.Fprintln(w)
	}
}

func printFactLine(w io.Writer, label string, fact facts.Fact) {
	val := fact.Value
	if val == "" {
		val = "none"
	}

	details := make([]string, 0, 2)
	if fact.Evidence != "" && fact.Evidence != "none" {
		details = append(details, fact.Evidence)
	}
	if fact.Confidence != "" {
		details = append(details, fmt.Sprintf("%s confidence", fact.Confidence))
	}

	if len(details) > 0 {
		fmt.Fprintf(w, "%-18s %s (%s)\n", label+":", val, strings.Join(details, ", "))
	} else {
		fmt.Fprintf(w, "%-18s %s\n", label+":", val)
	}
}

func printCommandStatus(w io.Writer, label, command string) {
	if command != "" {
		fmt.Fprintf(w, "%-10s %s\n", label, command)
	} else {
		fmt.Fprintf(w, "%-10s (none)\n", label)
	}
}

func init() {
	analyzeCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output analysis and setup plan as JSON")
	rootCmd.AddCommand(analyzeCmd)
}
