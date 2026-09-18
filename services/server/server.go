package main

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/service/combined"
)

func main() {
	application.RunAndExit(combined.New(), "service")
}
