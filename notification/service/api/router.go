package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification/service"
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
	}
}
