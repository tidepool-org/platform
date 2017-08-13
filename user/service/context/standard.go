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
	authClient           auth.Client
	dataClient           dataClient.Client
	metricClient         metricClient.Client
	userClient           userClient.Client
	confirmationStore    confirmationStore.Store
	confirmationsSession confirmationStore.ConfirmationsSession
	messageStore         messageStore.Store
	messagesSession      messageStore.MessagesSession
	permissionStore      permissionStore.Store
	permissionsSession   permissionStore.PermissionsSession
	profileStore         profileStore.Store
	profilesSession      profileStore.ProfilesSession
	sessionStore         sessionStore.Store
	sessionsSession      sessionStore.SessionsSession
	userStore            userStore.Store
	usersSession         userStore.UsersSession
	authDetails          auth.Details
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
	if s.usersSession != nil {
		s.usersSession.Close()
		s.usersSession = nil
	}
	if s.sessionsSession != nil {
		s.sessionsSession.Close()
		s.sessionsSession = nil
	}
	if s.profilesSession != nil {
		s.profilesSession.Close()
		s.profilesSession = nil
	}
	if s.permissionsSession != nil {
		s.permissionsSession.Close()
		s.permissionsSession = nil
	}
	if s.messagesSession != nil {
		s.messagesSession.Close()
		s.messagesSession = nil
	}
	if s.confirmationsSession != nil {
		s.confirmationsSession.Close()
		s.confirmationsSession = nil
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

func (s *Standard) ConfirmationsSession() confirmationStore.ConfirmationsSession {
	if s.confirmationsSession == nil {
		s.confirmationsSession = s.confirmationStore.NewConfirmationsSession(s.Context.Logger())
		s.confirmationsSession.SetAgent(s.AuthDetails())
	}
	return s.confirmationsSession
}

func (s *Standard) MessagesSession() messageStore.MessagesSession {
	if s.messagesSession == nil {
		s.messagesSession = s.messageStore.NewMessagesSession(s.Context.Logger())
		s.messagesSession.SetAgent(s.AuthDetails())
	}
	return s.messagesSession
}

func (s *Standard) PermissionsSession() permissionStore.PermissionsSession {
	if s.permissionsSession == nil {
		s.permissionsSession = s.permissionStore.NewPermissionsSession(s.Context.Logger())
		s.permissionsSession.SetAgent(s.AuthDetails())
	}
	return s.permissionsSession
}

func (s *Standard) ProfilesSession() profileStore.ProfilesSession {
	if s.profilesSession == nil {
		s.profilesSession = s.profileStore.NewProfilesSession(s.Context.Logger())
		s.profilesSession.SetAgent(s.AuthDetails())
	}
	return s.profilesSession
}

func (s *Standard) SessionsSession() sessionStore.SessionsSession {
	if s.sessionsSession == nil {
		s.sessionsSession = s.sessionStore.NewSessionsSession(s.Context.Logger())
		s.sessionsSession.SetAgent(s.AuthDetails())
	}
	return s.sessionsSession
}

func (s *Standard) UsersSession() userStore.UsersSession {
	if s.usersSession == nil {
		s.usersSession = s.userStore.NewUsersSession(s.Context.Logger())
		s.usersSession.SetAgent(s.AuthDetails())
	}
	return s.usersSession
}
