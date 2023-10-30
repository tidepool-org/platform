//go:build tools
// +build tools

package tools

import (
	_ "github.com/githubnemo/CompileDaemon" // CompileDaemon - for hot reloading
	_ "github.com/mjibson/esc"              // Esc - for embedding static assets TODO: is this used?
	_ "github.com/onsi/ginkgo/v2"           // Gingko - for running tests
	_ "golang.org/x/lint/golint"            // Golint - for Lint checks
	_ "golang.org/x/tools/cmd/goimports"    // Goimports - to check and fix missing or unused imports
)
