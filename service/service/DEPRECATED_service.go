package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
)

type DEPRECATEDService struct {
	*application.Application
	authClient *authClient.Client
}

func NewDEPRECATEDService(prefix string) (*DEPRECATEDService, error) {
	app, err := application.New(prefix, "service")
	if err != nil {
		return nil, err
	}

	return &DEPRECATEDService{
		Application: app,
	}, nil
}

func (s *DEPRECATEDService) Initialize() error {
	if err := s.Application.Initialize(); err != nil {
		return err
	}

	if err := s.initializeAuthClient(); err != nil {
		return err
	}

	return nil
}

func (s *DEPRECATEDService) Terminate() {
	if s.authClient != nil {
		s.authClient.Close()
		s.authClient = nil
	}

	s.Application.Terminate()
}

func (s *DEPRECATEDService) AuthClient() auth.Client {
	return s.authClient
}

func (s *DEPRECATEDService) initializeAuthClient() error {
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
