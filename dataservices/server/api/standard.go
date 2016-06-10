package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/dataservices/server"
	"github.com/tidepool-org/platform/dataservices/server/api/v1"
	"github.com/tidepool-org/platform/dataservices/server/context"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	logger           log.Logger
	store            store.Store
	client           client.Client
	versionReporter  version.Reporter
	api              *rest.Api
	statusMiddleware *rest.StatusMiddleware
}

func NewStandard(logger log.Logger, store store.Store, client client.Client, versionReporter version.Reporter) (*Standard, error) {
	if logger == nil {
		return nil, app.Error("api", "logger is missing")
	}
	if store == nil {
		return nil, app.Error("api", "store is missing")
	}
	if client == nil {
		return nil, app.Error("api", "client is missing")
	}
	if versionReporter == nil {
		return nil, app.Error("api", "versionReporter is missing")
	}

	standard := &Standard{
		logger:          logger,
		store:           store,
		client:          client,
		versionReporter: versionReporter,
		api:             rest.NewApi(),
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
	return context.WithContext(s.store, s.client, handler)
}
