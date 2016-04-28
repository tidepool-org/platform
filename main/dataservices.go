package main

import (
	"github.com/tidepool-org/platform/dataservices"
	"github.com/tidepool-org/platform/log"
)

func main() {
	logger := log.RootLogger()

	server, err := dataservices.NewServer(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failure creating dataservices server")
	}
	defer server.Close()

	if err := server.Run(); err != nil {
		logger.WithError(err).Fatal("Failure running dataservices server")
	}
}
