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

	if parsedURL.Host != "github.com" {
		return fmt.Errorf("only GitHub repositories are currently supported")
	}

	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("repository URL must have the format https://github.com/<owner>/<repository>")
	}

	return nil
}