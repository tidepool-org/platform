package application

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription/api"
	"github.com/tidepool-org/platform/prescription/container"
	serviceService "github.com/tidepool-org/platform/service/service"
)

type Application struct {
	*serviceService.Authenticated
	container container.Container
}

func New() *Application {
	authenticated := serviceService.NewAuthenticated()
	params := &container.Params{
		ConfigReporter:  authenticated.ConfigReporter(),
		Logger:          authenticated.Logger(),
		UserAgent:       authenticated.UserAgent(),
		VersionReporter: authenticated.VersionReporter(),
	}
	return &Application{
		Authenticated: authenticated,
		container:     container.New(params),
	}
}

func (a *Application) Initialize(provider application.Provider) error {
	if err := a.Authenticated.Initialize(provider); err != nil {
		return err
	}
	if err := a.initializeRouter(); err != nil {
		return err
	}
	return a.initializeContainer()
}

func (a *Application) initializeRouter() error {
	a.Logger().Debug("Creating prescription router")

	router, err := api.NewRouter(a.container)
	if err != nil {
		return errors.Wrap(err, "unable to create prescription api router")
	}

	a.Logger().Debug("Initializing router")
	if err = a.API().InitializeRouters(router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (a *Application) initializeContainer() error {
	a.Logger().Debug("Initializing application container")
	if err := a.container.Initialize(); err != nil {
		return errors.Wrap(err, "unable to initialize application container")
	}

	return nil
}

func (a *Application) Terminate() {
	a.Service.Terminate()
}
