package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	serviceContext "github.com/tidepool-org/platform/service/context"
)

type Context struct {
	*serviceContext.Context
	authClient auth.Client
}

func New(response rest.ResponseWriter, request *rest.Request, authClient auth.Client) (*Context, error) {
	if authClient == nil {
		return nil, errors.New("context", "auth client is missing")
	}

	ctx, err := serviceContext.New(response, request)
	if err != nil {
		return nil, err
	}

	return &Context{
		Context:    ctx,
		authClient: authClient,
	}, nil
}

func (c *Context) AuthClient() auth.Client {
	return c.authClient
}
