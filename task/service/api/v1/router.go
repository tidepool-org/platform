package v1

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/task/service"
)

type Router struct {
	service.Service
}

func NewRouter(svc service.Service) (*Router, error) {
	if svc == nil {
		return nil, errors.New("v1", "service is missing")
	}

	return &Router{
		Service: svc,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{}
}
