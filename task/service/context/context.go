package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	serviceContext "github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/task/service"
	"github.com/tidepool-org/platform/task/store"
)

type Context struct {
	service.Service
	*serviceContext.Context
	tasksSession store.TasksSession
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
	if c.tasksSession != nil {
		c.tasksSession.Close()
		c.tasksSession = nil
	}
}

func (c *Context) TasksSession() store.TasksSession {
	if c.tasksSession == nil {
		c.tasksSession = c.TaskStore().NewTasksSession(c.Logger())
		c.tasksSession.SetAgent(c.AuthDetails())
	}
	return c.tasksSession
}
