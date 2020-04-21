package application

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription/api"
	serviceService "github.com/tidepool-org/platform/service/service"
	"go.uber.org/fx"
)

type Application struct {
	*serviceService.Authenticated
	router   *api.Router
}

type Params struct {
	fx.In

	Authenticated *serviceService.Authenticated
	Router        *api.Router

	Lifecycle fx.Lifecycle
}

func NewApplication(p Params) *Application {
	return &Application{
		Authenticated: p.Authenticated,
		router:        p.Router,
	}
}

func (a *Application) Initialize(provider application.Provider) error {
	if err := a.Authenticated.Initialize(provider); err != nil {
		return err
	}
	return a.initializeRouter()
}

func (a *Application) initializeRouter() error {
	a.Logger().Debug("Initializing router")
	if err := a.API().InitializeRouters(a.router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}
