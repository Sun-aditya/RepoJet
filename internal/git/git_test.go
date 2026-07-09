package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCheckInstalled(t *testing.T) {
	err := CheckInstalled()

	if err != nil {
		t.Fatalf("CheckInstalled() returned an error: %v", err)
	}
}

func TestClone(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := filepath.Join(t.TempDir(), "cloned-repository")

	// Initialize a local Git repository.
	cmd := exec.Command("git", "init", sourceDir)

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf(
			"failed to initialize source repository: %v\n%s",
			err,
			output,
		)
	}

	// Create a file so the repository has content to clone.
	testFile := filepath.Join(sourceDir, "README.md")

	err := os.WriteFile(
		testFile,
		[]byte("# RepoJet Test Repository"),
		0644,
	)

	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Stage the file.
	cmd = exec.Command("git", "-C", sourceDir, "add", "README.md")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf(
			"failed to stage test file: %v\n%s",
			err,
			output,
		)
	}

	// Create the commit with local identity configuration.
	cmd = exec.Command(
		"git",
		"-C",
		sourceDir,
		"-c",
		"user.name=RepoJet Test",
		"-c",
		"user.email=repojet-test@example.com",
		"commit",
		"-m",
		"initial commit",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf(
			"failed to create test commit: %v\n%s",
			err,
			output,
		)
	}

	// Clone using our actual RepoJet Clone() function.
	err = Clone(sourceDir, destinationDir)

	if err != nil {
		t.Fatalf("Clone() returned an error: %v", err)
	}

	// Verify that the cloned repository contains README.md.
	clonedFile := filepath.Join(destinationDir, "README.md")

	_, err = os.Stat(clonedFile)

	if err != nil {
		t.Fatalf("expected cloned file does not exist: %v", err)
	}
}

func TestCloneFailure(t *testing.T) {
	destinationDir := filepath.Join(t.TempDir(), "cloned-repository")

	err := Clone(
		"/path/that/does/not/exist/repository.git",
		destinationDir,
	)

	if err == nil {
		t.Fatal("Clone() expected an error, got nil")
	}
}