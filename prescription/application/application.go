package application

import (
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/store/structured/mongoofficial"

	"github.com/tidepool-org/platform/prescription/api"
	"github.com/tidepool-org/platform/prescription/service"
	prescriptionMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/status"
	user "github.com/tidepool-org/platform/user/client"
)

var Prescription = fx.Options(
	user.ClientModule,
	mongoofficial.StoreModule,
	fx.Provide(
		prescriptionMongo.NewStore,
		prescriptionMongo.NewStatusReporter,
		service.NewService,
		fx.Annotated{
			Group:  "routers",
			Target: api.NewRouter,
		},
	),
	status.RouterModule,
)
