package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification/service"
	"github.com/tidepool-org/platform/notification/service/api"
	"github.com/tidepool-org/platform/notification/service/api/v1"
	"github.com/tidepool-org/platform/notification/store"
	notificationMongo "github.com/tidepool-org/platform/notification/store/mongo"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
)

type Service struct {
	*serviceService.Authenticated
	notificationStore *notificationMongo.Store
}

func New(prefix string) (*Service, error) {
	authenticated, err := serviceService.NewAuthenticated(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Authenticated: authenticated,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Authenticated.Initialize(); err != nil {
		return err
	}

	if err := s.initializeRouter(); err != nil {
		return err
	}
	return s.initializeNotificationStore()
}

func (s *Service) Terminate() {
	s.terminateNotificationStore()
	s.terminateRouter()

	s.Authenticated.Terminate()
}

func (s *Service) NotificationStore() store.Store {
	return s.notificationStore
}

func (s *Service) Status() *service.Status {
	return &service.Status{
		Version:           s.VersionReporter().Long(),
		NotificationStore: s.NotificationStore().Status(),
		Server:            s.API().Status(),
	}
}

func (s *Service) initializeRouter() error {
	routes := []*rest.Route{}

	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}
	routes = append(routes, apiRouter.Routes()...)

	s.Logger().Debug("Creating v1 router")

	v1Router, err := v1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}
	routes = append(routes, v1Router.Routes()...)

	s.Logger().Debug("Initializing router")

	if err = s.API().InitializeRouter(routes...); err != nil {
		return errors.Wrap(err, "unable to initialize router")
	}

	return nil
}

func (s *Service) terminateRouter() {
}

func (s *Service) initializeNotificationStore() error {
	s.Logger().Debug("Loading notification store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("notification", "store")); err != nil {
		return errors.Wrap(err, "unable to load notification store config")
	}

	s.Logger().Debug("Creating notification store")

	str, err := notificationMongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create notification store")
	}
	s.notificationStore = str

	return nil
}

func (s *Service) terminateNotificationStore() {
	if s.notificationStore != nil {
		s.Logger().Debug("Closing notification store")
		s.notificationStore.Close()

		s.Logger().Debug("Destroying notification store")
		s.notificationStore = nil
	}
}
