package application

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/prescription/api"
	"github.com/tidepool-org/platform/prescription/service"
	prescriptionMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/status"
	user "github.com/tidepool-org/platform/user/client"
)

var Prescription = fx.Options(
	user.ClientModule,
	fx.Provide(
		prescriptionMongo.NewStore,
		prescriptionMongo.NewStoreStatusReporter,
		service.NewService,
		fx.Annotated{
			Group:  "routers",
			Target: api.NewRouter,
		},
	),
	status.RouterModule,
)
