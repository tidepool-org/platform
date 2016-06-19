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

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	service.Context
	dataFactory             data.Factory
	dataStoreSession        store.Session
	dataDeduplicatorFactory deduplicator.Factory
	userServicesClient      client.Client
	requestUserID           string
}

func WithContext(dataFactory data.Factory, dataStore store.Store, dataDeduplicatorFactory deduplicator.Factory, userServicesClient client.Client, handler server.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context := service.NewStandard(response, request)

		dataStoreSession, err := dataStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new data store session for request", err)
			return
		}
		defer dataStoreSession.Close()

		handler(&Standard{
			Context:                 context,
			dataFactory:             dataFactory,
			dataStoreSession:        dataStoreSession,
			dataDeduplicatorFactory: dataDeduplicatorFactory,
			userServicesClient:      userServicesClient,
		})
	}
}

func (s *Standard) DataFactory() data.Factory {
	return s.dataFactory
}

func (s *Standard) DataStoreSession() store.Session {
	return s.dataStoreSession
}

func (s *Standard) DataDeduplicatorFactory() deduplicator.Factory {
	return s.dataDeduplicatorFactory
}

func (s *Standard) UserServicesClient() client.Client {
	return s.userServicesClient
}

func (s *Standard) RequestUserID() string {
	return s.requestUserID
}

func (s *Standard) SetRequestUserID(requestUserID string) {
	s.requestUserID = requestUserID
}
