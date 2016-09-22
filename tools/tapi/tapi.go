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

	"github.com/tidepool-org/platform/tools/tapi/cmd"
)

func main() {
	application, err := cmd.InitializeApplication()
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	if err = application.Run(os.Args); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}
