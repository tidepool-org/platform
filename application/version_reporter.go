package application

// WARNING: Concurrent modification of these global variables is unsupported (eg. multiple parallel tests)

import "github.com/tidepool-org/platform/version"

var (
	VersionBase        string
	VersionShortCommit string
	VersionFullCommit  string
)

func NewVersionReporter() (version.Reporter, error) {
	return version.NewReporter(VersionBase, VersionShortCommit, VersionFullCommit)
}
