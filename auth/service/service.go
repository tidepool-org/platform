package service

import (
	"context"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	authStore "github.com/tidepool-org/platform/auth/store"
	permission "github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/user"
)

type Service interface {
	service.Service

	Domain() string
	AuthStore() authStore.Store
	UserAccessor() user.UserAccessor
	UserProfileAccessor() user.UserProfileAccessor // UserProfileAccessor is separate from UserAccessor while the seagull migration is in progress because the user returned from UserAccessor is the keycloak user and their profile may not have been migrated yet
	PermissionsClient() permission.Client

	ProviderFactory() provider.Factory

	AuthServiceClient() Client
	TaskClient() task.Client
	ConfirmationClient() confirmationClient.ClientWithResponsesInterface
	DeviceCheck() apple.DeviceCheck

	Status(context.Context) *Status

	AppValidator() *appvalidate.Validator

	PartnerSecrets() *appvalidate.PartnerSecrets

	TwiistServiceAccountAuthorizer() auth.ServiceAccountAuthorizer
}

type Status struct {
	Version   string
	Server    any
	AuthStore any
}
