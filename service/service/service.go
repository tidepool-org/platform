package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/service/server"
)

type Service struct {
	*application.Application
	authClient *authClient.Client
	api        *api.Standard
	server     *server.Standard
}

func New(prefix string) (*Service, error) {
	app, err := application.New(prefix, "service")
	if err != nil {
		return nil, err
	}

	return &Service{
		Application: app,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Application.Initialize(); err != nil {
		return err
	}

	if err := s.initializeAuthClient(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	if err := s.initializeServer(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Terminate() {
	s.server = nil
	s.api = nil
	if s.authClient != nil {
		s.authClient.Close()
		s.authClient = nil
	}

	s.Application.Terminate()
}

func (s *Service) Run() error {
	if s.server == nil {
		return errors.New("service", "service not initialized")
	}

	return s.server.Serve()
}

func (s *Service) AuthClient() auth.Client {
	return s.authClient
}

func (s *Service) API() service.API {
	return s.api
}

func (s *Service) initializeAuthClient() error {
	s.Logger().Debug("Loading auth client config")

	cfg := authClient.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("auth", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load auth client config")
	}

	s.Logger().Debug("Creating auth client")

	clnt, err := authClient.NewClient(cfg, s.Name(), s.Logger())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create auth client")
	}
	s.authClient = clnt

	s.Logger().Debug("Starting auth client")

	if err = s.authClient.Start(); err != nil {
		return errors.Wrap(err, "service", "unable to start auth client")
	}

	return nil
}

func (s *Service) initializeAPI() error {
	s.Logger().Debug("Creating api")

	apy, err := api.NewStandard(s.VersionReporter(), s.Logger(), s.AuthClient())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create api")
	}
	s.api = apy

	s.Logger().Debug("Initializing middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "service", "unable to initialize middleware")
	}

	return nil
}

func (s *Service) initializeServer() error {
	s.Logger().Debug("Loading server config")

	cfg := server.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("server")); err != nil {
		return errors.Wrap(err, "service", "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	svr, err := server.NewStandard(s.Logger(), s.API(), cfg)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create server")
	}
	s.server = svr

	return nil
}
