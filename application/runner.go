package application

import (
	"fmt"
	"os"

	"github.com/tidepool-org/platform/errors"
)

type Runner interface {
	Initialize(provider Provider) error
	Terminate()

	Run() error
}

func RunAndExit(runner Runner, scopes ...string) {
	provider, err := NewProvider("TIDEPOOL", scopes...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	if err = Run(runner, provider); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func Run(runner Runner, provider Provider) error {
	if runner == nil {
		return errors.New("runner is missing")
	}
	if provider == nil {
		return errors.New("provider is missing")
	}

	defer runner.Terminate()

	if err := runner.Initialize(provider); err != nil {
		return errors.Wrap(err, "unable to initialize runner")
	}

	if err := runner.Run(); err != nil {
		return errors.Wrap(err, "unable to run runner")
	}

	return nil
}
