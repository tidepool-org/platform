package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	versionReporter  version.Reporter
	logger           log.Logger
	authClient       auth.Client
	api              *rest.Api
	statusMiddleware *rest.StatusMiddleware
}

func NewStandard(versionReporter version.Reporter, logger log.Logger, authClient auth.Client) (*Standard, error) {
	if versionReporter == nil {
		return nil, errors.New("api", "version reporter is missing")
	}
	if logger == nil {
		return nil, errors.New("api", "logger is missing")
	}
	if authClient == nil {
		return nil, errors.New("api", "auth client is missing")
	}

	return &Standard{
		versionReporter: versionReporter,
		logger:          logger,
		authClient:      authClient,
		api:             rest.NewApi(),
	}, nil
}

func (s *Standard) VersionReporter() version.Reporter {
	return s.versionReporter
}

func (s *Standard) Logger() log.Logger {
	return s.logger
}

func (s *Standard) AuthClient() auth.Client {
	return s.authClient
}

func (s *Standard) Status() *rest.Status {
	return s.statusMiddleware.GetStatus()
}

func (s *Standard) API() *rest.Api {
	return s.api
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
	authMiddleware, err := middleware.NewAuth(s.AuthClient())
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
		authMiddleware,
		gzipMiddleware,
	}

	s.api.Use(middlewareStack...)

	s.statusMiddleware = statusMiddleware

	return nil
}

func (s *Standard) InitializeRouter(routes ...*rest.Route) error {
	router, err := rest.MakeRouter(routes...)
	if err != nil {
		return errors.Wrap(err, "api", "unable to initializer router")
	}

	s.api.SetApp(router)

	return nil
}
