package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/metric"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	serviceContext "github.com/tidepool-org/platform/service/context"
	sessionStore "github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/user"
	userService "github.com/tidepool-org/platform/user/service"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Standard struct {
	*serviceContext.Responder
	authClient           auth.Client
	dataClient           dataClient.Client
	metricClient         metric.Client
	userClient           user.Client
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
}

func WithContext(authClient auth.Client, dataClient dataClient.Client, metricClient metric.Client, userClient user.Client,
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

func NewStandard(response rest.ResponseWriter, request *rest.Request, authClient auth.Client, dataClient dataClient.Client, metricClient metric.Client, userClient user.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store, profileStore profileStore.Store,
	sessionStore sessionStore.Store, userStore userStore.Store) (*Standard, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("user client is missing")
	}
	if confirmationStore == nil {
		return nil, errors.New("confirmation store is missing")
	}
	if messageStore == nil {
		return nil, errors.New("message store is missing")
	}
	if permissionStore == nil {
		return nil, errors.New("permission store is missing")
	}
	if profileStore == nil {
		return nil, errors.New("profile store is missing")
	}
	if sessionStore == nil {
		return nil, errors.New("session store is missing")
	}
	if userStore == nil {
		return nil, errors.New("user store is missing")
	}

	responder, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder:         responder,
		authClient:        authClient,
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

func (s *Standard) AuthClient() auth.Client {
	return s.authClient
}

func (s *Standard) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Standard) MetricClient() metric.Client {
	return s.metricClient
}

func (s *Standard) UserClient() user.Client {
	return s.userClient
}

func (s *Standard) ConfirmationsSession() confirmationStore.ConfirmationsSession {
	if s.confirmationsSession == nil {
		s.confirmationsSession = s.confirmationStore.NewConfirmationsSession()
	}
	return s.confirmationsSession
}

func (s *Standard) MessagesSession() messageStore.MessagesSession {
	if s.messagesSession == nil {
		s.messagesSession = s.messageStore.NewMessagesSession()
	}
	return s.messagesSession
}

func (s *Standard) PermissionsSession() permissionStore.PermissionsSession {
	if s.permissionsSession == nil {
		s.permissionsSession = s.permissionStore.NewPermissionsSession()
	}
	return s.permissionsSession
}

func (s *Standard) ProfilesSession() profileStore.ProfilesSession {
	if s.profilesSession == nil {
		s.profilesSession = s.profileStore.NewProfilesSession()
	}
	return s.profilesSession
}

func (s *Standard) SessionsSession() sessionStore.SessionsSession {
	if s.sessionsSession == nil {
		s.sessionsSession = s.sessionStore.NewSessionsSession()
	}
	return s.sessionsSession
}

func (s *Standard) UsersSession() userStore.UsersSession {
	if s.usersSession == nil {
		s.usersSession = s.userStore.NewUsersSession()
	}
	return s.usersSession
}
