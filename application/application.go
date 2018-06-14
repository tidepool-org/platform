package application

import (
	"os"

	"github.com/tidepool-org/platform/errors"
)

type Application struct {
	Provider
}

func New() *Application {
	return &Application{}
}

func (a *Application) Initialize(provider Provider) error {
	if provider == nil {
		return errors.New("provider is missing")
	}

	a.Provider = provider

	return nil
}

func (a *Application) Terminate() {
	os.Stdout.Sync()
	os.Stderr.Sync()
}
