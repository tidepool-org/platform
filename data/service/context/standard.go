package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Standard struct {
	service.Context
	metricClient            metricClient.Client
	userClient              userClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	authenticationDetails   userClient.AuthenticationDetails
	dataStore               dataStore.Store
	dataStoreSession        dataStore.Session
	syncTaskStore           syncTaskStore.Store
	syncTaskStoreSession    syncTaskStore.Session
}

func WithContext(metricClient metricClient.Client, userClient userClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, syncTaskStore syncTaskStore.Store, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		standard := &Standard{
			Context:                 context,
			metricClient:            metricClient,
			userClient:              userClient,
			dataFactory:             dataFactory,
			dataDeduplicatorFactory: dataDeduplicatorFactory,
			dataStore:               dataStore,
			syncTaskStore:           syncTaskStore,
		}

		defer func() {
			if standard.syncTaskStoreSession != nil {
				standard.syncTaskStoreSession.Close()
			}
			if standard.dataStoreSession != nil {
				standard.dataStoreSession.Close()
			}
		}()

		handler(standard)
	}
}

func (s *Standard) MetricClient() metricClient.Client {
	return s.metricClient
}

func (s *Standard) UserClient() userClient.Client {
	return s.userClient
}

func (s *Standard) DataFactory() data.Factory {
	return s.dataFactory
}

func (s *Standard) DataDeduplicatorFactory() deduplicator.Factory {
	return s.dataDeduplicatorFactory
}

func (s *Standard) DataStoreSession() dataStore.Session {
	if s.dataStoreSession == nil {
		s.dataStoreSession = s.dataStore.NewSession(s.Context.Logger())
		s.dataStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.dataStoreSession
}

func (s *Standard) SyncTaskStoreSession() syncTaskStore.Session {
	if s.syncTaskStoreSession == nil {
		s.syncTaskStoreSession = s.syncTaskStore.NewSession(s.Context.Logger())
		s.syncTaskStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.syncTaskStoreSession
}

func (s *Standard) AuthenticationDetails() userClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	if s.dataStoreSession != nil {
		s.dataStoreSession.SetAgent(authenticationDetails)
	}
	if s.syncTaskStoreSession != nil {
		s.syncTaskStoreSession.SetAgent(authenticationDetails)
	}
}
