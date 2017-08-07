package service

import (
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userStore "github.com/tidepool-org/platform/user/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Context interface {
	service.Context

	MetricServicesClient() metricservicesClient.Client
	UserServicesClient() userservicesClient.Client
	DataServicesClient() dataservicesClient.Client

	ConfirmationStoreSession() confirmationStore.Session
	MessageStoreSession() messageStore.Session
	PermissionStoreSession() permissionStore.Session
	ProfileStoreSession() profileStore.Session
	SessionStoreSession() sessionStore.Session
	UserStoreSession() userStore.Session

	AuthenticationDetails() userservicesClient.AuthenticationDetails
	SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails)
}

type HandlerFunc func(context Context)
