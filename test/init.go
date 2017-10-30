package test

import "os"

func init() {
	if os.Getenv("TIDEPOOL_ENV") != "test" {
		panic(`Test packages only supported in test environment!!!`)
	}
}
