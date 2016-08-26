package main

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"fmt"
	"os"

	"github.com/tidepool-org/platform/dataservices/service/service"
)

func main() {
	standard, err := service.NewStandard()
	if err != nil {
		fmt.Println("ERROR: Unable to create service:", err)
		os.Exit(1)
	}
	defer standard.Terminate()

	if err = standard.Initialize(); err != nil {
		fmt.Println("ERROR: Unable to initialize service:", err)
		os.Exit(1)
	}

	if err = standard.Run(); err != nil {
		fmt.Println("ERROR: Unable to run service:", err)
		os.Exit(1)
	}
}
