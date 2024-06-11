package service

import (
	"context"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/user"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/auth/store"
	permission "github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
)

type Service interface {
	service.Service

	Domain() string
	AuthStore() store.Store
	UserAccessor() user.UserAccessor
	UserProfileAccessor() user.UserProfileAccessor // UserProfileAccessor is separate from UserAccessor while the seagull migration is in progress because the user returned from UserAccessor is the keycloak user and their profile may not have been migrated yet
	PermissionsClient() permission.Client

	ProviderFactory() provider.Factory

	TaskClient() task.Client
	ConfirmationClient() confirmationClient.ClientWithResponsesInterface
	DeviceCheck() apple.DeviceCheck

	Status(context.Context) *Status
}

type Status struct {
	Version   string
	Server    interface{}
	AuthStore interface{}
}
