package api

import (
	"context"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type StatusProvider interface {
	Status(context.Context) interface{}
}

type StatusRouter struct {
	StatusProvider
}

func NewStatusRouter(statusProvider StatusProvider) (*StatusRouter, error) {
	if statusProvider == nil {
		return nil, errors.New("status provider is missing")
	}

	return &StatusRouter{
		StatusProvider: statusProvider,
	}, nil
}

func (s *StatusRouter) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/status", s.StatusGet),
	}
}

func (s *StatusRouter) StatusGet(res rest.ResponseWriter, req *rest.Request) {
	request.MustNewResponder(res, req).Data(http.StatusOK, s.Status(req.Context()))
}
