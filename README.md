# RepoJet

RepoJet is an open-source command-line tool that analyzes software repositories and generates a setup plan for running them locally.

The long-term goal is to allow developers to provide a repository URL and let RepoJet inspect the project, determine its runtime and dependency requirements, prepare an isolated development environment, and start the project with minimal manual setup.

> RepoJet currently features an end-to-end analyzer and setup planner for Node.js repositories hosted on GitHub.

## Why RepoJet?

Running an unfamiliar repository often requires several manual steps:

- Reading the project documentation.
- Cloning the repository.
- Identifying the framework and runtime.
- Determining the required runtime version.
- Identifying the package manager.
- Installing dependencies.
- Configuring environment variables.
- Starting databases and other services.
- Running setup commands and migrations.
- Finding the correct command to start the application.

RepoJet automates this workflow:

```bash
repojet run https://github.com/owner/repository
```

## Current Status

RepoJet currently supports full end-to-end analysis and setup planning for **Node.js** projects:

- **CLI Interface**: Cobra-based CLI supporting `repojet analyze <github-url>` and `--json` output.
- **GitHub URL Validation**: Validates HTTPS GitHub repository URLs, rejects unsupported subpaths (`/tree/...`, `/blob/...`), and normalizes trailing slashes.
- **Isolated Workspace**: Creates temporary workspaces with automatic cleanup on exit or error.
- **Safe Repository Cloning**: Shallow clones with `--depth 1`, `--no-recurse-submodules`, disabled hooks (`core.hooksPath=/dev/null`), `GIT_TERMINAL_PROMPT=0`, and context cancellation/timeouts.
- **Repository Scanner**: Safely walks repositories capping file reads at 1 MB, skipping `node_modules`, `.git`, and symlinks.
- **Node.js Analyzer**: Inspects `package.json`, detects frameworks, package managers, Node versions, and scripts, recording evidence and confidence for every fact.
- **Setup Planner**: Computes ordered setup steps and precise install (with frozen lockfiles), dev, build, start, and test commands.
- **Automated Tests & CI**: Comprehensive table-driven tests across fixtures (`testdata/`), golden file JSON validation, and GitHub Actions CI.

## Real Command Usage & Example Output

### Human-Readable Report

```bash
go run . analyze https://github.com/owner/repository
```

Example output:

```text
Repository
------------------------------
URL: https://github.com/owner/repository

Project
------------------------------
Runtime:           Node.js (package.json, high confidence)
Framework:         Next.js (package.json (dependencies: next), high confidence)
Package Manager:   pnpm (pnpm-lock.yaml, high confidence)
Node Requirement:  >=18 (package.json (engines.node), high confidence)
Lockfile:          pnpm-lock.yaml
Env Files:         .env.example

Commands
------------------------------
Install:   pnpm install --frozen-lockfile
Dev:       pnpm dev
Build:     pnpm build
Start:     pnpm start
Test:      pnpm test

Setup Plan
------------------------------
1. Install Node.js (>=18)
   nvm install 18
2. Enable pnpm
   corepack enable
3. Install dependencies
   pnpm install --frozen-lockfile
4. Start development server
   pnpm dev
```

### JSON Output (`--json`)

```bash
go run . analyze https://github.com/owner/repository --json
```

Example JSON output:

```json
{
  "repository": "https://github.com/owner/repository",
  "facts": {
    "runtime": {
      "value": "Node.js",
      "evidence": "package.json",
      "confidence": "high"
    },
    "framework": {
      "value": "Next.js",
      "evidence": "package.json (dependencies: next)",
      "confidence": "high"
    },
    "package_manager": {
      "value": "pnpm",
      "evidence": "pnpm-lock.yaml",
      "confidence": "high"
    },
    "node_version": {
      "value": ">=18",
      "evidence": "package.json (engines.node)",
      "confidence": "high"
    },
    "lockfile": "pnpm-lock.yaml",
    "scripts": {
      "build": "next build",
      "dev": "next dev",
      "start": "next start",
      "test": "jest"
    },
    "env_files": [
      ".env.example"
    ],
    "warnings": []
  },
  "plan": {
    "steps": [
      {
        "order": 1,
        "title": "Install Node.js (>=18)",
        "command": "nvm install 18"
      },
      {
        "order": 2,
        "title": "Enable pnpm",
        "command": "corepack enable"
      },
      {
        "order": 3,
        "title": "Install dependencies",
        "command": "pnpm install --frozen-lockfile"
      },
      {
        "order": 4,
        "title": "Start development server",
        "command": "pnpm dev"
      }
    ],
    "commands": {
      "install": "pnpm install --frozen-lockfile",
      "dev": "pnpm dev",
      "build": "pnpm build",
      "start": "pnpm start",
      "test": "pnpm test"
    },
    "warnings": []
  }
}
```

## Supported Detections

| Category | Supported Detections & Sources |
| :--- | :--- |
| **Runtime** | Node.js (`package.json`) |
| **Frameworks** | Next.js, Vite, Express, React (checks `dependencies`, `devDependencies`, and config files: `next.config.*`, `vite.config.*`) |
| **Package Managers** | `pnpm`, `npm`, `yarn`, `bun` (from `"packageManager"` field and lockfiles: `pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`, `bun.lockb`, `bun.lock`) |
| **Node Version** | `engines.node`, `.nvmrc`, `.node-version` (incompatibility warnings, range normalization) |
| **Scripts** | `dev`, `start`, `build`, `test` |
| **Environment** | `.env.example` |

## Limitations

- **Monorepos**: Workspaces (`"workspaces"` in `package.json`) are detected and flagged with a warning, but monorepos are not yet supported.
- **Analysis & Planning Only**: RepoJet generates a setup plan and command specifications, but does not yet execute code or run Docker containers.
- **Runtimes**: Currently supports Node.js. Additional ecosystems (Python, Go, Rust) will follow the generic `Analyzer` interface.

## Project Structure

```text
RepoJet/
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   ├── root.go
│   ├── analyze.go
│   └── analyze_test.go
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go
│   │   ├── node.go
│   │   ├── node_test.go
│   │   └── packagemanager/
│   │       ├── detector.go
│   │       └── detector_test.go
│   ├── facts/
│   │   ├── facts.go
│   │   └── facts_test.go
│   ├── git/
│   │   ├── git.go
│   │   └── git_test.go
│   ├── planner/
│   │   ├── planner.go
│   │   └── planner_test.go
│   ├── repository/
│   │   ├── validator.go
│   │   └── validator_test.go
│   ├── scanner/
│   │   ├── scanner.go
│   │   └── scanner_test.go
│   └── workspace/
│       ├── workspace.go
│       └── workspace_test.go
├── testdata/
│   ├── bun-project/
│   ├── express-yarn/
│   ├── golden/
│   ├── invalid-package-json/
│   ├── multiple-lockfiles/
│   ├── nextjs-pnpm/
│   ├── no-dev-start/
│   ├── no-package-json/
│   └── vite-npm/
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── Makefile
└── README.md
```

## Getting Started

### Prerequisites

- Go (1.20+)
- Git

### Build & Run

```bash
# Build binary
make build

# Run analyzer
./repojet analyze https://github.com/owner/repository

# Run analyzer with JSON output
./repojet analyze https://github.com/owner/repository --json
```

### Run Tests & Linter

```bash
# Run tests
make test

# Run vet / linter
make lint
```

## License

This project is licensed under the [MIT License](LICENSE).
