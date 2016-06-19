package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/dataservices/server/api/v1"
	"github.com/tidepool-org/platform/dataservices/server/context"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
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
	api                     *rest.Api
	statusMiddleware        *rest.StatusMiddleware
}

func NewStandard(logger log.Logger, dataFactory data.Factory, dataStore store.Store, dataDeduplicatorFactory deduplicator.Factory, userServicesClient client.Client, versionReporter version.Reporter) (*Standard, error) {
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
	if versionReporter == nil {
		return nil, app.Error("api", "version reporter is missing")
	}

	standard := &Standard{
		logger:                  logger,
		dataStore:               dataStore,
		dataFactory:             dataFactory,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
		userServicesClient:      userServicesClient,
		versionReporter:         versionReporter,
		api:                     rest.NewApi(),
	}
	if err := standard.initMiddleware(); err != nil {
		return nil, err
	}
	if err := standard.initRouter(); err != nil {
		return nil, err
	}

	return standard, nil
}

func (s *Standard) Handler() http.Handler {
	return s.api.MakeHandler()
}

func (s *Standard) initMiddleware() error {

	s.logger.Debug("Creating API middleware")

	loggerMiddleware, err := service.NewLoggerMiddleware(s.logger)
	if err != nil {
		return err
	}
	traceMiddleware, err := service.NewTraceMiddleware()
	if err != nil {
		return err
	}
	accessLogMiddleware, err := service.NewAccessLogMiddleware()
	if err != nil {
		return err
	}
	recoverMiddleware, err := service.NewRecoverMiddleware()
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

func (s *Standard) initRouter() error {

	s.logger.Debug("Creating API router")

	router, err := rest.MakeRouter(
		rest.Get("/status", s.GetStatus),
		rest.Get("/version", s.GetVersion),
		rest.Post("/api/v1/users/:userid/datasets", s.withContext(v1.Authenticate(v1.UsersDatasetsCreate))),
		rest.Put("/api/v1/datasets/:datasetid", s.withContext(v1.Authenticate(v1.DatasetsUpdate))),
		rest.Post("/api/v1/datasets/:datasetid/data", s.withContext(v1.Authenticate(v1.DatasetsDataCreate))),
	)
	if err != nil {
		return app.ExtError(err, "api", "unable to setup router")
	}

	s.api.SetApp(router)

	return nil
}

func (s *Standard) withContext(handler server.HandlerFunc) rest.HandlerFunc {
	return context.WithContext(s.dataFactory, s.dataStore, s.dataDeduplicatorFactory, s.userServicesClient, handler)
}
