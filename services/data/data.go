package main

import (
	"os"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data/service/service"
)

func main() {
	os.Exit(application.Run(service.NewStandard("TIDEPOOL")))
}
