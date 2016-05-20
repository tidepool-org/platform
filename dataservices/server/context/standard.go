package context

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

	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	service.Context
	store         store.Session
	client        client.Client
	requestUserID string
}

func WithContext(dataStore store.Store, client client.Client, handler server.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context := context.NewStandard(response, request)

		store, err := dataStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new data store session for request", err)
			return
		}
		defer store.Close()

		handler(&Standard{
			Context: context,
			store:   store,
			client:  client,
		})
	}
}

func (s *Standard) Store() store.Session {
	return s.store
}

func (s *Standard) Client() client.Client {
	return s.client
}

func (s *Standard) RequestUserID() string {
	return s.requestUserID
}

func (s *Standard) SetRequestUserID(requestUserID string) {
	s.requestUserID = requestUserID
}
