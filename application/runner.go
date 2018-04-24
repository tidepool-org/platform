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

const (
	Success = 0
	Failure = 1
)

func Run(runner Runner, err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to create:", err)
		return Failure
	}
	if runner == nil {
		fmt.Fprintln(os.Stderr, "ERROR: Runner is missing")
		return Failure
	}

	defer runner.Terminate()

	if err = runner.Initialize(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to initialize:", err)
		return Failure
	}

	if err = runner.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to run:", err)
		return Failure
	}

	return Success
}
