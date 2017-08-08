package service

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataStore "github.com/tidepool-org/platform/data/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/service"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Context interface {
	service.Context

	MetricClient() metricClient.Client
	UserClient() userClient.Client

	DataFactory() data.Factory
	DataDeduplicatorFactory() deduplicator.Factory

	DataStoreSession() dataStore.Session
	SyncTaskStoreSession() syncTaskStore.Session

	AuthenticationDetails() userClient.AuthenticationDetails
	SetAuthenticationDetails(authenticationDetails userClient.AuthenticationDetails)
}

type HandlerFunc func(context Context)
