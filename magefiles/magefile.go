//go:build mage

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Run

// Tools runs asdf install
func Tools() error {
	return sh.Run("asdf", "install")
}

// Build runs check, then compiles the binary to bin/gba
func Build() error {
	mg.Deps(Check)
	if err := os.MkdirAll("bin", 0755); err != nil {
		return err
	}
	return sh.Run("go", "build", "-o", "bin/gba", ".")
}

// Check runs pre-commit run --all-files
func Check() error {
	return sh.Run("pre-commit", "run", "--all-files")
}

// Clean removes bin/gba
func Clean() error {
	return sh.Rm("bin/gba")
}

// Install runs build, then copies gba to user (false) or system (true) bin directory
func Install(global bool) error {
	mg.Deps(Build)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var destDir, binName string
	switch runtime.GOOS {
	case "windows":
		binName = "gba.exe"
		if global {
			destDir = filepath.Join(os.Getenv("ProgramFiles"), "gba")
		} else {
			destDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "gba")
		}
	default: // linux, darwin
		binName = "gba"
		if global {
			destDir = "/usr/local/bin"
		} else {
			destDir = filepath.Join(home, ".local", "bin")
		}
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	return sh.Copy(filepath.Join(destDir, binName), "bin/gba")
}

// Test runs check, then go test -v ./...
func Test() error {
	mg.Deps(Check)
	return sh.Run("go", "test", "-v", "./...")
}

// Run runs build, then executes bin/gba
func Run() error {
	mg.Deps(Build)
	return sh.Run("./bin/gba")
}

// Lint runs golangci-lint, codespell, and govulncheck
func Lint() error {
	if err := sh.Run("golangci-lint", "run", "./..."); err != nil {
		return err
	}
	if err := sh.Run("codespell"); err != nil {
		return err
	}
	return sh.Run("go", "tool", "golang.org/x/vuln/cmd/govulncheck", "./...")
}
