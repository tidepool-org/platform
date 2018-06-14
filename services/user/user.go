package main

import (
	"github.com/tidepool-org/platform/application"
	userServiceService "github.com/tidepool-org/platform/user/service/service"
)

func main() {
	application.RunAndExit(userServiceService.NewStandard(), "service")
}
