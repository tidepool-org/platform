package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification/service"
	"github.com/tidepool-org/platform/notification/store"
	serviceContext "github.com/tidepool-org/platform/service/context"
)

type Context struct {
	service.Service
	*serviceContext.Context
	notificationStoreSession store.StoreSession
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
	if c.notificationStoreSession != nil {
		c.notificationStoreSession.Close()
		c.notificationStoreSession = nil
	}
}

func (c *Context) NotificationStoreSession() store.StoreSession {
	if c.notificationStoreSession == nil {
		c.notificationStoreSession = c.NotificationStore().NewSession(c.Logger())
		c.notificationStoreSession.SetAgent(c.AuthDetails())
	}
	return c.notificationStoreSession
}
