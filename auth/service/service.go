package service

import (
	"context"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/appvalidate"
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
	DeviceCheck() apple.DeviceCheck

	Status(context.Context) *Status

	AppValidator() *appvalidate.Validator
}

type Status struct {
	Version   string
	Server    interface{}
	AuthStore interface{}
}
