package test

import (
	"os"

	"github.com/tidepool-org/platform/application/version"
)

func init() {
	if os.Getenv("TIDEPOOL_ENV") != "test" {
		panic(`Test packages only supported while running in test environment (TIDEPOOL_ENV="test")`)
	}

	version.Base = "0.0.0"
	version.ShortCommit = "0000000"
	version.FullCommit = "0000000000000000000000000000000000000000"
}
