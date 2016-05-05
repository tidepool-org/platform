package main

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"os"

	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/log"
)

func main() {
	logger := log.RootLogger()

	server, err := server.New(logger)
	if err != nil {
		logger.WithError(err).Error("Failure creating dataservices server")
		os.Exit(1)
	}
	defer server.Close()

	if err := server.Run(); err != nil {
		logger.WithError(err).Error("Failure running dataservices server")
		os.Exit(1)
	}
}
