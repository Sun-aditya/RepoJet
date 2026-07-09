package cmd

import (
	"fmt"

	gitmanager "repojet/internal/git"
	"repojet/internal/repository"
	"repojet/internal/workspace"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <github-url>",
	Short: "Analyze a GitHub repository",
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

		// Step 4: Guarantee workspace cleanup.
		defer workspace.Remove(workspacePath)

		// Step 5: Clone repository into the workspace.
		fmt.Println("Cloning repository...")

		if err := gitmanager.Clone(repositoryURL, workspacePath); err != nil {
			return err
		}

		fmt.Println("Repository cloned successfully.")
		fmt.Println("Workspace:", workspacePath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}