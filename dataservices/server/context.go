package server

import (
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

type Context interface {
	service.Context

	DataStoreSession() store.Session
	UserServicesClient() client.Client

	RequestUserID() string
	SetRequestUserID(requestUserID string)
}

type HandlerFunc func(context Context)
