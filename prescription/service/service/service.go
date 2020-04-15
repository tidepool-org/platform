package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/prescription/service/api"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Service struct {
	*serviceService.Authenticated
	prescriptionStore  *mongo.Store
	prescriptionClient prescription.Client
	userClient         user.Client
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
	if err := s.initializePrescriptionStore(); err != nil {
		return err
	}
	if err := s.initializeRouter(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
		return err
	}

	return s.initializePrescriptionClient()
}

func (s *Service) Terminate() {
	s.terminatePrescriptionStore()
	s.terminatePrescriptionClient()
	s.terminateUserClient()
	s.Service.Terminate()
}

func (s *Service) Status() *service.Status {
	return &service.Status{
		Version: s.VersionReporter().Long(),
		Store:   s.prescriptionStore.Status(),
		Server:  s.API().Status(),
	}
}

func (s *Service) PrescriptionStore() store.Store {
	return s.prescriptionStore
}

func (s *Service) UserClient() user.Client {
	return s.userClient
}

func (s *Service) PrescriptionClient() prescription.Client {
	return s.prescriptionClient
}

func (s *Service) initializePrescriptionStore() error {
	s.Logger().Debug("Initializing prescription store")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("prescription", "store")); err != nil {
		return errors.Wrap(err, "unable to load prescription store config")
	}

	s.Logger().Debug("Creating prescription store")
	st, err := mongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create prescription store")
	}
	s.prescriptionStore = st

	s.Logger().Debug("Ensuring prescription store indexes")
	err = s.prescriptionStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure prescription store indexes")
	}

	return nil
}

func (s *Service) initializePrescriptionClient() error {
	s.Logger().Debug("Prescription client")

	clnt, err := NewClient(s.Logger(), s.PrescriptionStore())
	if err != nil {
		return errors.Wrap(err, "unable to create prescription client")
	}
	s.prescriptionClient = clnt

	return nil
}

func (s *Service) initializeUserClient() error {
	s.Logger().Debug("Loading user client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	if err := cfg.Load(s.ConfigReporter().WithScopes("user", "client")); err != nil {
		return errors.Wrap(err, "unable to get user client config")
	}

	s.Logger().Debug("Creating user client")

	clnt, err := userClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	s.userClient = clnt

	return nil
}

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating prescription router")

	router, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create prescription api router")
	}

	s.Logger().Debug("Initializing router")

	if err = s.API().InitializeRouters(router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (s *Service) terminatePrescriptionStore() {
	if s.prescriptionStore != nil {
		s.Logger().Debug("Terminating prescription store")
		if err := s.prescriptionStore.Close(); err != nil {
			s.Logger().WithError(err).Warn("Unable to terminate prescription store")
		}

		s.prescriptionStore = nil
	}
}

func (s *Service) terminatePrescriptionClient() {
	if s.prescriptionClient != nil {
		s.Logger().Debug("Destroying prescription client")
		s.prescriptionClient = nil
	}
}

func (s *Service) terminateUserClient() {
	if s.userClient != nil {
		s.Logger().Debug("Destroying user client")
		s.userClient = nil
	}
}

func (s *Service) terminateRouter() {
}