package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	versionReporter     version.Reporter
	environmentReporter environment.Reporter
	logger              log.Logger
	api                 *rest.Api
	statusMiddleware    *rest.StatusMiddleware
}

func NewStandard(versionReporter version.Reporter, environmentReporter environment.Reporter, logger log.Logger) (*Standard, error) {
	if versionReporter == nil {
		return nil, app.Error("api", "version reporter is missing")
	}
	if environmentReporter == nil {
		return nil, app.Error("api", "environment reporter is missing")
	}
	if logger == nil {
		return nil, app.Error("api", "logger is missing")
	}

	return &Standard{
		versionReporter:     versionReporter,
		environmentReporter: environmentReporter,
		logger:              logger,
		api:                 rest.NewApi(),
	}, nil
}

func (s *Standard) VersionReporter() version.Reporter {
	return s.versionReporter
}

func (s *Standard) EnvironmentReporter() environment.Reporter {
	return s.environmentReporter
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) API() *rest.Api {
	return s.api
}

func (s *Standard) StatusMiddleware() *rest.StatusMiddleware {
	return s.statusMiddleware
}

func (s *Standard) Handler() http.Handler {
	return s.api.MakeHandler()
}

func (s *Standard) InitializeMiddleware() error {
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
