package service

import (
	"github.com/tidepool-org/platform/auth"
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

	AuthClient() auth.Client
	MetricClient() metricClient.Client
	UserClient() userClient.Client
	DataClient() dataClient.Client

	ConfirmationsSession() confirmationStore.ConfirmationsSession
	MessagesSession() messageStore.MessagesSession
	PermissionsSession() permissionStore.PermissionsSession
	ProfilesSession() profileStore.ProfilesSession
	SessionsSession() sessionStore.SessionsSession
	UsersSession() userStore.UsersSession
}

type HandlerFunc func(context Context)
