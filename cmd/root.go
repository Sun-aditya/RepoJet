package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repojet",
	Short: "Analyze and run repositories with minimal setup",
	Long: `RepoJet analyzes software repositories, determines their
runtime and dependency requirements, and prepares them to run.`,

	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}