package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	authContext "github.com/tidepool-org/platform/auth/context"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	metricClient "github.com/tidepool-org/platform/metric/client"
	serviceContext "github.com/tidepool-org/platform/service/context"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Standard struct {
	*authContext.Context
	metricClient            metricClient.Client
	userClient              userClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               dataStore.Store
	dataSession             dataStore.DataSession
	syncTaskStore           syncTaskStore.Store
	syncTasksSession        syncTaskStore.SyncTasksSession
}

func WithContext(authClient auth.Client, metricClient metricClient.Client, userClient userClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, syncTaskStore syncTaskStore.Store, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, metricClient, userClient,
			dataFactory, dataDeduplicatorFactory, dataStore, syncTaskStore)
		if standardErr != nil {
			if responder, responderErr := serviceContext.NewResponder(response, request); responderErr != nil {
				response.WriteHeader(http.StatusInternalServerError)
			} else {
				responder.RespondWithInternalServerFailure("Unable to create new context for request", standardErr)
			}
			return
		}
		defer standard.Close()

		handler(standard)
	}
}

func NewStandard(response rest.ResponseWriter, request *rest.Request,
	authClient auth.Client, metricClient metricClient.Client, userClient userClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, syncTaskStore syncTaskStore.Store) (*Standard, error) {
	if metricClient == nil {
		return nil, errors.New("context", "metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("context", "user client is missing")
	}
	if dataFactory == nil {
		return nil, errors.New("context", "data factory is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("context", "data deduplicator factory is missing")
	}
	if dataStore == nil {
		return nil, errors.New("context", "data store is missing")
	}
	if syncTaskStore == nil {
		return nil, errors.New("context", "sync task store is missing")
	}

	context, err := authContext.New(response, request, authClient)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Context:                 context,
		metricClient:            metricClient,
		userClient:              userClient,
		dataFactory:             dataFactory,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		dataStore:               dataStore,
		syncTaskStore:           syncTaskStore,
	}, nil
}

func (s *Standard) Close() {
	if s.syncTasksSession != nil {
		s.syncTasksSession.Close()
		s.syncTasksSession = nil
	}
	if s.dataSession != nil {
		s.dataSession.Close()
		s.dataSession = nil
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

func (s *Standard) DataSession() dataStore.DataSession {
	if s.dataSession == nil {
		s.dataSession = s.dataStore.NewDataSession(s.Logger())
		s.dataSession.SetAgent(s.AuthDetails())
	}
	return s.dataSession
}

func (s *Standard) SyncTasksSession() syncTaskStore.SyncTasksSession {
	if s.syncTasksSession == nil {
		s.syncTasksSession = s.syncTaskStore.NewSyncTasksSession(s.Logger())
		s.syncTasksSession.SetAgent(s.AuthDetails())
	}
	return s.syncTasksSession
}
