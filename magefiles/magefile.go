////go:build mage

package main

import (
	"errors"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Check

// Tools runs asdf install
func Tools() error {
	return sh.Run("asdf", "install")
}

// Check runs pre-commit run --all-files
func Check() error {
	return sh.Run("pre-commit", "run", "--all-files")
}

// Test runs check, then go test -v ./...
func Test() error {
	mg.Deps(Check)
	return sh.Run("go", "test", "-v", "./...")
}

// Lint runs golangci-lint, codespell, and govulncheck
func Lint() error {
	return errors.Join(
		sh.Run("golangci-lint", "run", "./..."),
		// sh.Run("codespell"),
		sh.Run("go", "tool", "golang.org/x/vuln/cmd/govulncheck", "./..."),
	)
}
