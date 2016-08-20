package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/dataservices/service/context"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	logger                  log.Logger
	dataFactory             data.Factory
	dataStore               store.Store
	dataDeduplicatorFactory deduplicator.Factory
	userServicesClient      client.Client
	versionReporter         version.Reporter
	environmentReporter     environment.Reporter
	api                     *rest.Api
	statusMiddleware        *rest.StatusMiddleware
}

func NewStandard(versionReporter version.Reporter, environmentReporter environment.Reporter, logger log.Logger, dataFactory data.Factory, dataStore store.Store, dataDeduplicatorFactory deduplicator.Factory, userServicesClient client.Client, routes []service.Route) (*Standard, error) {
	if versionReporter == nil {
		return nil, app.Error("api", "version reporter is missing")
	}
	if environmentReporter == nil {
		return nil, app.Error("api", "environment reporter is missing")
	}
	if logger == nil {
		return nil, app.Error("api", "logger is missing")
	}
	if dataFactory == nil {
		return nil, app.Error("api", "data factory is missing")
	}
	if dataStore == nil {
		return nil, app.Error("api", "data store is missing")
	}
	if dataDeduplicatorFactory == nil {
		return nil, app.Error("api", "data deduplicator factory is missing")
	}
	if userServicesClient == nil {
		return nil, app.Error("api", "user services client is missing")
	}
	if routes == nil {
		return nil, app.Error("api", "routes is missing")
	}

	standard := &Standard{
		versionReporter:         versionReporter,
		environmentReporter:     environmentReporter,
		logger:                  logger,
		dataStore:               dataStore,
		dataFactory:             dataFactory,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		userServicesClient:      userServicesClient,
		api:                     rest.NewApi(),
	}
	if err := standard.initMiddleware(); err != nil {
		return nil, err
	}
	if err := standard.initRouter(routes); err != nil {
		return nil, err
	}

	return standard, nil
}

func (s *Standard) Handler() http.Handler {
	return s.api.MakeHandler()
}

func (s *Standard) initMiddleware() error {

	s.logger.Debug("Creating API middleware")

	loggerMiddleware, err := middleware.NewLogger(s.logger)
	if err != nil {
		return err
	}
	traceMiddleware, err := middleware.NewTrace()
	if err != nil {
		return err
	}
	accessLogMiddleware, err := middleware.NewAccessLog()
	if err != nil {
		return err
	}
	recoverMiddleware, err := middleware.NewRecover()
	if err != nil {
		return err
	}

	statusMiddleware := &rest.StatusMiddleware{}
	timerMiddleware := &rest.TimerMiddleware{}
	recorderMiddleware := &rest.RecorderMiddleware{}
	gzipMiddleware := &rest.GzipMiddleware{}

	middlewareStack := []rest.Middleware{
		loggerMiddleware,
		traceMiddleware,
		accessLogMiddleware,
		statusMiddleware,
		timerMiddleware,
		recorderMiddleware,
		recoverMiddleware,
		gzipMiddleware,
	}

	s.api.Use(middlewareStack...)

	s.statusMiddleware = statusMiddleware

	return nil
}

func (s *Standard) initRouter(routes []service.Route) error {

	s.logger.Debug("Creating API router")

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
		return app.ExtError(err, "api", "unable to create router")
	}

	s.api.SetApp(router)

	return nil
}

func (s *Standard) withContext(handler service.HandlerFunc) rest.HandlerFunc {
	return context.WithContext(s.dataFactory, s.dataStore, s.dataDeduplicatorFactory, s.userServicesClient, handler)
}
