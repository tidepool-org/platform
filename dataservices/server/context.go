package server

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

type Context interface {
	service.Context

	DataFactory() data.Factory
	DataStoreSession() store.Session
	DataDeduplicatorFactory() deduplicator.Factory
	UserServicesClient() client.Client

	RequestUserID() string
	SetRequestUserID(requestUserID string)
}

type HandlerFunc func(context Context)
