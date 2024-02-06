package context

import (
	"net/http"

	dataStore "github.com/tidepool-org/platform/data/store"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	serviceContext "github.com/tidepool-org/platform/service/context"
)

type Standard struct {
	*serviceContext.Responder
	authClient       auth.Client
	permissionClient permission.Client
	dataStore        dataStore.Store
	dataRepository   dataStore.DataRepository
	isUploadIdUsed   bool
}

func WithContext(authClient auth.Client, permissionClient permission.Client,
	store dataStore.Store, isUploadIdUsed bool, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, permissionClient, store, isUploadIdUsed)
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
	store dataStore.Store, isUploadIdUsed bool) (*Standard, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if permissionClient == nil {
		return nil, errors.New("permission client is missing")
	}
	if store == nil {
		return nil, errors.New("data store DEPRECATED is missing")
	}

	responder, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder:        responder,
		authClient:       authClient,
		permissionClient: permissionClient,
		dataStore:        store,
		isUploadIdUsed:   isUploadIdUsed,
	}, nil
}

func (s *Standard) Close() {
	if s.dataRepository != nil {
		s.dataRepository = nil
	}
}

func (s *Standard) AuthClient() auth.Client {
	return s.authClient
}

func (s *Standard) PermissionClient() permission.Client {
	return s.permissionClient
}

func (s *Standard) DataRepository() dataStore.DataRepository {
	if s.dataRepository == nil {
		s.dataRepository = s.dataStore.NewDataRepository()
	}
	return s.dataRepository
}

func (s *Standard) IsUploadIdUsed() bool {
	return s.isUploadIdUsed
}
