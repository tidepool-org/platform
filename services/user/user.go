package main

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/user/service/service"
)

func main() {
	application.Run(service.NewStandard("TIDEPOOL"))
}
