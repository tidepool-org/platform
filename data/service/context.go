package service

import (
	"github.com/tidepool-org/platform/auth"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
)

type Context interface {
	service.Context

	AuthClient() auth.Client
	PermissionClient() permission.Client

	DataRepository() dataStore.DataRepository
	IsUploadIdUsed() bool
}

type HandlerFunc func(context Context)
