package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
)

type Service struct {
	*application.Application
	authClient *authClient.Client
}

func New(prefix string) (*Service, error) {
	app, err := application.New(prefix)
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

	return nil
}

func (s *Service) Terminate() {
	if s.authClient != nil {
		s.authClient.Close()
		s.authClient = nil
	}

	s.Application.Terminate()
}

func (s *Service) initializeAuthClient() error {
	s.Logger().Debug("Loading auth client config")

	authClientConfig := authClient.NewConfig()
	if err := authClientConfig.Load(s.ConfigReporter().WithScopes("auth", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load auth client config")
	}

	s.Logger().Debug("Creating auth client")

	authClient, err := authClient.NewClient(authClientConfig, s.Name(), s.Logger())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create auth client")
	}
	s.authClient = authClient

	s.Logger().Debug("Starting auth client")

	if err = s.authClient.Start(); err != nil {
		return errors.Wrap(err, "service", "unable to start auth client")
	}

	return nil
}

func (s *Service) AuthClient() auth.Client {
	return s.authClient
}
