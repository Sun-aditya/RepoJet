package git

import (
	"context"
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
	return CloneContext(context.Background(), repositoryURL, destination)
}

func CloneContext(ctx context.Context, repositoryURL string, destination string) error {
	cmd := exec.CommandContext(
		ctx,
		"git",
		"clone",
		"--depth",
		"1",
		repositoryURL,
		destination,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("git clone timed out or was canceled: %w", ctx.Err())
		}
		return fmt.Errorf(
			"failed to clone repository: %w: %s",
			err,
			string(output),
		)
	}

	return nil
}
