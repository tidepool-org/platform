package application

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/devices"

	structuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	"github.com/tidepool-org/platform/prescription/api"
	"github.com/tidepool-org/platform/prescription/service"
	prescriptionMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/status"
	user "github.com/tidepool-org/platform/user/client"
)

var Prescription = fx.Options(
	devices.ClientModule,
	user.ClientModule,
	structuredMongo.StoreModule,
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
