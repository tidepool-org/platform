package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userClient "github.com/tidepool-org/platform/user/client"
	userService "github.com/tidepool-org/platform/user/service"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Standard struct {
	service.Context
	metricClient             metricClient.Client
	userClient               userClient.Client
	dataClient               dataClient.Client
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
	authenticationDetails    userClient.AuthenticationDetails
}

func WithContext(metricClient metricClient.Client, userClient userClient.Client, dataClient dataClient.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store, profileStore profileStore.Store,
	sessionStore sessionStore.Store, userStore userStore.Store, handler userService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		standard := &Standard{
			Context:           context,
			metricClient:      metricClient,
			userClient:        userClient,
			dataClient:        dataClient,
			confirmationStore: confirmationStore,
			messageStore:      messageStore,
			permissionStore:   permissionStore,
			profileStore:      profileStore,
			sessionStore:      sessionStore,
			userStore:         userStore,
		}

		defer func() {
			if standard.userStoreSession != nil {
				standard.userStoreSession.Close()
			}
			if standard.sessionStoreSession != nil {
				standard.sessionStoreSession.Close()
			}
			if standard.profileStoreSession != nil {
				standard.profileStoreSession.Close()
			}
			if standard.permissionStoreSession != nil {
				standard.permissionStoreSession.Close()
			}
			if standard.messageStoreSession != nil {
				standard.messageStoreSession.Close()
			}
			if standard.confirmationStoreSession != nil {
				standard.confirmationStoreSession.Close()
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

func (s *Standard) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Standard) ConfirmationStoreSession() confirmationStore.Session {
	if s.confirmationStoreSession == nil {
		s.confirmationStoreSession = s.confirmationStore.NewSession(s.Context.Logger())
		s.confirmationStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.confirmationStoreSession
}

func (s *Standard) MessageStoreSession() messageStore.Session {
	if s.messageStoreSession == nil {
		s.messageStoreSession = s.messageStore.NewSession(s.Context.Logger())
		s.messageStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.messageStoreSession
}

func (s *Standard) PermissionStoreSession() permissionStore.Session {
	if s.permissionStoreSession == nil {
		s.permissionStoreSession = s.permissionStore.NewSession(s.Context.Logger())
		s.permissionStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.permissionStoreSession
}

func (s *Standard) ProfileStoreSession() profileStore.Session {
	if s.profileStoreSession == nil {
		s.profileStoreSession = s.profileStore.NewSession(s.Context.Logger())
		s.profileStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.profileStoreSession
}

func (s *Standard) SessionStoreSession() sessionStore.Session {
	if s.sessionStoreSession == nil {
		s.sessionStoreSession = s.sessionStore.NewSession(s.Context.Logger())
		s.sessionStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.sessionStoreSession
}

func (s *Standard) UserStoreSession() userStore.Session {
	if s.userStoreSession == nil {
		s.userStoreSession = s.userStore.NewSession(s.Context.Logger())
		s.userStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.userStoreSession
}

func (s *Standard) AuthenticationDetails() userClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	if s.confirmationStoreSession != nil {
		s.confirmationStoreSession.SetAgent(authenticationDetails)
	}
	if s.messageStoreSession != nil {
		s.messageStoreSession.SetAgent(authenticationDetails)
	}
	if s.permissionStoreSession != nil {
		s.permissionStoreSession.SetAgent(authenticationDetails)
	}
	if s.profileStoreSession != nil {
		s.profileStoreSession.SetAgent(authenticationDetails)
	}
	if s.sessionStoreSession != nil {
		s.sessionStoreSession.SetAgent(authenticationDetails)
	}
	if s.userStoreSession != nil {
		s.userStoreSession.SetAgent(authenticationDetails)
	}
}
