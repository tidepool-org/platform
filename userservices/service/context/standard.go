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
	messageStoreSession      messageStore.Session
	notificationStoreSession notificationStore.Session
	permissionStoreSession   permissionStore.Session
	profileStoreSession      profileStore.Session
	sessionStoreSession      sessionStore.Session
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

		messageStoreSession, err := messageStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new message store session for request", err)
			return
		}
		defer messageStoreSession.Close()

		notificationStoreSession, err := notificationStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new notification store session for request", err)
			return
		}
		defer notificationStoreSession.Close()

		permissionStoreSession, err := permissionStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new permission store session for request", err)
			return
		}
		defer permissionStoreSession.Close()

		profileStoreSession, err := profileStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new profile store session for request", err)
			return
		}
		defer profileStoreSession.Close()

		sessionStoreSession, err := sessionStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new session store session for request", err)
			return
		}
		defer sessionStoreSession.Close()

		userStoreSession, err := userStore.NewSession(context.Logger())
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new user store session for request", err)
			return
		}
		defer userStoreSession.Close()

		handler(&Standard{
			Context:                  context,
			metricServicesClient:     metricServicesClient,
			userServicesClient:       userServicesClient,
			dataServicesClient:       dataServicesClient,
			messageStoreSession:      messageStoreSession,
			notificationStoreSession: notificationStoreSession,
			permissionStoreSession:   permissionStoreSession,
			profileStoreSession:      profileStoreSession,
			sessionStoreSession:      sessionStoreSession,
			userStoreSession:         userStoreSession,
		})
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
	return s.messageStoreSession
}

func (s *Standard) NotificationStoreSession() notificationStore.Session {
	return s.notificationStoreSession
}

func (s *Standard) PermissionStoreSession() permissionStore.Session {
	return s.permissionStoreSession
}

func (s *Standard) ProfileStoreSession() profileStore.Session {
	return s.profileStoreSession
}

func (s *Standard) SessionStoreSession() sessionStore.Session {
	return s.sessionStoreSession
}

func (s *Standard) UserStoreSession() userStore.Session {
	return s.userStoreSession
}

func (s *Standard) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails

	s.messageStoreSession.SetAgent(authenticationDetails)
	s.notificationStoreSession.SetAgent(authenticationDetails)
	s.permissionStoreSession.SetAgent(authenticationDetails)
	s.profileStoreSession.SetAgent(authenticationDetails)
	s.sessionStoreSession.SetAgent(authenticationDetails)
	s.userStoreSession.SetAgent(authenticationDetails)
}
