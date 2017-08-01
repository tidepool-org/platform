package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/dataservices/service/context"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service/api"
	taskStore "github.com/tidepool-org/platform/task/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	*api.Standard
	metricServicesClient    metricservicesClient.Client
	userServicesClient      userservicesClient.Client
	dataFactory             data.Factory
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               dataStore.Store
	taskStore               taskStore.Store
}

func NewStandard(versionReporter version.Reporter, logger log.Logger,
	metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client,
	dataFactory data.Factory, dataDeduplicatorFactory deduplicator.Factory,
	dataStore dataStore.Store, taskStore taskStore.Store) (*Standard, error) {
	if versionReporter == nil {
		return nil, errors.New("api", "version reporter is missing")
	}
	if logger == nil {
		return nil, errors.New("api", "logger is missing")
	}
	if metricServicesClient == nil {
		return nil, errors.New("api", "metric services client is missing")
	}
	if userServicesClient == nil {
		return nil, errors.New("api", "user services client is missing")
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
	if taskStore == nil {
		return nil, errors.New("api", "task store is missing")
	}

	standard, err := api.NewStandard(versionReporter, logger)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard:                standard,
		metricServicesClient:    metricServicesClient,
		userServicesClient:      userServicesClient,
		dataFactory:             dataFactory,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		dataStore:               dataStore,
		taskStore:               taskStore,
	}, nil
}

func (s *Standard) InitializeRouter(routes []service.Route) error {
	baseRoutes := []service.Route{
		service.MakeRoute("GET", "/status", s.GetStatus),
		service.MakeRoute("GET", "/version", s.GetVersion),
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

	s.API().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler service.HandlerFunc) rest.HandlerFunc {
	return context.WithContext(s.metricServicesClient, s.userServicesClient,
		s.dataFactory, s.dataDeduplicatorFactory,
		s.dataStore, s.taskStore, handler)
}
