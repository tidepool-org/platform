package main

import (
	"github.com/tidepool-org/platform/application"
	prescriptionApplication "github.com/tidepool-org/platform/prescription/application"
)

func main() {
	application.RunAndExit(prescriptionApplication.New(), "service")
}
