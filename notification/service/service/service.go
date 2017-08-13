package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification"
	"github.com/tidepool-org/platform/notification/service/api"
	"github.com/tidepool-org/platform/notification/service/api/v1"
	"github.com/tidepool-org/platform/notification/store"
	notificationMongo "github.com/tidepool-org/platform/notification/store/mongo"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
)

type Service struct {
	*serviceService.Service
	notificationStore *notificationMongo.Store
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
	if err := s.initializeNotificationStore(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Terminate() {
	if s.notificationStore != nil {
		s.notificationStore.Close()
		s.notificationStore = nil
	}

	s.Service.Terminate()
}

func (s *Service) NotificationStore() store.Store {
	return s.notificationStore
}

func (s *Service) Status() *notification.Status {
	return &notification.Status{
		Version:           s.VersionReporter().Long(),
		NotificationStore: s.NotificationStore().GetStatus(),
		Server:            s.API().Status(),
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

func (s *Service) initializeNotificationStore() error {
	s.Logger().Debug("Loading notification store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("notification", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load notification store config")
	}
	cfg.Collection = "notifications"

	s.Logger().Debug("Creating notification store")

	str, err := notificationMongo.New(s.Logger(), cfg)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create notification store")
	}
	s.notificationStore = str

	return nil
}
