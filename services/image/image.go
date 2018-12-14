package main

import (
	"github.com/tidepool-org/platform/application"
	imageService "github.com/tidepool-org/platform/image/service"
)

func main() {
	application.RunAndExit(imageService.New(), "service")
}
