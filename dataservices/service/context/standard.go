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
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/service"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	taskStore "github.com/tidepool-org/platform/task/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	commonService.Context
	metricServicesClient    metricservicesClient.Client
	userServicesClient      userservicesClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	authenticationDetails   userservicesClient.AuthenticationDetails
	dataStoreSession        dataStore.Session
	taskStoreSession        taskStore.Session
}

func WithContext(metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, taskStore taskStore.Store, handler service.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		dataStoreSession, err := dataStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new data store session for request", err)
			return
		}
		defer dataStoreSession.Close()

		taskStoreSession, err := taskStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new task store session for request", err)
			return
		}
		defer taskStoreSession.Close()

		handler(&Standard{
			Context:                 context,
			metricServicesClient:    metricServicesClient,
			userServicesClient:      userServicesClient,
			dataFactory:             dataFactory,
			dataDeduplicatorFactory: dataDeduplicatorFactory,
			dataStoreSession:        dataStoreSession,
			taskStoreSession:        taskStoreSession,
		})
	}
}

func (s *Standard) MetricServicesClient() metricservicesClient.Client {
	return s.metricServicesClient
}

func (s *Standard) UserServicesClient() userservicesClient.Client {
	return s.userServicesClient
}

func (s *Standard) DataFactory() data.Factory {
	return s.dataFactory
}

func (s *Standard) DataDeduplicatorFactory() deduplicator.Factory {
	return s.dataDeduplicatorFactory
}

func (s *Standard) DataStoreSession() dataStore.Session {
	return s.dataStoreSession
}

func (s *Standard) TaskStoreSession() taskStore.Session {
	return s.taskStoreSession
}

func (s *Standard) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	s.dataStoreSession.SetAgent(authenticationDetails)
}
