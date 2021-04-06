package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/service/server"
)

type Service struct {
	*application.Application
	secret         string
	authClientImpl auth.Client
	api            *api.API
	server         *server.Standard
}

func New() *Service {
	return &Service{
		Application: application.New(),
	}
}

func (s *Service) Initialize(provider application.Provider) error {
	if err := s.Application.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeSecret(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	return s.initializeServer()
}

func (s *Service) Terminate() {
	s.terminateServer()
	s.terminateAPI()
	s.terminateSecret()

	s.Application.Terminate()
}

func (s *Service) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	s.Logger().Debug("Finalizing middleware")

	if err := s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "unable to initialize middleware")
	}

	return s.server.Serve()
}

func (s *Service) Secret() string {
	return s.secret
}

func (s *Service) AuthClient() auth.Client {
	return s.authClientImpl
}

func (s *Service) SetAuthClient(authClientImpl auth.Client) {
	s.authClientImpl = authClientImpl
}

func (s *Service) API() service.API {
	return s.api
}

func (s *Service) initializeSecret() error {
	s.Logger().Debug("Initializing secret")

	secret := s.ConfigReporter().GetWithDefault("secret", "")
	if secret == "" {
		return errors.New("secret is missing")
	}
	s.secret = secret

	return nil
}

func (s *Service) terminateSecret() {
	if s.secret != "" {
		s.Logger().Debug("Terminating secret")
		s.secret = ""
	}
}

func (s *Service) initializeAPI() error {
	s.Logger().Debug("Creating api")

	a, err := api.New(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api")
	}
	s.api = a

	return nil
}

func (s *Service) terminateAPI() {
	if s.api != nil {
		s.Logger().Debug("Destroying api")
		s.api = nil
	}
}

func (s *Service) initializeServer() error {
	s.Logger().Debug("Loading server config")

	cfg := server.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("server")); err != nil {
		return errors.Wrap(err, "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	svr, err := server.NewStandard(cfg, s.Logger(), s.API())
	if err != nil {
		return errors.Wrap(err, "unable to create server")
	}
	s.server = svr

	return nil
}

func (s *Service) terminateServer() {
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			s.Logger().Errorf("Error while destroying the server: %v", err)
		}
		s.server = nil
	}
}
