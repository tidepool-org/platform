package dataservices

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
)

type Context struct {
	*service.Context
	storeSession store.Session
}

type HandlerFunc func(context *Context)

func WithContext(dataStore store.Store, handler HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context := service.NewContext(response, request)

		storeSession, err := dataStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithServerFailure("Unable to create new data store session for request", err)
			return
		}
		defer storeSession.Close()

		handler(&Context{
			Context:      context,
			storeSession: storeSession,
		})
	}
}

func (c *Context) Store() store.Session {
	return c.storeSession
}
