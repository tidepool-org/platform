package service

import (
	"context"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification/service"
	"github.com/tidepool-org/platform/notification/service/api"
	notificationServiceApiV1 "github.com/tidepool-org/platform/notification/service/api/v1"
	"github.com/tidepool-org/platform/notification/store"
	notificationMongo "github.com/tidepool-org/platform/notification/store/mongo"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Service struct {
	*serviceService.Authenticated
	notificationStore *notificationMongo.Store
}

func New() *Service {
	return &Service{
		Authenticated: serviceService.NewAuthenticated(),
	}
}

func (s *Service) Initialize(provider application.Provider) error {
	if err := s.Authenticated.Initialize(provider); err != nil {
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

func (s *Service) Status(ctx context.Context) *service.Status {
	return &service.Status{
		Version:           s.VersionReporter().Long(),
		NotificationStore: s.notificationStore.Status(ctx),
		Server:            s.API().Status(),
	}
}

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}

	s.Logger().Debug("Creating v1 router")

	v1Router, err := notificationServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}

	s.Logger().Debug("Initializing routers")

	if err = s.API().InitializeRouters(apiRouter, v1Router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (s *Service) terminateRouter() {
}

func (s *Service) initializeNotificationStore() error {
	s.Logger().Debug("Loading notification store config")

	cfg := storeStructuredMongo.NewConfig(nil)
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load notification store config")
	}

	s.Logger().Debug("Creating notification store")

	params := storeStructuredMongo.Params{DatabaseConfig: cfg}
	str, err := notificationMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create notification store")
	}
	s.notificationStore = str

	return nil
}

func (s *Service) terminateNotificationStore() {
	if s.notificationStore != nil {
		s.Logger().Debug("Closing notification store")
		s.notificationStore.Terminate(nil)

		s.Logger().Debug("Destroying notification store")
		s.notificationStore = nil
	}
}
