package main

import (
	"github.com/tidepool-org/platform/application"
	taskServiceService "github.com/tidepool-org/platform/task/service/service"
)

func main() {
	application.RunAndExit(taskServiceService.New(), "service")
}
