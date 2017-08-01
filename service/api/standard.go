package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	versionReporter  version.Reporter
	logger           log.Logger
	api              *rest.Api
	headerMiddleware *middleware.Header
	statusMiddleware *rest.StatusMiddleware
}

func NewStandard(versionReporter version.Reporter, logger log.Logger) (*Standard, error) {
	if versionReporter == nil {
		return nil, errors.New("api", "version reporter is missing")
	}
	if logger == nil {
		return nil, errors.New("api", "logger is missing")
	}

	return &Standard{
		versionReporter: versionReporter,
		logger:          logger,
		api:             rest.NewApi(),
	}, nil
}

func (s *Standard) VersionReporter() version.Reporter {
	return s.versionReporter
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) API() *rest.Api {
	return s.api
}

func (s *Standard) HeaderMiddleware() *middleware.Header {
	return s.headerMiddleware
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
	headerMiddleware, err := middleware.NewHeader()
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
		headerMiddleware,
		accessLogMiddleware,
		statusMiddleware,
		timerMiddleware,
		recorderMiddleware,
		recoverMiddleware,
		gzipMiddleware,
	}

	s.api.Use(middlewareStack...)

	s.headerMiddleware = headerMiddleware
	s.statusMiddleware = statusMiddleware

	return nil
}
