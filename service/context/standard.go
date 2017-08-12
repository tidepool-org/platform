package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	*Responder
	authDetails auth.Details
}

func NewStandard(response rest.ResponseWriter, request *rest.Request) (*Standard, error) {
	responder, err := NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder: responder,
	}, nil
}

func (s *Standard) AuthDetails() auth.Details {
	if s.authDetails == nil {
		s.authDetails = service.GetRequestAuthDetails(s.Request())
	}

	return s.authDetails
}
