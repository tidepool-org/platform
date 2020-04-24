package application

// WARNING: Concurrent modification of these global variables is unsupported (eg. multiple parallel tests)

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/version"
)

var (
	VersionBase           string
	VersionShortCommit    string
	VersionFullCommit     string
	VersionReporterModule = fx.Provide(NewVersionReporter)
)

func NewVersionReporter() (version.Reporter, error) {
	return version.NewReporter(VersionBase, VersionShortCommit, VersionFullCommit)
}
