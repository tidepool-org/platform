package v1

import (
	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/errors"
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
	return append(append(r.OAuthRoutes(), r.ProviderSessionsRoutes()...), r.RestrictedTokensRoutes()...)
}
