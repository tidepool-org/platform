package context

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/clinics"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataSourceService "github.com/tidepool-org/platform/data/source/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	serviceContext "github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/summary"
	summaryReporters "github.com/tidepool-org/platform/summary/reporters"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/twiist"
)

type Standard struct {
	*serviceContext.Responder
	authClient                     auth.Client
	metricClient                   metric.Client
	permissionClient               permission.Client
	dataDeduplicatorFactory        dataDeduplicator.Factory
	dataStore                      dataStore.Store
	dataRepository                 dataStore.DataRepository
	summaryRepository              dataStore.SummaryRepository
	bucketsRepository              dataStore.BucketsRepository
	summarizerRegistry             *summary.SummarizerRegistry
	summaryReporter                *summaryReporters.PatientRealtimeDaysReporter
	syncTaskStore                  syncTaskStore.Store
	syncTasksRepository            syncTaskStore.SyncTaskRepository
	dataClient                     dataClient.Client
	clinicsClient                  clinics.Client
	dataSourceClient               dataSourceService.Client
	alertsRepository               alerts.Repository
	twiistServiceAccountAuthorizer twiist.ServiceAccountAuthorizer
}

func WithContext(authClient auth.Client, metricClient metric.Client, permissionClient permission.Client,
	dataDeduplicatorFactory dataDeduplicator.Factory,
	store dataStore.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client, dataSourceClient dataSourceService.Client, twiistServiceAccountAuthorizer twiist.ServiceAccountAuthorizer, handler dataService.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		standard, standardErr := NewStandard(response, request, authClient, metricClient, permissionClient,
			dataDeduplicatorFactory, store, syncTaskStore, dataClient, dataSourceClient, twiistServiceAccountAuthorizer)
		if standardErr != nil {
			if responder, responderErr := serviceContext.NewResponder(response, request); responderErr != nil {
				response.WriteHeader(http.StatusInternalServerError)
			} else {
				responder.RespondWithInternalServerFailure("Unable to create new context for request", standardErr)
			}
			return
		}
		defer standard.Close()

		handler(standard)
	}
}

func NewStandard(response rest.ResponseWriter, request *rest.Request,
	authClient auth.Client, metricClient metric.Client, permissionClient permission.Client,
	dataDeduplicatorFactory dataDeduplicator.Factory,
	store dataStore.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client, dataSourceClient dataSourceService.Client, twiistServiceAccountAuthorizer twiist.ServiceAccountAuthorizer) (*Standard, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("metric client is missing")
	}
	if permissionClient == nil {
		return nil, errors.New("permission client is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("data deduplicator factory is missing")
	}
	if store == nil {
		return nil, errors.New("data store DEPRECATED is missing")
	}
	if syncTaskStore == nil {
		return nil, errors.New("sync task store is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if twiistServiceAccountAuthorizer == nil {
		return nil, errors.New("twiist service account authorizer is missing")
	}

	responder, err := serviceContext.NewResponder(response, request)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Responder:                      responder,
		authClient:                     authClient,
		metricClient:                   metricClient,
		permissionClient:               permissionClient,
		dataDeduplicatorFactory:        dataDeduplicatorFactory,
		dataStore:                      store,
		syncTaskStore:                  syncTaskStore,
		dataClient:                     dataClient,
		dataSourceClient:               dataSourceClient,
		twiistServiceAccountAuthorizer: twiistServiceAccountAuthorizer,
	}, nil
}

func (s *Standard) Close() {
	s.twiistServiceAccountAuthorizer = nil
	s.alertsRepository = nil
	s.dataSourceClient = nil
	s.clinicsClient = nil
	s.dataClient = nil
	s.syncTasksRepository = nil
	s.syncTaskStore = nil
	s.summaryReporter = nil
	s.summarizerRegistry = nil
	s.bucketsRepository = nil
	s.summaryRepository = nil
	s.dataRepository = nil
	s.dataStore = nil
	s.dataDeduplicatorFactory = nil
	s.permissionClient = nil
	s.metricClient = nil
	s.authClient = nil
}

func (s *Standard) AuthClient() auth.Client {
	return s.authClient
}

func (s *Standard) MetricClient() metric.Client {
	return s.metricClient
}

func (s *Standard) PermissionClient() permission.Client {
	return s.permissionClient
}

func (s *Standard) DataDeduplicatorFactory() dataDeduplicator.Factory {
	return s.dataDeduplicatorFactory
}

func (s *Standard) DataRepository() dataStore.DataRepository {
	if s.dataRepository == nil {
		s.dataRepository = s.dataStore.NewDataRepository()
	}
	return s.dataRepository
}

func (s *Standard) SummaryRepository() dataStore.SummaryRepository {
	if s.summaryRepository == nil {
		s.summaryRepository = s.dataStore.NewSummaryRepository()
	}
	return s.summaryRepository
}

func (s *Standard) SummarizerRegistry() *summary.SummarizerRegistry {
	if s.summarizerRegistry == nil {
		s.summarizerRegistry = summary.New(s.SummaryRepository().GetStore(), s.BucketsRepository().GetStore(), s.DataRepository(), s.GetMongoClient())
	}
	return s.summarizerRegistry
}

func (s *Standard) SummaryReporter() *summaryReporters.PatientRealtimeDaysReporter {
	if s.summaryReporter == nil {
		s.summaryReporter = summaryReporters.NewReporter(s.SummarizerRegistry())
	}
	return s.summaryReporter
}

func (s *Standard) BucketsRepository() dataStore.BucketsRepository {
	if s.bucketsRepository == nil {
		s.bucketsRepository = s.dataStore.NewBucketsRepository()
	}
	return s.bucketsRepository
}

func (s *Standard) SyncTaskRepository() syncTaskStore.SyncTaskRepository {
	if s.syncTasksRepository == nil {
		s.syncTasksRepository = s.syncTaskStore.NewSyncTaskRepository()
	}
	return s.syncTasksRepository
}

func (s *Standard) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Standard) ClinicsClient() clinics.Client {
	if s.clinicsClient == nil {
		var err error
		s.clinicsClient, err = clinics.NewClient(s.AuthClient())
		if err != nil {
			s.Logger().Error("unable to create clinics client")
		}
	}

	return s.clinicsClient
}

func (s *Standard) DataSourceClient() dataSourceService.Client {
	return s.dataSourceClient
}

func (s *Standard) TwiistServiceAccountAuthorizer() twiist.ServiceAccountAuthorizer {
	return s.twiistServiceAccountAuthorizer
}

func (s *Standard) AlertsRepository() alerts.Repository {
	if s.alertsRepository == nil {
		s.alertsRepository = s.dataStore.NewAlertsRepository()
	}
	return s.alertsRepository
}

func (s *Standard) GetMongoClient() *mongo.Client {
	return s.dataStore.GetClient()
}
