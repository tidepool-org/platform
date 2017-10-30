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
	*serviceContext.Responder
	notificationsSession store.NotificationsSession
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
		return nil, errors.New("service is missing")
	}

	rspdr, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Context{
		Service:   svc,
		Responder: rspdr,
	}, nil
}

func (c *Context) Close() {
	if c.notificationsSession != nil {
		c.notificationsSession.Close()
		c.notificationsSession = nil
	}
}

func (c *Context) NotificationsSession() store.NotificationsSession {
	if c.notificationsSession == nil {
		c.notificationsSession = c.NotificationStore().NewNotificationsSession()
	}
	return c.notificationsSession
}
