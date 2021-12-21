package main

import (
	"go.uber.org/fx"

	provider "github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/prescription/application"
	"github.com/tidepool-org/platform/service/service"
)

func main() {
	fx.New(
		provider.ProviderModule,
		application.Dependencies,
		application.Prescription,
		service.APIServiceModule,
		fx.Invoke(service.Start),
	).Run()
}
