package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/twiist"

	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataContext "github.com/tidepool-org/platform/data/service/context"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
)

type Standard struct {
	*api.API
	metricClient                   metric.Client
	permissionClient               permission.Client
	dataDeduplicatorFactory        deduplicator.Factory
	dataStore                      dataStore.Store
	syncTaskStore                  syncTaskStore.Store
	dataClient                     dataClient.Client
	dataSourceClient               dataSource.Client
	twiistServiceAccountAuthorizer twiist.ServiceAccountAuthorizer
}

func NewStandard(svc service.Service, metricClient metric.Client, permissionClient permission.Client,
	dataDeduplicatorFactory deduplicator.Factory,
	store dataStore.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client, dataSourceClient dataSource.Client, twiistServiceAccountAuthorizer twiist.ServiceAccountAuthorizer) (*Standard, error) {
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

	a, err := api.New(svc)
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
		dataSourceClient:               dataSourceClient,
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
	return dataContext.WithContext(s.AuthClient(), s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient, s.dataSourceClient, s.twiistServiceAccountAuthorizer, handler)
}
