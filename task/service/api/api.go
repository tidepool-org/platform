package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	taskService "github.com/tidepool-org/platform/task/service"
)

type Router struct {
	taskService.Service
}

func NewRouter(service taskService.Service) (*Router, error) {
	if service == nil {
		return nil, errors.New("service is missing")
	}

	return &Router{
		Service: service,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/status", r.StatusGet),
	}
}

func (r *Router) StatusGet(res rest.ResponseWriter, req *rest.Request) {
	request.MustNewResponder(res, req).Data(http.StatusOK, r.Status())
}
