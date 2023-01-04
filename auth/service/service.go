package service

import (
	"context"

	"github.com/tidepool-org/platform/apple"

	clinicClient "github.com/tidepool-org/clinic/client"
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
)

type Service interface {
	service.Service

	Domain() string
	AuthStore() store.Store

	ProviderFactory() provider.Factory

	TaskClient() task.Client
	ConfirmationClient() confirmationClient.ClientWithResponsesInterface
	ClinicClient() clinicClient.ClientWithResponsesInterface
	DeviceCheck() apple.DeviceCheck

	Status(context.Context) *Status
}

type Status struct {
	Version   string
	Server    interface{}
	AuthStore interface{}
}
