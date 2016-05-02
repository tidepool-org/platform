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
