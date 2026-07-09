# RepoJet

RepoJet is an open-source command-line tool that aims to simplify the process of setting up and running software repositories locally.

The long-term goal is to allow developers to provide a repository URL and let RepoJet analyze the project, determine its runtime and dependency requirements, prepare an isolated development environment, and start the project with minimal manual setup.

> RepoJet is currently in early development. The first version focuses on analyzing JavaScript and TypeScript repositories hosted on GitHub.

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

RepoJet aims to automate as much of this workflow as possible.

The intended experience is:

```bash
repojet run https://github.com/owner/repository
```

RepoJet will eventually analyze the repository, generate a setup plan, prepare an isolated environment, install dependencies, start required services, run the application, and report the local URL.

## Current Status

RepoJet is currently in the first stage of development.

The current CLI supports:

- A Cobra-based command-line interface.
- The `analyze` command.
- GitHub repository URL validation.
- Temporary workspace creation and automatic cleanup.
- Basic Git availability detection.
- Basic shallow-clone functionality.
- Automated tests for repository URL validation and workspace management.

The current command is:

```bash
go run . analyze https://github.com/owner/repository
```

At this stage, RepoJet does not yet analyze repository contents or start applications.

## Current Workflow

```text
repojet analyze <github-url>
        |
        v
Validate GitHub URL
        |
        v
Create Temporary Workspace
        |
        v
Clone Repository
        |
        v
Repository Analysis
        |
        v
Generate Setup Plan
```

Repository cloning, analysis, and setup-plan generation are being implemented incrementally.

## Planned Analysis Pipeline

The first working analyzer will focus on Node.js repositories.

```text
GitHub Repository
        |
        v
Clone Repository
        |
        v
Scan Important Files
        |
        v
Parse package.json
        |
        v
Detect Package Manager
        |
        v
Detect Framework
        |
        v
Detect Node.js Requirements
        |
        v
Detect Available Scripts
        |
        v
Generate Setup Plan
        |
        v
Display Analysis Result
```

Initially supported project types are planned to include:

- Node.js
- React
- Vite
- Next.js
- Express

Support for additional ecosystems will be added incrementally.

## Example

The intended analysis command is:

```bash
repojet analyze https://github.com/owner/project
```

Example output:

```text
RepoJet

Repository
--------------------------------

URL              https://github.com/owner/project

Project
--------------------------------

Runtime          Node.js
Framework        Next.js
Package Manager  pnpm
Node Requirement >=20

Commands
--------------------------------

Install          pnpm install
Development      pnpm dev
Build            pnpm build
Production       pnpm start

Setup Plan
--------------------------------

1. Install a compatible Node.js version
2. Install pnpm
3. Install project dependencies
4. Start the development server

Analysis completed successfully.
```

The example above represents the intended MVP behavior and is not fully implemented yet.

## Architecture

RepoJet follows a modular architecture.

```text
CLI
 |
 v
Input Validation
 |
 v
Workspace Manager
 |
 v
Git Manager
 |
 v
Repository Scanner
 |
 v
Repository Analyzers
 |
 v
Repository Facts
 |
 v
Setup Planner
 |
 v
Setup Plan
 |
 v
Execution Engine
 |
 v
Container Runtime
 |
 v
Running Application
```

The project separates repository understanding from execution.

The core design is:

```text
UNDERSTAND
    |
    v
PLAN
    |
    v
EXECUTE
```

RepoJet will first inspect the repository and collect facts, then generate a setup plan, and only then execute the approved plan.

## Project Structure

```text
RepoJet/
├── cmd/
│   ├── root.go
│   └── analyze.go
│
├── internal/
│   ├── git/
│   │   └── git.go
│   │
│   ├── repository/
│   │   ├── validator.go
│   │   └── validator_test.go
│   │
│   └── workspace/
│       ├── workspace.go
│       └── workspace_test.go
│
├── go.mod
├── go.sum
├── main.go
└── README.md
```

The structure will evolve as new components are implemented.

## Getting Started

### Prerequisites

Currently, development requires:

- Go
- Git
- Linux, macOS, Windows, or WSL2

Docker will become a prerequisite when isolated repository execution is implemented.

### Clone the Repository

```bash
git clone https://github.com/<your-username>/repojet.git
cd repojet
```

### Install Dependencies

```bash
go mod download
```

### Run RepoJet

```bash
go run . --help
```

Analyze a repository:

```bash
go run . analyze https://github.com/owner/repository
```

### Run Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

### Future Ideas

- `.repojet.yml` configuration files.
- Repository analysis confidence scores.
- Explainable detection results and evidence.
- AI-assisted analysis for unsupported repositories.
- GitHub authentication and private repository support.
- Cross-platform installers and package-manager distribution.
- VS Code integration.
- Local graphical dashboard.

## Development Principles

RepoJet is being built around several principles:

**Analyze before executing.**

Repository code should not be executed until RepoJet understands the project and generates a setup plan.

**Prefer deterministic detection.**

Configuration files, manifests, lockfiles, and project metadata should be used before AI-based inference.

**Keep analysis and execution separate.**

Repository analysis should produce a structured setup plan that can be inspected, tested, and later consumed by different execution engines.

**Treat repository code as untrusted.**

Future execution features should use isolation and explicit safety controls rather than blindly executing repository scripts on the host system.

**Build incrementally.**

Each component should be implemented and tested independently before being connected to the complete pipeline.

## Contributing

RepoJet is currently in early development.

Contribution guidelines and development documentation will be added as the project becomes more stable.

## License

A license has not been selected yet.