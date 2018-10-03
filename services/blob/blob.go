package main

import (
	"github.com/tidepool-org/platform/application"
	blobService "github.com/tidepool-org/platform/blob/service"
)

func main() {
	application.RunAndExit(blobService.New(), "service")
}
