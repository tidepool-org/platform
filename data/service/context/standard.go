package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metric"
	serviceContext "github.com/tidepool-org/platform/service/context"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/user"
)

type Standard struct {
	*serviceContext.Responder
	authClient              auth.Client
	metricClient            metric.Client
	userClient              user.Client
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               dataStore.Store
	dataStoreDEPRECATED     dataStoreDEPRECATED.Store
	dataSession             dataStoreDEPRECATED.DataSession
	syncTaskStore           syncTaskStore.Store
	syncTasksSession        syncTaskStore.SyncTaskSession
	dataClient              dataClient.Client
}

func WithContext(authClient auth.Client, metricClient metric.Client, userClient user.Client,
	dataDeduplicatorFactory deduplicator.Factory, dataStore dataStore.Store,
	dataStoreDEPRECATED dataStoreDEPRECATED.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, metricClient, userClient,
			dataDeduplicatorFactory, dataStore, dataStoreDEPRECATED, syncTaskStore, dataClient)
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
	authClient auth.Client, metricClient metric.Client, userClient user.Client,
	dataDeduplicatorFactory deduplicator.Factory, dataStore dataStore.Store,
	dataStoreDEPRECATED dataStoreDEPRECATED.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client) (*Standard, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("user client is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("data deduplicator factory is missing")
	}
	if dataStore == nil {
		return nil, errors.New("data store is missing")
	}
	if dataStoreDEPRECATED == nil {
		return nil, errors.New("data store DEPRECATED is missing")
	}
	if syncTaskStore == nil {
		return nil, errors.New("sync task store is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}

	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}

	responder, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder:               responder,
		authClient:              authClient,
		metricClient:            metricClient,
		userClient:              userClient,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		dataStore:               dataStore,
		dataStoreDEPRECATED:     dataStoreDEPRECATED,
		syncTaskStore:           syncTaskStore,
		dataClient:              dataClient,
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

func (s *Standard) AuthClient() auth.Client {
	return s.authClient
}

func (s *Standard) MetricClient() metric.Client {
	return s.metricClient
}

func (s *Standard) UserClient() user.Client {
	return s.userClient
}

func (s *Standard) DataDeduplicatorFactory() deduplicator.Factory {
	return s.dataDeduplicatorFactory
}

func (s *Standard) DataSession() dataStoreDEPRECATED.DataSession {
	if s.dataSession == nil {
		s.dataSession = s.dataStoreDEPRECATED.NewDataSession()
	}
	return s.dataSession
}

func (s *Standard) SyncTaskSession() syncTaskStore.SyncTaskSession {
	if s.syncTasksSession == nil {
		s.syncTasksSession = s.syncTaskStore.NewSyncTaskSession()
	}
	return s.syncTasksSession
}

func (s *Standard) DataClient() dataClient.Client {
	return s.dataClient
}
