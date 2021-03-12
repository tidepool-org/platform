package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataService "github.com/tidepool-org/platform/data/service"
	dataContext "github.com/tidepool-org/platform/data/service/context"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
)

type Standard struct {
	*api.API
	permissionClient        permission.Client
	dataDeduplicatorFactory deduplicator.Factory
	dataStoreDEPRECATED     dataStoreDEPRECATED.Store
	syncTaskStore           syncTaskStore.Store
	dataClient              dataClient.Client
}

func NewStandard(svc service.Service, permissionClient permission.Client,
	dataDeduplicatorFactory deduplicator.Factory,
	dataStoreDEPRECATED dataStoreDEPRECATED.Store, syncTaskStore syncTaskStore.Store, dataClient dataClient.Client) (*Standard, error) {
	if permissionClient == nil {
		return nil, errors.New("permission client is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, errors.New("data deduplicator factory is missing")
	}
	if dataStoreDEPRECATED == nil {
		return nil, errors.New("data store DEPRECATED is missing")
	}
	if syncTaskStore == nil {
		return nil, errors.New("sync task store is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}

	a, err := api.New(svc)
	if err != nil {
		return nil, err
	}

	return &Standard{
		API:                     a,
		permissionClient:        permissionClient,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		dataStoreDEPRECATED:     dataStoreDEPRECATED,
		syncTaskStore:           syncTaskStore,
		dataClient:              dataClient,
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
		return errors.Wrap(err, "unable to create router")
	}

	s.DEPRECATEDAPI().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler dataService.HandlerFunc) rest.HandlerFunc {
	return dataContext.WithContext(s.AuthClient(), s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStoreDEPRECATED, s.syncTaskStore, s.dataClient, handler)
}
