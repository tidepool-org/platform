package main

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/go-common/tracing"

	provider "github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/prescription/application"
	"github.com/tidepool-org/platform/service/service"
)

func main() {
	fx.New(
		tracing.TracingModule,
		provider.ProviderModule,
		application.Prescription,
		service.APIServiceModule,
		fx.Invoke(tracing.StartTracer, service.Start),
	).Run()
}
