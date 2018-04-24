package version

// WARNING: Concurrent modification of these global variables is unsupported (eg. multiple parallel tests)

import "github.com/tidepool-org/platform/version"

var (
	Base        string
	ShortCommit string
	FullCommit  string
)

func NewReporter() (version.Reporter, error) {
	return version.NewReporter(Base, ShortCommit, FullCommit)
}
