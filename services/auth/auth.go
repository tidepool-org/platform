package main

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth/service/service"
)

func main() {
	application.Run(service.New("TIDEPOOL"))
}
