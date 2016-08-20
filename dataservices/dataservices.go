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
	"fmt"
	"os"

	"github.com/tidepool-org/platform/dataservices/service/service"
)

func main() {
	standardService, err := service.NewStandard()
	if err != nil {
		fmt.Printf("ERROR: Failure creating service: %s\n", err.Error())
		os.Exit(1)
	}
	defer standardService.Close()

	if err = standardService.Initialize(); err != nil {
		fmt.Printf("ERROR: Failure initializing service: %s\n", err.Error())
		os.Exit(1)
	}

	if err = standardService.Run(); err != nil {
		fmt.Printf("ERROR: Failure running service: %s\n", err.Error())
		os.Exit(1)
	}
}
