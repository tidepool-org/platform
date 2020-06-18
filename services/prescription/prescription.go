package main

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/mailer"

	provider "github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/prescription/application"
	"github.com/tidepool-org/platform/service/service"
)

func main() {
	fx.New(
		provider.ProviderModule,
		mailer.Module,
		application.Prescription,
		service.APIServiceModule,
		fx.Invoke(service.Start),
	).Run()
}
