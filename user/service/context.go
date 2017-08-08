package service

import (
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userClient "github.com/tidepool-org/platform/user/client"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Context interface {
	service.Context

	MetricClient() metricClient.Client
	UserClient() userClient.Client
	DataClient() dataClient.Client

	ConfirmationStoreSession() confirmationStore.Session
	MessageStoreSession() messageStore.Session
	PermissionStoreSession() permissionStore.Session
	ProfileStoreSession() profileStore.Session
	SessionStoreSession() sessionStore.Session
	UserStoreSession() userStore.Session

	AuthenticationDetails() userClient.AuthenticationDetails
	SetAuthenticationDetails(authenticationDetails userClient.AuthenticationDetails)
}

type HandlerFunc func(context Context)
