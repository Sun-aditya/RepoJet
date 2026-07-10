package cmd

import (
	"fmt"

	gitmanager "repojet/internal/git"
	"repojet/internal/repository"
	"repojet/internal/scanner"
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

		// Step 5: Clone repository.
		fmt.Println("Cloning repository...")

		if err := gitmanager.Clone(repositoryURL, workspacePath); err != nil {
			return err
		}

		fmt.Println("Repository cloned successfully.")

		// Step 6: Scan repository.
		fmt.Println("Scanning repository...")

		facts, err := scanner.Scan(workspacePath)
		if err != nil {
			return fmt.Errorf("failed to scan repository: %w", err)
		}

		// Step 7: Print scan result.
		if facts.PackageJSON != "" {
			fmt.Println("package.json: found")
			fmt.Println("Path:", facts.PackageJSON)
		} else {
			fmt.Println("package.json: not found")
		}

		fmt.Println("Repository scan completed.")

		return nil
	},
}

func printScanResult(facts *scanner.RepositoryFacts) {
	fmt.Println()
	fmt.Println("Detected Files")
	fmt.Println("------------------------------")

	printFileStatus("package.json", facts.PackageJSON)
	printFileStatus("package-lock.json", facts.PackageLock)
	printFileStatus("pnpm-lock.yaml", facts.PnpmLock)
	printFileStatus("yarn.lock", facts.YarnLock)

	printFileStatus(".nvmrc", facts.Nvmrc)
	printFileStatus(".node-version", facts.NodeVersion)

	printFileStatus("Dockerfile", facts.Dockerfile)
	printFileStatus("docker-compose", facts.DockerCompose)

	printFileStatus(".env.example", facts.EnvExample)
	printFileStatus("README.md", facts.Readme)

	printFileStatus("prisma/schema.prisma", facts.PrismaSchema)

	fmt.Println()
}

func printFileStatus(name string, path string) {
	if path != "" {
		fmt.Printf("%-24s found\n", name)
		return
	}

	fmt.Printf("%-24s not found\n", name)
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}