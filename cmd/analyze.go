package cmd

import (
	"fmt"

	"repojet/internal/analyzer/packagemanager"
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
		printScanResult(facts)

		fmt.Println("Repository scan completed.")

		// Step 8: Detect package manager.
		fmt.Println("Detecting package manager...")

		packageManagerResult, err := packagemanager.Detect(facts)
		if err != nil {
			return fmt.Errorf(
				"failed to detect package manager: %w",
				err,
			)
		}

		// Step 9: Print package manager result.
		printPackageManagerResult(packageManagerResult)

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

func printPackageManagerResult(result *packagemanager.Result) {
	fmt.Println()
	fmt.Println("Package Manager")
	fmt.Println("------------------------------")

	if result == nil {
		fmt.Println("Name:    unknown")
		fmt.Println("Version: unknown")
		fmt.Println("Source:  none")
		fmt.Println()
		return
	}

	fmt.Println("Name:   ", result.Name)

	if result.Version != "" {
		fmt.Println("Version:", result.Version)
	} else {
		fmt.Println("Version: unknown")
	}

	fmt.Println("Source: ", result.Source)
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
