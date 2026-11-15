package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
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
		"-c", "core.hooksPath=/dev/null",
		"clone",
		"--depth", "1",
		"--no-recurse-submodules",
		repositoryURL,
		destination,
	)

	// Prevent git from prompting for terminal credentials and hanging
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// Ensure the process is forcefully killed if context times out or is canceled
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	cmd.WaitDelay = 2 * time.Second

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
