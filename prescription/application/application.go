package application

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/auth/client"

	"github.com/tidepool-org/platform/clinics"

	"github.com/tidepool-org/platform/devices"

	structuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	"github.com/tidepool-org/platform/prescription/api"
	"github.com/tidepool-org/platform/prescription/service"
	prescriptionMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/status"
)

var Prescription = fx.Options(
	devices.ClientModule,
	structuredMongo.StoreModule,
	clinics.ClientModule,
	client.ProvideServiceName("prescription"),
	client.ExternalClientModule,
	fx.Provide(
		prescriptionMongo.NewStore,
		prescriptionMongo.NewStatusReporter,
		service.NewDeviceSettingsValidator,
		service.NewService,
		fx.Annotated{
			Group:  "routers",
			Target: api.NewRouter,
		},
	),
	status.RouterModule,
)
