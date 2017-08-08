package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/service"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	commonService.Context
	metricServicesClient    metricservicesClient.Client
	userServicesClient      userservicesClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	authenticationDetails   userservicesClient.AuthenticationDetails
	dataStore               dataStore.Store
	dataStoreSession        dataStore.Session
	syncTaskStore           syncTaskStore.Store
	syncTaskStoreSession    syncTaskStore.Session
}

func WithContext(metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, syncTaskStore syncTaskStore.Store, handler service.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		standard := &Standard{
			Context:                 context,
			metricServicesClient:    metricServicesClient,
			userServicesClient:      userServicesClient,
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

func (s *Standard) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	if s.dataStoreSession != nil {
		s.dataStoreSession.SetAgent(authenticationDetails)
	}
	if s.syncTaskStoreSession != nil {
		s.syncTaskStoreSession.SetAgent(authenticationDetails)
	}
}
