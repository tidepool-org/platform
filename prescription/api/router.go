package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/clinics"

	"github.com/tidepool-org/platform/prescription/service"

	"github.com/tidepool-org/platform/prescription"
	router "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/service/api"
)

type Router struct {
	deviceSettingsValidator service.DeviceSettingsValidator
	prescriptionService     prescription.Service
	clinicsClient           clinics.Client
}

type Params struct {
	fx.In

	ClinicsClient           clinics.Client
	DeviceSettingsValidator service.DeviceSettingsValidator
	PrescriptionService     prescription.Service
}

func NewRouter(p Params) router.Router {
	return &Router{
		clinicsClient:           p.ClinicsClient,
		deviceSettingsValidator: p.DeviceSettingsValidator,
		prescriptionService:     p.PrescriptionService,
	}
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/clinics/:clinicId/prescriptions", api.RequireUser(r.ListClinicPrescriptions)),
		rest.Post("/v1/clinics/:clinicId/prescriptions", api.RequireUser(r.CreatePrescription)),
		rest.Get("/v1/clinics/:clinicId/prescriptions/:prescriptionId", api.RequireUser(r.GetClinicPrescription)),
		rest.Post("/v1/clinics/:clinicId/prescriptions/:prescriptionId/revisions", api.RequireUser(r.AddRevision)),
		rest.Delete("/v1/clinics/:clinicId/prescriptions/:prescriptionId", api.RequireUser(r.DeletePrescription)),

		rest.Post("/v1/patients/:userId/prescriptions", api.RequireUser(r.ClaimPrescription)),
		rest.Get("/v1/patients/:userId/prescriptions", api.Require(r.ListUserPrescriptions)),
		rest.Get("/v1/patients/:userId/prescriptions/:prescriptionId", api.Require(r.GetPatientPrescription)),
		rest.Patch("/v1/patients/:userId/prescriptions/:prescriptionId", api.RequireUser(r.UpdateState)),
	}
}
