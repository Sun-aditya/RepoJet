package git

import (
	"fmt"
	"os/exec"
)

func CheckInstalled() error {
	_, err := exec.LookPath("git")

	if err != nil {
		return fmt.Errorf("git is not installed or not available in PATH: %w", err)
	}

	return nil
}

func Clone(repositoryURL string, destination string) error {
	cmd := exec.Command(
		"git",
		"clone",
		"--depth",
		"1",
		repositoryURL,
		destination,
	)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"failed to clone repository: %w: %s",
			err,
			string(output),
		)
	}

	return nil
}