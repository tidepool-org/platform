package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	serviceContext "github.com/tidepool-org/platform/service/context"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
)

type Standard struct {
	*serviceContext.Responder
	authClient              auth.Client
	permissionClient        permission.Client
	dataDeduplicatorFactory deduplicator.Factory
	dataStoreDEPRECATED     dataStoreDEPRECATED.Store
	dataSession             dataStoreDEPRECATED.DataSession
	syncTaskStore           syncTaskStore.Store
	syncTasksSession        syncTaskStore.SyncTaskSession
	dataClient              dataClient.Client
}

func WithContext(authClient auth.Client, permissionClient permission.Client,
	dataDeduplicatorFactory deduplicator.Factory,
	dataStoreDEPRECATED dataStoreDEPRECATED.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, permissionClient,
			dataDeduplicatorFactory, dataStoreDEPRECATED, syncTaskStore, dataClient)
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
	authClient auth.Client, permissionClient permission.Client,
	dataDeduplicatorFactory deduplicator.Factory,
	dataStoreDEPRECATED dataStoreDEPRECATED.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client) (*Standard, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if permissionClient == nil {
		return nil, errors.New("permission client is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("data deduplicator factory is missing")
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

	responder, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder:               responder,
		authClient:              authClient,
		permissionClient:        permissionClient,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
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

func (s *Standard) PermissionClient() permission.Client {
	return s.permissionClient
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
