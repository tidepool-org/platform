package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	serviceContext "github.com/tidepool-org/platform/service/context"
)

type Context struct {
	service.Service
	*serviceContext.Context
	authStoreSession store.StoreSession
}

func MustNew(svc service.Service, response rest.ResponseWriter, request *rest.Request) *Context {
	ctx, err := New(svc, response, request)
	if err != nil {
		panic(err)
	}

	return ctx
}

func New(svc service.Service, response rest.ResponseWriter, request *rest.Request) (*Context, error) {
	if svc == nil {
		return nil, errors.New("context", "service is missing")
	}

	ctx, err := serviceContext.New(response, request)
	if err != nil {
		return nil, err
	}

	return &Context{
		Service: svc,
		Context: ctx,
	}, nil
}

func (c *Context) Close() {
	if c.authStoreSession != nil {
		c.authStoreSession.Close()
		c.authStoreSession = nil
	}
}

func (c *Context) AuthStoreSession() store.StoreSession {
	if c.authStoreSession == nil {
		c.authStoreSession = c.AuthStore().NewSession(c.Logger())
		c.authStoreSession.SetAgent(c.AuthDetails())
	}
	return c.authStoreSession
}
