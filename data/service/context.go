package service

import (
	"github.com/tidepool-org/platform/auth"
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

	AuthClient() auth.Client
	MetricClient() metricClient.Client
	UserClient() userClient.Client

	DataFactory() data.Factory
	DataDeduplicatorFactory() deduplicator.Factory

	DataSession() dataStore.DataSession
	SyncTasksSession() syncTaskStore.SyncTasksSession
}

type HandlerFunc func(context Context)
