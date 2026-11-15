package repository

import (
	"fmt"
	"net/url"
	"strings"
)

func ValidateURL(repositoryURL string) error {
	parsedURL, err := url.Parse(repositoryURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("repository URL must use HTTPS")
	}

	if strings.ToLower(parsedURL.Host) != "github.com" {
		return fmt.Errorf("only GitHub repositories are currently supported")
	}

	trimmedPath := strings.Trim(parsedURL.Path, "/")
	if trimmedPath == "" {
		return fmt.Errorf("repository URL must have the format https://github.com/<owner>/<repository>")
	}

	parts := strings.Split(trimmedPath, "/")

	if len(parts) > 2 {
		if parts[2] == "tree" || parts[2] == "blob" {
			return fmt.Errorf(
				"subdirectories or branches in URL are not supported; provide the repository root URL (https://github.com/%s/%s)",
				parts[0],
				strings.TrimSuffix(parts[1], ".git"),
			)
		}
		return fmt.Errorf("repository URL must have the format https://github.com/<owner>/<repository>")
	}

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repository URL must have the format https://github.com/<owner>/<repository>")
	}

	return nil
}
