package context

import (
	"github.com/mdblp/go-json-rest/rest"
)

type Context struct {
	*Responder
}

func New(res rest.ResponseWriter, req *rest.Request) (*Context, error) {
	rspdr, err := NewResponder(res, req)
	if err != nil {
		return nil, err
	}

	return &Context{
		Responder: rspdr,
	}, nil
}
