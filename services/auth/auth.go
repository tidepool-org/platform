package main

import (
	"github.com/tidepool-org/platform/application"
	authServiceService "github.com/tidepool-org/platform/auth/service/service"
)

func main() {
	application.RunAndExit(authServiceService.New(), "service")
}
