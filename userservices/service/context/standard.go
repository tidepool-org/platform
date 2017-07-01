package context

import (
	"github.com/ant0ine/go-json-rest/rest"

	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	notificationStore "github.com/tidepool-org/platform/notification/store"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userStore "github.com/tidepool-org/platform/user/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
)

type Standard struct {
	commonService.Context
	metricServicesClient     metricservicesClient.Client
	userServicesClient       userservicesClient.Client
	dataServicesClient       dataservicesClient.Client
	messageStore             messageStore.Store
	messageStoreSession      messageStore.Session
	notificationStore        notificationStore.Store
	notificationStoreSession notificationStore.Session
	permissionStore          permissionStore.Store
	permissionStoreSession   permissionStore.Session
	profileStore             profileStore.Store
	profileStoreSession      profileStore.Session
	sessionStore             sessionStore.Store
	sessionStoreSession      sessionStore.Session
	userStore                userStore.Store
	userStoreSession         userStore.Session
	authenticationDetails    userservicesClient.AuthenticationDetails
}

func WithContext(metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client, dataServicesClient dataservicesClient.Client,
	messageStore messageStore.Store, notificationStore notificationStore.Store, permissionStore permissionStore.Store, profileStore profileStore.Store,
	sessionStore sessionStore.Store, userStore userStore.Store, handler service.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		standard := &Standard{
			Context:              context,
			metricServicesClient: metricServicesClient,
			userServicesClient:   userServicesClient,
			dataServicesClient:   dataServicesClient,
			messageStore:         messageStore,
			notificationStore:    notificationStore,
			permissionStore:      permissionStore,
			profileStore:         profileStore,
			sessionStore:         sessionStore,
			userStore:            userStore,
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
			if standard.notificationStoreSession != nil {
				standard.notificationStoreSession.Close()
			}
			if standard.messageStoreSession != nil {
				standard.messageStoreSession.Close()
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

func (s *Standard) DataServicesClient() dataservicesClient.Client {
	return s.dataServicesClient
}

func (s *Standard) MessageStoreSession() messageStore.Session {
	if s.messageStoreSession == nil {
		s.messageStoreSession = s.messageStore.NewSession(s.Context.Logger())
		s.messageStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.messageStoreSession
}

func (s *Standard) NotificationStoreSession() notificationStore.Session {
	if s.notificationStoreSession == nil {
		s.notificationStoreSession = s.notificationStore.NewSession(s.Context.Logger())
		s.notificationStoreSession.SetAgent(s.authenticationDetails)
	}
	return s.notificationStoreSession
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

func (s *Standard) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	if s.messageStoreSession != nil {
		s.messageStoreSession.SetAgent(authenticationDetails)
	}
	if s.notificationStoreSession != nil {
		s.notificationStoreSession.SetAgent(authenticationDetails)
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
