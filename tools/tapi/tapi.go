package main

import (
	"fmt"
	"os"

	"github.com/tidepool-org/platform/tools/tapi/cmd"
)

// TODO: TECH DEBT - Convert to new tool package paradigm

func main() {
	application, err := cmd.InitializeApplication()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	if err = application.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
