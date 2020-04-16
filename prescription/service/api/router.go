package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/service/api"
)

type Router struct {
	service.Service
}

func NewRouter(svc service.Service) (*Router, error) {
	if svc == nil {
		return nil, errors.New("service is missing")
	}

	return &Router{
		Service: svc,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/status", r.StatusGet),
		rest.Post("/v1/prescriptions", api.Require(r.CreatePrescription)),
		rest.Get("/v1/prescriptions", api.Require(r.ListPrescriptions)),
		rest.Post("/v1/prescriptions/claim", api.Require(r.ClaimPrescription)),
		rest.Get("/v1/prescriptions/:prescriptionId", api.Require(r.GetPrescription)),
		rest.Patch("/v1/prescriptions/:prescriptionId", api.Require(r.UpdateState)),
		rest.Delete("/v1/prescriptions/:prescriptionId", api.Require(r.DeletePrescription)),
		rest.Post("/v1/prescriptions/:prescriptionId/revisions", api.Require(r.AddRevision)),
	}
}
