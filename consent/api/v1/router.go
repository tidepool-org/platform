package v1

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/consent"
	serviceApi "github.com/tidepool-org/platform/service/api"
)

type Router struct {
	service consent.Service
}

func NewRouter(service consent.Service) (*Router, error) {
	return &Router{service: service}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/consents", serviceApi.RequireUser(r.ListConsents)),
		rest.Get("/v1/consents/:type", serviceApi.RequireUser(r.GetConsentByType)),
		rest.Get("/v1/consents/:type/versions", serviceApi.RequireUser(r.GetConsentVersions)),

		rest.Get("/v1/users/:userId/consents", serviceApi.RequireUser(r.ListConsentRecords)),
		rest.Post("/v1/users/:userId/consents", serviceApi.RequireUser(r.CreateConsentRecord)),
		rest.Get("/v1/users/:userId/consents/:id", serviceApi.RequireUser(r.GetConsentRecord)),
		rest.Patch("/v1/users/:userId/consents/:id", serviceApi.RequireUser(r.UpdateConsentRecord)),
		rest.Delete("/v1/users/:userId/consents/:id", serviceApi.RequireUser(r.UpdateConsentRecord)),
	}
}
