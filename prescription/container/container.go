package container

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/prescription"
	prescriptionService "github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/prescription/status"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
	"github.com/tidepool-org/platform/version"
)

type Container interface {
	Initialize() error
	PrescriptionStore() store.Store
	PrescriptionService() prescription.Service
	UserClient() user.Client
	StatusReporter() status.Reporter
}

type PrescriptionContainer struct {
	configReporter      config.Reporter
	logger              log.Logger
	prescriptionStore   *mongo.Store
	prescriptionService prescription.Service
	statusReporter      status.Reporter
	userAgent           string
	userClient          user.Client
	versionReporter     version.Reporter
}

type Params struct {
	ConfigReporter  config.Reporter
	Logger          log.Logger
	UserAgent       string
	VersionReporter version.Reporter
}

func New(params *Params) Container {
	return &PrescriptionContainer{
		configReporter:  params.ConfigReporter,
		logger:          params.Logger,
		userAgent:       params.UserAgent,
		versionReporter: params.VersionReporter,
	}
}

func (s *PrescriptionContainer) Initialize() error {
	if err := s.initializePrescriptionStore(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
		return err
	}

	return s.initializePrescriptionService()
}

func (s *PrescriptionContainer) Terminate() {
	s.terminatePrescriptionStore()
	s.terminatePrescriptionClient()
	s.terminateUserClient()
}

func (s *PrescriptionContainer) PrescriptionStore() store.Store {
	return s.prescriptionStore
}

func (s *PrescriptionContainer) UserClient() user.Client {
	return s.userClient
}

func (s *PrescriptionContainer) PrescriptionService() prescription.Service {
	return s.prescriptionService
}

func (s *PrescriptionContainer) StatusReporter() status.Reporter {
	return s.statusReporter
}

func (s *PrescriptionContainer) initializePrescriptionStore() error {
	s.logger.Debug("Initializing prescription store")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.configReporter.WithScopes("prescription", "store")); err != nil {
		return errors.Wrap(err, "unable to load prescription store config")
	}

	s.logger.Debug("Creating prescription store")
	st, err := mongo.NewStore(cfg, s.logger)
	if err != nil {
		return errors.Wrap(err, "unable to create prescription store")
	}
	s.prescriptionStore = st

	s.logger.Debug("Ensuring prescription store indexes")
	err = s.prescriptionStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure prescription store indexes")
	}

	return nil
}

func (s *PrescriptionContainer) initializePrescriptionService() error {
	s.logger.Debug("Prescription client")

	clnt, err := prescriptionService.New(s.logger, s.PrescriptionStore())
	if err != nil {
		return errors.Wrap(err, "unable to create prescription client")
	}
	s.prescriptionService = clnt

	return nil
}

func (s *PrescriptionContainer) initializeUserClient() error {
	s.logger.Debug("Loading user client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.userAgent
	if err := cfg.Load(s.configReporter.WithScopes("user", "client")); err != nil {
		return errors.Wrap(err, "unable to get user client config")
	}

	s.logger.Debug("Creating user client")

	clnt, err := userClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	s.userClient = clnt

	return nil
}

func (s *PrescriptionContainer) initializeStatusReporter() error {
	s.logger.Debug("Initializing status reporter")
	s.statusReporter = status.NewReporter(s.versionReporter, s.prescriptionStore)
	return nil
}

func (s *PrescriptionContainer) terminatePrescriptionStore() {
	if s.prescriptionStore != nil {
		s.logger.Debug("Terminating prescription store")
		if err := s.prescriptionStore.Close(); err != nil {
			s.logger.WithError(err).Warn("Unable to terminate prescription store")
		}

		s.prescriptionStore = nil
	}
}

func (s *PrescriptionContainer) terminatePrescriptionClient() {
	if s.prescriptionService != nil {
		s.logger.Debug("Destroying prescription client")
		s.prescriptionService = nil
	}
}

func (s *PrescriptionContainer) terminateUserClient() {
	if s.userClient != nil {
		s.logger.Debug("Destroying user client")
		s.userClient = nil
	}
}

func (s *PrescriptionContainer) terminateRouter() {
}
