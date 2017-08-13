package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/service"
)

type Context struct {
	*Responder
	authDetails auth.Details
}

func New(response rest.ResponseWriter, request *rest.Request) (*Context, error) {
	rspdr, err := NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Context{
		Responder: rspdr,
	}, nil
}

func (c *Context) AuthDetails() auth.Details {
	if c.authDetails == nil {
		c.authDetails = service.GetRequestAuthDetails(c.Request())
	}

	return c.authDetails
}
