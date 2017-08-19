package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataContext "github.com/tidepool-org/platform/data/service/context"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	usersClient "github.com/tidepool-org/platform/user/client"
)

type Standard struct {
	*api.API
	metricClient            metricClient.Client
	userClient              usersClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               dataStore.Store
	syncTaskStore           syncTaskStore.Store
}

func NewStandard(svc service.Service, metricClient metricClient.Client, userClient usersClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, syncTaskStore syncTaskStore.Store) (*Standard, error) {
	if metricClient == nil {
		return nil, errors.New("api", "metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("api", "user client is missing")
	}
	if dataFactory == nil {
		return nil, errors.New("api", "data factory is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("api", "data deduplicator factory is missing")
	}
	if dataStore == nil {
		return nil, errors.New("api", "data store is missing")
	}
	if syncTaskStore == nil {
		return nil, errors.New("api", "sync task store is missing")
	}

	a, err := api.New(svc)
	if err != nil {
		return nil, err
	}

	return &Standard{
		API:                     a,
		metricClient:            metricClient,
		userClient:              userClient,
		dataFactory:             dataFactory,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		dataStore:               dataStore,
		syncTaskStore:           syncTaskStore,
	}, nil
}

func (s *Standard) DEPRECATEDInitializeRouter(routes []dataService.Route) error {
	baseRoutes := []dataService.Route{
		dataService.MakeRoute("GET", "/status", s.StatusGet),
		dataService.MakeRoute("GET", "/version", s.VersionGet),
	}

	routes = append(baseRoutes, routes...)

	var contextRoutes []*rest.Route
	for _, route := range routes {
		contextRoutes = append(contextRoutes, &rest.Route{
			HttpMethod: route.Method,
			PathExp:    route.Path,
			Func:       s.withContext(route.Handler),
		})
	}

	router, err := rest.MakeRouter(contextRoutes...)
	if err != nil {
		return errors.Wrap(err, "api", "unable to create router")
	}

	s.DEPRECATEDAPI().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler dataService.HandlerFunc) rest.HandlerFunc {
	return dataContext.WithContext(s.AuthClient(), s.metricClient, s.userClient,
		s.dataFactory, s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, handler)
}
