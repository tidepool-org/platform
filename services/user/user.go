package main

import (
	"github.com/tidepool-org/platform/application"
	userService "github.com/tidepool-org/platform/user/service"
)

func main() {
	application.RunAndExit(userService.New(), "service")
}
