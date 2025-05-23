package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataDuplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataService "github.com/tidepool-org/platform/data/service"
	dataServiceContext "github.com/tidepool-org/platform/data/service/context"
	dataSourceService "github.com/tidepool-org/platform/data/source/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	serviceApi "github.com/tidepool-org/platform/service/api"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/work"
)

type Standard struct {
	*serviceApi.API
	metricClient                   metric.Client
	permissionClient               permission.Client
	dataDeduplicatorFactory        dataDuplicator.Factory
	dataStore                      dataStore.Store
	syncTaskStore                  syncTaskStore.Store
	dataClient                     dataClient.Client
	dataRawClient                  dataRaw.Client
	dataSourceClient               dataSourceService.Client
	workClient                     work.Client
	twiistServiceAccountAuthorizer auth.ServiceAccountAuthorizer
}

func NewStandard(svc service.Service, metricClient metric.Client, permissionClient permission.Client,
	dataDeduplicatorFactory dataDuplicator.Factory,
	store dataStore.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client,
	dataRawClient dataRaw.Client, dataSourceClient dataSourceService.Client, workClient work.Client,
	twiistServiceAccountAuthorizer auth.ServiceAccountAuthorizer) (*Standard, error) {
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
	if dataRawClient == nil {
		return nil, errors.New("data raw client is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if workClient == nil {
		return nil, errors.New("work client is missing")
	}
	if twiistServiceAccountAuthorizer == nil {
		return nil, errors.New("twiist service account authorizer is missing")
	}

	a, err := serviceApi.New(svc)
	if err != nil {
		return nil, err
	}

	return &Standard{
		API:                            a,
		metricClient:                   metricClient,
		permissionClient:               permissionClient,
		dataDeduplicatorFactory:        dataDeduplicatorFactory,
		dataStore:                      store,
		syncTaskStore:                  syncTaskStore,
		dataClient:                     dataClient,
		dataRawClient:                  dataRawClient,
		dataSourceClient:               dataSourceClient,
		workClient:                     workClient,
		twiistServiceAccountAuthorizer: twiistServiceAccountAuthorizer,
	}, nil
}

func (s *Standard) DEPRECATEDInitializeRouter(routes []dataService.Route) error {
	baseRoutes := []dataService.Route{
		dataService.Get("/status", s.StatusGet),
		dataService.Get("/version", s.VersionGet),
	}

	routes = append(baseRoutes, routes...)

	var contextRoutes []*rest.Route
	for _, route := range routes {
		contextRoutes = append(contextRoutes, route.ToRestRoute(s.withContext))
	}

	router, err := rest.MakeRouter(contextRoutes...)
	if err != nil {
		return errors.Wrap(err, "unable to create router")
	}

	s.DEPRECATEDAPI().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler dataService.HandlerFunc) rest.HandlerFunc {
	return dataServiceContext.WithContext(s.AuthClient(), s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient,
		s.dataRawClient, s.dataSourceClient, s.workClient,
		s.twiistServiceAccountAuthorizer, handler)
}
