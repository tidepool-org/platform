package test

import (
	"github.com/tidepool-org/platform/application/version"
	"github.com/tidepool-org/platform/test"
)

func init() {
	test.Init()

	version.Base = "0.0.0"
	version.ShortCommit = "0000000"
	version.FullCommit = "0000000000000000000000000000000000000000"
}
