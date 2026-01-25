# Agent Instructions for `git-bug-ax`

This document provides instructions for AI agents to effectively work within the `git-bug-ax` repository.

## Project Overview

This is a Go project that uses `mage` for build automation. The main application logic is in `main.go` and the `pkg/` directory, while build-related scripts are in `magefiles/`.

## Tooling Setup

This project uses `asdf` to manage runtime and tool versions, which are defined in the `.tool-versions` file. To install all the required tools, run:

```bash
mage tools
```
This command will run `asdf install` to ensure all necessary tools like Go, mage, and pre-commit are installed with the correct versions.

## Essential Commands

All project tasks are managed through `mage`. You must have `mage` installed to run these commands (see Tooling Setup).

- **Build the project**:
  ```bash
  mage build
  ```
  This command compiles the Go binary and places it in the `bin/` directory.

- **Run tests**:
  ```bash
  mage test
  ```
  This command runs the test suite. It also runs `pre-commit` checks first.

- **Run linters and checks**:
  ```bash
  mage check
  ```
  This runs `pre-commit` hooks on all files.

  ```bash
  mage lint
  ```
  This runs `golangci-lint`, `codespell`, and `govulncheck`.

- **Run the application**:
  ```bash
  mage run
  ```
  This builds and runs the main application.

- **Clean build artifacts**:
  ```bash
  mage clean
  ```
  This removes the compiled binary from the `bin/` directory.

- **Install the application**:
  ```bash
  mage install
  ```
  This installs the binary to the user's local bin directory. Use `mage "install global=true"` for a system-wide installation.

## Code Structure

- `main.go`: The main entry point for the application.
- `pkg/`: Contains reusable library code organized by functionality.
- `magefiles/`: Contains the `magefile.go` which defines the build, test, and run commands.
- `go.mod`, `go.sum`: Go module files defining project dependencies.
- `bin/`: This directory contains the compiled application binary (e.g., `bin/gbax`). It is created by the build process.

## Code Style and Conventions

- The project uses `pre-commit` to enforce code style and quality. Run `mage check` before submitting changes.
- All build, test, and linting logic is centralized in `magefiles/magefile.go`. To understand how a process works, refer to this file.

## Gotchas

- Ensure you have `mage` and `pre-commit` installed and available in your `PATH`.
- Always run `mage check` or `mage test` before making commits to ensure your changes pass the required checks.
