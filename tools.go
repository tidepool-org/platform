//go:build tools
// +build tools

package tools

import (
	_ "github.com/githubnemo/CompileDaemon" // CompileDaemon - for hot reloading
	_ "github.com/onsi/ginkgo/v2"           // Gingko - for running tests
	_ "go.uber.org/mock/mockgen"            // Mockgen - for test mocks
	_ "go.uber.org/mock/mockgen/model"      // Mockgen - for test mocks
	_ "golang.org/x/lint/golint"            // Golint - for Lint checks
	_ "golang.org/x/tools/cmd/goimports"    // Goimports - to check and fix missing or unused imports
)
