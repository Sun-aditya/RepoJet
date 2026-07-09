package repository

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name          string
		repositoryURL string
		wantError     bool
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
			name:          "invalid plain text",
			repositoryURL: "hello",
			wantError:     true,
		},
		{
			name:          "invalid HTTP URL",
			repositoryURL: "http://github.com/example/project",
			wantError:     true,
		},
		{
			name:          "unsupported Git provider",
			repositoryURL: "https://gitlab.com/example/project",
			wantError:     true,
		},
		{
			name:          "missing owner and repository",
			repositoryURL: "https://github.com/",
			wantError:     true,
		},
		{
			name:          "missing repository",
			repositoryURL: "https://github.com/example",
			wantError:     true,
		},
		{
			name:          "extra path after repository",
			repositoryURL: "https://github.com/example/project/issues",
			wantError:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateURL(test.repositoryURL)

			gotError := err != nil

			if gotError != test.wantError {
				t.Errorf(
					"ValidateURL(%q) error = %v, wantError = %v",
					test.repositoryURL,
					err,
					test.wantError,
				)
			}
		})
	}
}