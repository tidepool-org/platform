package application

import (
	"fmt"
	"os"
)

type Runner interface {
	Initialize() error
	Terminate()

	Run() error
}

func Run(runner Runner, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to create:", err)
		os.Exit(1)
	}
	if runner == nil {
		fmt.Fprintln(os.Stderr, "ERROR: Runner is missing")
		os.Exit(1)
	}

	defer runner.Terminate()

	if err = runner.Initialize(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to initialize:", err)
		os.Exit(1)
	}

	if err = runner.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to run:", err)
		os.Exit(1)
	}
}
