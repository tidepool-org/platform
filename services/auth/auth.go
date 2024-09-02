package main

import (
	"github.com/tidepool-org/platform/application"
	authService "github.com/tidepool-org/platform/auth/service/service"
)

func main() {
	application.RunAndExit(authService.New(), "service")
}
