package service

import (
	"github.com/tidepool-org/platform/auth"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/metric"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/user"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Context interface {
	service.Context

	AuthClient() auth.Client
	MetricClient() metric.Client
	UserClient() user.Client
	DataClient() dataClient.Client

	ConfirmationSession() confirmationStore.ConfirmationSession
	MessagesSession() messageStore.MessagesSession
	PermissionsSession() permissionStore.PermissionsSession
	ProfilesSession() profileStore.ProfilesSession
	SessionsSession() sessionStore.SessionsSession
	UsersSession() userStore.UsersSession
}

type HandlerFunc func(context Context)
