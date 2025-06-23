package service

import (
	"context"

	"github.com/tidepool-org/platform/twiist"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth/store"
	permission "github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/user"
)

type Service interface {
	service.Service

	Domain() string
	AuthStore() store.Store
	UserAccessor() user.UserAccessor
	ProfileAccessor() user.ProfileAccessor
	PermissionsClient() permission.ExtendedClient

	ProviderFactory() provider.Factory

	TaskClient() task.Client
	ConfirmationClient() confirmationClient.ClientWithResponsesInterface
	DeviceCheck() apple.DeviceCheck

	Status(context.Context) *Status

	AppValidator() *appvalidate.Validator

	PartnerSecrets() *appvalidate.PartnerSecrets

	TwiistServiceAccountAuthorizer() twiist.ServiceAccountAuthorizer
}

type Status struct {
	Version   string
	Server    interface{}
	AuthStore interface{}
}
