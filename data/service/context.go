package service

import (
	abbottService "github.com/tidepool-org/platform-plugin-abbott/abbott/service"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/clinics"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataSourceService "github.com/tidepool-org/platform/data/source/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/summary"
	summaryReporters "github.com/tidepool-org/platform/summary/reporters"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/work"
)

type Context interface {
	service.Context

	AuthClient() auth.Client
	MetricClient() metric.Client
	PermissionClient() permission.Client

	DataDeduplicatorFactory() deduplicator.Factory

	DataRepository() dataStore.DataRepository
	SummaryRepository() dataStore.SummaryRepository
	SyncTaskRepository() syncTaskStore.SyncTaskRepository
	AlertsRepository() alerts.Repository

	SummarizerRegistry() *summary.SummarizerRegistry
	SummaryReporter() *summaryReporters.PatientRealtimeDaysReporter
	DataClient() dataClient.Client

	ClinicsClient() clinics.Client
	DataRawClient() dataRaw.Client
	DataSourceClient() dataSourceService.Client
	WorkClient() work.Client

	AbbottServiceRequestAuthorizer() abbottService.RequestAuthorizer
	TwiistServiceAccountAuthorizer() auth.ServiceAccountAuthorizer
}

type HandlerFunc func(context Context)
