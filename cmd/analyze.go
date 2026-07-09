package cmd

import (
	"fmt"

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

		err := repository.ValidateURL(repositoryURL)
		if err != nil {
			return err
		}

		workspacePath, err := workspace.Create()
		if err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		defer workspace.Remove(workspacePath)

		fmt.Println("Repository URL is valid:")
		fmt.Println(repositoryURL)

		fmt.Println("Temporary workspace created:")
		fmt.Println(workspacePath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}