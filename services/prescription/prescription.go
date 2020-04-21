package main

import (
	"context"
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth/store/mongo"
	"github.com/tidepool-org/platform/prescription/api"
	prescription "github.com/tidepool-org/platform/prescription/application"
	"github.com/tidepool-org/platform/prescription/service"
	tidepoolService "github.com/tidepool-org/platform/service/service"
	"github.com/tidepool-org/platform/status"
	user "github.com/tidepool-org/platform/user/client"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		application.ProviderModule,
		tidepoolService.AuthenticatedModule,
		user.ClientModule,
		status.ReporterModule,
		fx.Provide(
			mongo.NewStore,
			service.NewService,
			api.NewRouter,
			prescription.NewApplication,
		),
		fx.Invoke(RunApplication),
	).Run()
}

func RunApplication(app *prescription.Application, provider application.Provider, lifecycle fx.Lifecycle) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				if err := app.Initialize(provider); err != nil {
					return err
				}

				go app.Run()
				return nil
			},
			OnStop: func(context.Context) error {
				app.Terminate()
				return nil
			},
		},
	)
}

