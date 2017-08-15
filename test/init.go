package test

import "os"

func init() {
	Init()
}

func Init() {
	if os.Getenv("TIDEPOOL_ENV") != "test" {
		panic(`Test packages only supported in test environment!!!`)
	}
}
