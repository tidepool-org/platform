package main

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data/service/service"
)

func main() {
	application.Run(service.NewStandard("dataservice", "TIDEPOOL"))
}
