package api

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	dataService "github.com/tidepool-org/platform/data/service"
	dataContext "github.com/tidepool-org/platform/data/service/context"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

type Standard struct {
	*api.API
	permissionClient permission.Client
	dataStore        dataStore.Store
}

func NewStandard(svc service.Service, permissionClient permission.Client,
	store dataStore.Store) (*Standard, error) {
	if permissionClient == nil {
		return nil, errors.New("permission client is missing")
	}
	if store == nil {
		return nil, errors.New("data store DEPRECATED is missing")
	}

	a, err := api.New(svc)
	if err != nil {
		return nil, err
	}

	return &Standard{
		API:              a,
		permissionClient: permissionClient,
		dataStore:        store,
	}, nil
}

func (s *Standard) DEPRECATEDInitializeRouter(routes []dataService.Route, isUploadIdUsed bool) error {
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
			Func:       s.withContext(route.Handler, isUploadIdUsed),
		})
	}
	metricRoute := rest.Get("/metrics", func(w rest.ResponseWriter, r *rest.Request) {
		promhttp.Handler().ServeHTTP(w.(http.ResponseWriter), r.Request)
	})

	contextRoutes = append(contextRoutes, metricRoute)

	router, err := rest.MakeRouter(contextRoutes...)
	if err != nil {
		return errors.Wrap(err, "unable to create router")
	}

	s.DEPRECATEDAPI().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler dataService.HandlerFunc, isUploadIdUsed bool) rest.HandlerFunc {

	return dataContext.WithContext(s.AuthClient(), s.permissionClient, s.dataStore, isUploadIdUsed, handler)
}
