package analyzer

import "repojet/internal/facts"

// Analyzer defines the generic interface for inspecting a repository and extracting facts.
// Keeping this generic allows Node.js, Python, and future runtime analyzers to be added easily.
type Analyzer interface {
	Name() string
	Detect(repoPath string) bool
	Analyze(repoPath string) (*facts.Facts, error)
}
