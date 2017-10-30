package service

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/service"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/user"
)

type Context interface {
	service.Context

	AuthClient() auth.Client
	MetricClient() metric.Client
	UserClient() user.Client

	DataFactory() data.Factory
	DataDeduplicatorFactory() deduplicator.Factory

	DataSession() dataStoreDEPRECATED.DataSession
	SyncTaskSession() syncTaskStore.SyncTaskSession

	DataClient() dataClient.Client
}

type HandlerFunc func(context Context)
