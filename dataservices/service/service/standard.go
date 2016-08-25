package service

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/delegate"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/dataservices/service"
	"github.com/tidepool-org/platform/dataservices/service/api"
	"github.com/tidepool-org/platform/dataservices/service/api/v1"
	"github.com/tidepool-org/platform/dataservices/service/server"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	commonMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	versionReporter         version.Reporter
	environmentReporter     environment.Reporter
	configLoader            config.Loader
	logger                  log.Logger
	dataFactory             data.Factory
	dataStore               store.Store
	dataDeduplicatorFactory deduplicator.Factory
	userServicesClient      client.Client
	dataServicesAPI         service.API
	dataServicesServer      service.Server
}

func NewStandard() (*Standard, error) {
	return &Standard{}, nil
}

func (s *Standard) Close() {
	s.dataServicesServer = nil
	s.dataServicesAPI = nil
	if s.userServicesClient != nil {
		s.userServicesClient.Close()
		s.userServicesClient = nil
	}
	s.dataDeduplicatorFactory = nil
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
	s.dataFactory = nil
	s.logger = nil
	s.configLoader = nil
	s.environmentReporter = nil
	s.versionReporter = nil
}

func (s *Standard) Initialize() error {
	if s.dataServicesServer != nil {
		return app.Error("dataservices", "service already initialized")
	}

	if err := s.initializeVersionReporter(); err != nil {
		return err
	}
	if err := s.initializeEnvironmentReporter(); err != nil {
		return err
	}
	if err := s.initializeConfigLoader(); err != nil {
		return err
	}
	if err := s.initializeLogger(); err != nil {
		return err
	}
	if err := s.initializeDataFactory(); err != nil {
		return err
	}
	if err := s.initializeDataStore(); err != nil {
		return err
	}
	if err := s.initializeDataDeduplicatorFactory(); err != nil {
		return err
	}
	if err := s.initializeUserServicesClient(); err != nil {
		return err
	}
	if err := s.initializeDataServicesAPI(); err != nil {
		return err
	}
	if err := s.initializeDataServicesServer(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Run() error {
	if s.dataServicesServer == nil {
		return app.Error("dataservices", "service not initialized")
	}

	return s.dataServicesServer.Serve()
}

func (s *Standard) API() service.API {
	return s.dataServicesAPI
}

func (s *Standard) Server() service.Server {
	return s.dataServicesServer
}

func (s *Standard) initializeVersionReporter() error {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create version reporter")
	}
	s.versionReporter = versionReporter

	return nil
}

func (s *Standard) initializeEnvironmentReporter() error {
	environmentReporter, err := environment.NewDefaultReporter()
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create environment reporter")
	}
	s.environmentReporter = environmentReporter

	return nil
}

func (s *Standard) initializeConfigLoader() error {
	configLoader, err := config.NewLoader(s.environmentReporter, filepath.Join(os.Getenv("TIDEPOOL_CONFIG_DIRECTORY"), "dataservices"), "TIDEPOOL")
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create config loader")
	}
	s.configLoader = configLoader

	return nil
}

func (s *Standard) initializeLogger() error {
	loggerConfig := &log.Config{}
	if err := s.configLoader.Load("logger", loggerConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load logger config")
	}

	logger, err := log.NewStandard(s.versionReporter, loggerConfig)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create logger")
	}
	s.logger = logger

	s.logger.Warn(fmt.Sprintf("Logger level is %s", loggerConfig.Level))
	return nil
}

func (s *Standard) initializeDataFactory() error {
	s.logger.Debug("Creating data factory")

	dataFactory, err := factory.NewStandard()
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create data factory")
	}
	s.dataFactory = dataFactory

	return nil
}

func (s *Standard) initializeDataStore() error {
	s.logger.Debug("Loading data store config")

	dataStoreConfig := &commonMongo.Config{}
	if err := s.configLoader.Load("data_store", dataStoreConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load data store config")
	}
	dataStoreConfig.Collection = "deviceData"

	s.logger.Debug("Creating data store")

	dataStore, err := mongo.New(s.logger, dataStoreConfig)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create data store")
	}
	s.dataStore = dataStore

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.logger.Debug("Creating truncate data deduplicator factory")

	truncateDeduplicatorFactory, err := truncate.NewFactory()
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create truncate data deduplicator factory")
	}

	s.logger.Debug("Creating data deduplicator factory")

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
	}

	dataDeduplicatorFactory, err := delegate.NewFactory(factories)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = dataDeduplicatorFactory

	return nil
}

func (s *Standard) initializeUserServicesClient() error {
	s.logger.Debug("Loading user services client config")

	userServicesClientConfig := &client.Config{}
	if err := s.configLoader.Load("userservices_client", userServicesClientConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load user services client config")
	}

	s.logger.Debug("Creating user services client")

	userServicesClient, err := client.NewStandard(s.logger, userServicesClientConfig)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create user services client")
	}
	s.userServicesClient = userServicesClient

	s.logger.Debug("Starting user services client")
	if err = s.userServicesClient.Start(); err != nil {
		return app.ExtError(err, "dataservices", "unable to start user services client")
	}

	return nil
}

func (s *Standard) initializeDataServicesAPI() error {
	s.logger.Debug("Creating data services api")

	dataServicesAPI, err := api.NewStandard(s.versionReporter, s.environmentReporter, s.logger, s.dataFactory, s.dataStore, s.dataDeduplicatorFactory, s.userServicesClient, v1.Routes())
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create data services api")
	}
	s.dataServicesAPI = dataServicesAPI

	return nil
}

func (s *Standard) initializeDataServicesServer() error {
	s.logger.Debug("Loading data services server config")

	dataServicesServerConfig := &server.Config{}
	if err := s.configLoader.Load("dataservices_server", dataServicesServerConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load data services server config")
	}

	s.logger.Debug("Creating data services server")

	dataServicesServer, err := server.NewStandard(s.logger, s.dataServicesAPI, dataServicesServerConfig)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create data services server")
	}
	s.dataServicesServer = dataServicesServer

	return nil
}
