package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/service/api"
	"github.com/tidepool-org/platform/auth/service/api/v1"
	"github.com/tidepool-org/platform/auth/store"
	authMongo "github.com/tidepool-org/platform/auth/store/mongo"
	"github.com/tidepool-org/platform/errors"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
)

type Service struct {
	*serviceService.Service
	authStore *authMongo.Store
}

func New(prefix string) (*Service, error) {
	svc, err := serviceService.New(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Service: svc,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Service.Initialize(); err != nil {
		return err
	}

	if err := s.initializeRouter(); err != nil {
		return err
	}
	if err := s.initializeAuthStore(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Terminate() {
	if s.authStore != nil {
		s.authStore.Close()
		s.authStore = nil
	}

	s.Service.Terminate()
}

func (s *Service) AuthStore() store.Store {
	return s.authStore
}

func (s *Service) Status() *auth.Status {
	return &auth.Status{
		Version:   s.VersionReporter().Long(),
		AuthStore: s.AuthStore().GetStatus(),
		Server:    s.API().Status(),
	}
}

func (s *Service) initializeRouter() error {
	routes := []*rest.Route{}

	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create api router")
	}
	routes = append(routes, apiRouter.Routes()...)

	s.Logger().Debug("Creating v1 router")

	v1Router, err := v1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create v1 router")
	}
	routes = append(routes, v1Router.Routes()...)

	s.Logger().Debug("Initializing router")

	if err = s.API().InitializeRouter(routes...); err != nil {
		return errors.Wrap(err, "service", "unable to initialize router")
	}

	return nil
}

func (s *Service) initializeAuthStore() error {
	s.Logger().Debug("Loading auth store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("auth", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load auth store config")
	}
	cfg.Collection = "auths"

	s.Logger().Debug("Creating auth store")

	str, err := authMongo.New(s.Logger(), cfg)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create auth store")
	}
	s.authStore = str

	return nil
}
