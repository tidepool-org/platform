package main

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
