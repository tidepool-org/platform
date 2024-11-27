package service

import (
	"context"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

var APIServiceModule = fx.Provide(
	NewAuthenticated,
	NewAPIService,
)

type APIService struct {
	logger   log.Logger
	provider application.Provider
	routers  []service.Router
	svc      *Authenticated
}

type Params struct {
	fx.In

	Logger   log.Logger
	Provider application.Provider
	Routers  []service.Router `group:"routers"`
	Service  *Authenticated
}

// NewAPIService instantiates APIService in an 'fx' dependency injection application, and
// is meant to be used a stand-alone web micro-service which accepts an array of routers to
// handle incoming http requests.
//
// It should be provided as a dependency to an 'fx' application. It expects one or more
// routers to be present in the dependency graph and added in the "routers" value group.
//
// This is an example of an application with 3 routers - a health check router and two
// routers for handling 2 different API version requests.
//
//	var StatusRouterModule = fx.Provide(fx.Annotated{
//		   Group: "routers", // Params requires routers to be tagged with "routers" group
//		   Target: NewStatusRouter, // "NewStatusRouter(...) service.Router" is a constructor function
//	})
//
// fx.New(
//
//	  fx.Provide(DefaultProvider), // Required by StartApplication(provider application.Provider)
//	  fx.Provide(NewStoreStatusReporter), // Injected to StatusRouterModule
//	  StatusRouterModule,
//	  fx.Provide(fx.Annotated{
//		     Group:  "routers",
//		     Target: V1PrescriptionsRouter, // "V1PrescriptionsRouter(...) service.Router" is a constructor function
//	  }),
//	  fx.Provide(fx.Annotated{
//		     Group:  "routers",
//		     Target: V2PrescriptionsRouter, // "V1PrescriptionsRouter(...) service.Router" is a constructor function
//	  }),
//	  fx.Provide(NewAPIService),
//	  fx.Invoke(Start)
//
// ).Run()
func NewAPIService(p Params) (*APIService, error) {
	if len(p.Routers) == 0 {
		return nil, errors.New("application routers are missing")
	}

	return &APIService{
		logger:   p.Logger,
		provider: p.Provider,
		svc:      p.Service,
		routers:  p.Routers,
	}, nil
}

func (a *APIService) Initialize() error {
	if err := a.svc.Initialize(a.provider); err != nil {
		return err
	}
	return a.initializeRouters()
}

func (a *APIService) Run() error {
	return a.svc.Run()
}

func (a *APIService) Terminate() {
	a.svc.Terminate()
}

func (a *APIService) initializeRouters() error {
	a.logger.Debug("Initializing routers")
	if err := a.svc.API().InitializeRouters(a.routers...); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

type StartParams struct {
	fx.In

	App       *APIService
	Lifecycle fx.Lifecycle
}

func Start(p StartParams) {
	p.Lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				if err := p.App.Initialize(); err != nil {
					return err
				}

				go p.App.Run()
				return nil
			},
			OnStop: func(context.Context) error {
				p.App.Terminate()
				return nil
			},
		},
	)
}
