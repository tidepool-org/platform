package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/prescription/service"

	"github.com/tidepool-org/platform/prescription"
	router "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/service/api"
)

type Router struct {
	deviceSettingsValidator service.DeviceSettingsValidator
	prescriptionService     prescription.Service
	userClient              user.Client
}

type Params struct {
	fx.In

	DeviceSettingsValidator service.DeviceSettingsValidator
	PrescriptionService     prescription.Service
	UserClient              user.Client
}

func NewRouter(p Params) router.Router {
	return &Router{
		deviceSettingsValidator: p.DeviceSettingsValidator,
		prescriptionService:     p.PrescriptionService,
		userClient:              p.UserClient,
	}
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/prescriptions", api.Require(r.CreatePrescription)),
		rest.Get("/v1/prescriptions", api.Require(r.ListCurrentUserPrescriptions)),
		rest.Get("/v1/users/:userId/prescriptions", api.Require(r.ListUserPrescriptions)),
		rest.Post("/v1/prescriptions/claim", api.Require(r.ClaimPrescription)),
		rest.Get("/v1/prescriptions/:prescriptionId", api.Require(r.GetPrescription)),
		rest.Patch("/v1/prescriptions/:prescriptionId", api.Require(r.UpdateState)),
		rest.Delete("/v1/prescriptions/:prescriptionId", api.Require(r.DeletePrescription)),
		rest.Post("/v1/prescriptions/:prescriptionId/revisions", api.Require(r.AddRevision)),
	}
}
