package service

import (
	"context"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/appvalidate"
	authStore "github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/twiist"
)

type Service interface {
	service.Service

	Domain() string
	AuthStore() authStore.Store

	ProviderFactory() provider.Factory

	AuthServiceClient() Client
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
	Server    any
	AuthStore any
}
