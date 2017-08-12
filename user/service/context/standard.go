package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	authContext "github.com/tidepool-org/platform/auth/context"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	serviceContext "github.com/tidepool-org/platform/service/context"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userClient "github.com/tidepool-org/platform/user/client"
	userService "github.com/tidepool-org/platform/user/service"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Standard struct {
	*authContext.Context
	authClient               auth.Client
	dataClient               dataClient.Client
	metricClient             metricClient.Client
	userClient               userClient.Client
	confirmationStore        confirmationStore.Store
	confirmationStoreSession confirmationStore.Session
	messageStore             messageStore.Store
	messageStoreSession      messageStore.Session
	permissionStore          permissionStore.Store
	permissionStoreSession   permissionStore.Session
	profileStore             profileStore.Store
	profileStoreSession      profileStore.Session
	sessionStore             sessionStore.Store
	sessionStoreSession      sessionStore.Session
	userStore                userStore.Store
	userStoreSession         userStore.Session
	authDetails              auth.Details
}

func WithContext(authClient auth.Client, dataClient dataClient.Client, metricClient metricClient.Client, userClient userClient.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store, profileStore profileStore.Store,
	sessionStore sessionStore.Store, userStore userStore.Store, handler userService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, dataClient, metricClient, userClient,
			confirmationStore, messageStore, permissionStore, profileStore, sessionStore, userStore)
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

func NewStandard(response rest.ResponseWriter, request *rest.Request, authClient auth.Client, dataClient dataClient.Client, metricClient metricClient.Client, userClient userClient.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store, profileStore profileStore.Store,
	sessionStore sessionStore.Store, userStore userStore.Store) (*Standard, error) {
	if dataClient == nil {
		return nil, errors.New("context", "data client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("context", "metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("context", "user client is missing")
	}
	if confirmationStore == nil {
		return nil, errors.New("context", "confirmation store is missing")
	}
	if messageStore == nil {
		return nil, errors.New("context", "message store is missing")
	}
	if permissionStore == nil {
		return nil, errors.New("context", "permission store is missing")
	}
	if profileStore == nil {
		return nil, errors.New("context", "profile store is missing")
	}
	if sessionStore == nil {
		return nil, errors.New("context", "session store is missing")
	}
	if userStore == nil {
		return nil, errors.New("context", "user store is missing")
	}

	context, err := authContext.New(response, request, authClient)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Context:           context,
		dataClient:        dataClient,
		metricClient:      metricClient,
		userClient:        userClient,
		confirmationStore: confirmationStore,
		messageStore:      messageStore,
		permissionStore:   permissionStore,
		profileStore:      profileStore,
		sessionStore:      sessionStore,
		userStore:         userStore,
	}, nil
}

func (s *Standard) Close() {
	if s.userStoreSession != nil {
		s.userStoreSession.Close()
		s.userStoreSession = nil
	}
	if s.sessionStoreSession != nil {
		s.sessionStoreSession.Close()
		s.sessionStoreSession = nil
	}
	if s.profileStoreSession != nil {
		s.profileStoreSession.Close()
		s.profileStoreSession = nil
	}
	if s.permissionStoreSession != nil {
		s.permissionStoreSession.Close()
		s.permissionStoreSession = nil
	}
	if s.messageStoreSession != nil {
		s.messageStoreSession.Close()
		s.messageStoreSession = nil
	}
	if s.confirmationStoreSession != nil {
		s.confirmationStoreSession.Close()
		s.confirmationStoreSession = nil
	}
}

func (s *Standard) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Standard) MetricClient() metricClient.Client {
	return s.metricClient
}

func (s *Standard) UserClient() userClient.Client {
	return s.userClient
}

func (s *Standard) ConfirmationStoreSession() confirmationStore.Session {
	if s.confirmationStoreSession == nil {
		s.confirmationStoreSession = s.confirmationStore.NewSession(s.Context.Logger())
		s.confirmationStoreSession.SetAgent(s.AuthDetails())
	}
	return s.confirmationStoreSession
}

func (s *Standard) MessageStoreSession() messageStore.Session {
	if s.messageStoreSession == nil {
		s.messageStoreSession = s.messageStore.NewSession(s.Context.Logger())
		s.messageStoreSession.SetAgent(s.AuthDetails())
	}
	return s.messageStoreSession
}

func (s *Standard) PermissionStoreSession() permissionStore.Session {
	if s.permissionStoreSession == nil {
		s.permissionStoreSession = s.permissionStore.NewSession(s.Context.Logger())
		s.permissionStoreSession.SetAgent(s.AuthDetails())
	}
	return s.permissionStoreSession
}

func (s *Standard) ProfileStoreSession() profileStore.Session {
	if s.profileStoreSession == nil {
		s.profileStoreSession = s.profileStore.NewSession(s.Context.Logger())
		s.profileStoreSession.SetAgent(s.AuthDetails())
	}
	return s.profileStoreSession
}

func (s *Standard) SessionStoreSession() sessionStore.Session {
	if s.sessionStoreSession == nil {
		s.sessionStoreSession = s.sessionStore.NewSession(s.Context.Logger())
		s.sessionStoreSession.SetAgent(s.AuthDetails())
	}
	return s.sessionStoreSession
}

func (s *Standard) UserStoreSession() userStore.Session {
	if s.userStoreSession == nil {
		s.userStoreSession = s.userStore.NewSession(s.Context.Logger())
		s.userStoreSession.SetAgent(s.AuthDetails())
	}
	return s.userStoreSession
}
