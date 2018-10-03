package main

import (
	"github.com/tidepool-org/platform/application"
	notificationServiceService "github.com/tidepool-org/platform/notification/service/service"
)

func main() {
	application.RunAndExit(notificationServiceService.New(), "service")
}
