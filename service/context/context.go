package context

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type Context struct {
	*Responder
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
