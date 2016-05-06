package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
)

type Context struct {
	*service.Context
	store         store.Session
	client        client.Client
	RequestUserID string
}

type HandlerFunc func(context *Context)

func WithContext(dataStore store.Store, client client.Client, handler HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context := service.NewContext(response, request)

		store, err := dataStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithServerFailure("Unable to create new data store session for request", err)
			return
		}
		defer store.Close()

		handler(&Context{
			Context: context,
			store:   store,
			client:  client,
		})
	}
}

func (c *Context) Store() store.Session {
	return c.store
}

func (c *Context) Client() client.Client {
	return c.client
}
