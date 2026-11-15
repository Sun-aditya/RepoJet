package repository

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name          string
		repositoryURL string
		wantError     bool
		errorContains string
	}{
		{
			name:          "valid GitHub repository URL",
			repositoryURL: "https://github.com/example/project",
			wantError:     false,
		},
		{
			name:          "valid GitHub repository URL with git suffix",
			repositoryURL: "https://github.com/example/project.git",
			wantError:     false,
		},
		{
			name:          "valid GitHub repository URL with trailing slash",
			repositoryURL: "https://github.com/example/project/",
			wantError:     false,
		},
		{
			name:          "valid GitHub repository URL with uppercase owner and repo",
			repositoryURL: "https://github.com/Facebook/React",
			wantError:     false,
		},
		{
			name:          "valid GitHub repository URL with uppercase host",
			repositoryURL: "https://GITHUB.COM/example/project",
			wantError:     false,
		},
		{
			name:          "invalid URL with tree and branch subpath",
			repositoryURL: "https://github.com/example/project/tree/main/packages/core",
			wantError:     true,
			errorContains: "subdirectories or branches in URL are not supported",
		},
		{
			name:          "invalid URL with blob subpath",
			repositoryURL: "https://github.com/example/project/blob/main/README.md",
			wantError:     true,
			errorContains: "subdirectories or branches in URL are not supported",
		},
		{
			name:          "invalid plain text",
			repositoryURL: "hello",
			wantError:     true,
			errorContains: "repository URL must use HTTPS",
		},
		{
			name:          "invalid HTTP URL",
			repositoryURL: "http://github.com/example/project",
			wantError:     true,
			errorContains: "repository URL must use HTTPS",
		},
		{
			name:          "unsupported Git provider",
			repositoryURL: "https://gitlab.com/example/project",
			wantError:     true,
			errorContains: "only GitHub repositories are currently supported",
		},
		{
			name:          "missing owner and repository",
			repositoryURL: "https://github.com/",
			wantError:     true,
			errorContains: "format https://github.com/<owner>/<repository>",
		},
		{
			name:          "missing repository",
			repositoryURL: "https://github.com/example",
			wantError:     true,
			errorContains: "format https://github.com/<owner>/<repository>",
		},
		{
			name:          "extra path after repository",
			repositoryURL: "https://github.com/example/project/issues",
			wantError:     true,
			errorContains: "format https://github.com/<owner>/<repository>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateURL(test.repositoryURL)

			if test.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if test.errorContains != "" && !strings.Contains(err.Error(), test.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), test.errorContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
