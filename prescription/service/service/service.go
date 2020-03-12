package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/prescription/store/mongo"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Service struct {
	*serviceService.Service
	prescriptionStore *mongo.Store
}

func New() *Service {
	return &Service{
		Service: serviceService.New(),
	}
}

func (s *Service) Initialize(provider application.Provider) error {
	if err := s.Service.Initialize(provider); err != nil {
		return err
	}
	return s.initializePrescriptionStore()
}

func (s *Service) Terminate() {
	s.terminatePrescriptionStore()
	s.Service.Terminate()
}

func (s *Service) Status() *service.Status {
	return &service.Status{
		Version: s.VersionReporter().Long(),
		Store:   s.prescriptionStore.Status(),
		Server:  s.API().Status(),
	}
}

func (s *Service) PrescriptionStore() *mongo.Store {
	return s.prescriptionStore
}

func (s *Service) initializePrescriptionStore() error {
	s.Logger().Debug("Initializing prescription store")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("prescription", "store")); err != nil {
		return errors.Wrap(err, "unable to load prescription store config")
	}

	s.Logger().Debug("Creating prescription store")
	store, err := mongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create prescription store")
	}
	s.prescriptionStore = store

	s.Logger().Debug("Ensuring prescription store indexes")
	err = s.prescriptionStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure prescription store indexes")
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
