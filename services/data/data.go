package main

import (
	"github.com/tidepool-org/platform/application"
	dataServiceService "github.com/tidepool-org/platform/data/service/service"
)

func main() {
	application.RunAndExit(dataServiceService.NewStandard(), "service")
}
