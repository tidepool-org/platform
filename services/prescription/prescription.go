package main

import (
	"github.com/tidepool-org/platform/application"
	prescriptionServiceService "github.com/tidepool-org/platform/prescription/service/service"
)

func main() {
	application.RunAndExit(prescriptionServiceService.New(), "service")
}
