package combined

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/application"
	configEnv "github.com/tidepool-org/platform/config/env"
	prescriptionApplication "github.com/tidepool-org/platform/prescription/application"
	"github.com/tidepool-org/platform/service/server"
	serviceService "github.com/tidepool-org/platform/service/service"
)

// prescription adapts the Fx dependency graph to the common service lifecycle.
// Its Run errors are supervised along with those of the other components.
type prescriptionRunner struct {
	app         *fx.App
	api         *serviceService.APIService
	factory     server.Factory
	initialized bool
}

func (p *prescriptionRunner) SetServerFactory(factory server.Factory) { p.factory = factory }

func (p *prescriptionRunner) Initialize(provider application.Provider) error {
	p.app = fx.New(
		fx.NopLogger,
		fx.Provide(func() application.Provider { return provider }),
		fx.Provide(configEnv.NewDefaultReporter),
		application.ProviderComponentsModule,
		prescriptionApplication.Dependencies,
		prescriptionApplication.Prescription,
		fx.Provide(func() *serviceService.Authenticated {
			svc := serviceService.NewAuthenticated()
			svc.SetServerFactory(p.factory)
			return svc
		}),
		fx.Provide(serviceService.NewAPIService),
		fx.Populate(&p.api),
	)
	if err := p.app.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.app.StartTimeout())
	defer cancel()
	if err := p.app.Start(ctx); err != nil {
		return err
	}
	p.initialized = true
	return p.api.Initialize()
}

func (p *prescriptionRunner) Run() error { return p.api.Run() }

func (p *prescriptionRunner) Terminate() {
	if p.initialized {
		p.api.Terminate()
	}
	if p.app != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = p.app.Stop(ctx)
	}
}
