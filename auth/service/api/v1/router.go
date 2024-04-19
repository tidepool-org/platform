package v1

import (
	"github.com/ant0ine/go-json-rest/rest"

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
	routes := [][]*rest.Route{
		r.OAuthRoutes(),
		r.ProviderSessionsRoutes(),
		r.RestrictedTokensRoutes(),
		r.DeviceCheckRoutes(),
		r.DeviceTokensRoutes(),
		r.AppValidateRoutes(),
	}
	acc := make([]*rest.Route, 0)
	for _, r := range routes {
		acc = append(acc, r...)
	}
	return acc
}
