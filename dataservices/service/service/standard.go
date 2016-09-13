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
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/delegate"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/dataservices/service/api"
	"github.com/tidepool-org/platform/dataservices/service/api/v1"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	commonMongo "github.com/tidepool-org/platform/store/mongo"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	*service.Standard
	metricServicesClient    *metricservicesClient.Standard
	userServicesClient      *userservicesClient.Standard
	dataFactory             *factory.Standard
	dataStore               *mongo.Store
	dataDeduplicatorFactory deduplicator.Factory
	dataServicesAPI         *api.Standard
	dataServicesServer      *server.Standard
}

func NewStandard() (*Standard, error) {
	standard, err := service.NewStandard("dataservices", "TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard: standard,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.Standard.Initialize(); err != nil {
		return err
	}

	if err := s.initializeMetricServicesClient(); err != nil {
		return err
	}
	if err := s.initializeUserServicesClient(); err != nil {
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
	if err := s.initializeDataServicesAPI(); err != nil {
		return err
	}
	if err := s.initializeDataServicesServer(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.dataServicesServer = nil
	s.dataServicesAPI = nil
	s.dataDeduplicatorFactory = nil
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
	s.dataFactory = nil
	if s.userServicesClient != nil {
		s.userServicesClient.Close()
		s.userServicesClient = nil
	}
	s.metricServicesClient = nil

	s.Standard.Terminate()
}

func (s *Standard) Run() error {
	if s.dataServicesServer == nil {
		return app.Error("service", "service not initialized")
	}

	return s.dataServicesServer.Serve()
}

func (s *Standard) initializeMetricServicesClient() error {
	s.Logger().Debug("Loading metric services client config")

	metricServicesClientConfig := &metricservicesClient.Config{}
	if err := s.ConfigLoader().Load("metricservices_client", metricServicesClientConfig); err != nil {
		return app.ExtError(err, "service", "unable to load metric services client config")
	}

	s.Logger().Debug("Creating metric services client")

	metricServicesClient, err := metricservicesClient.NewStandard(s.VersionReporter(), s.Name(), metricServicesClientConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create metric services client")
	}
	s.metricServicesClient = metricServicesClient

	return nil
}

func (s *Standard) initializeUserServicesClient() error {
	s.Logger().Debug("Loading user services client config")

	userServicesClientConfig := &userservicesClient.Config{}
	if err := s.ConfigLoader().Load("userservices_client", userServicesClientConfig); err != nil {
		return app.ExtError(err, "service", "unable to load user services client config")
	}

	s.Logger().Debug("Creating user services client")

	userServicesClient, err := userservicesClient.NewStandard(s.Logger(), s.Name(), userServicesClientConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create user services client")
	}
	s.userServicesClient = userServicesClient

	s.Logger().Debug("Starting user services client")
	if err = s.userServicesClient.Start(); err != nil {
		return app.ExtError(err, "service", "unable to start user services client")
	}

	return nil
}

func (s *Standard) initializeDataFactory() error {
	s.Logger().Debug("Creating data factory")

	dataFactory, err := factory.NewStandard()
	if err != nil {
		return app.ExtError(err, "service", "unable to create data factory")
	}
	s.dataFactory = dataFactory

	return nil
}

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store config")

	dataStoreConfig := &commonMongo.Config{}
	if err := s.ConfigLoader().Load("data_store", dataStoreConfig); err != nil {
		return app.ExtError(err, "service", "unable to load data store config")
	}
	dataStoreConfig.Collection = "deviceData"

	s.Logger().Debug("Creating data store")

	dataStore, err := mongo.New(s.Logger(), dataStoreConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create data store")
	}
	s.dataStore = dataStore

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating truncate data deduplicator factory")

	truncateDeduplicatorFactory, err := truncate.NewFactory()
	if err != nil {
		return app.ExtError(err, "service", "unable to create truncate data deduplicator factory")
	}

	s.Logger().Debug("Creating data deduplicator factory")

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
	}

	dataDeduplicatorFactory, err := delegate.NewFactory(factories)
	if err != nil {
		return app.ExtError(err, "service", "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = dataDeduplicatorFactory

	return nil
}

func (s *Standard) initializeDataServicesAPI() error {
	s.Logger().Debug("Creating data services api")

	dataServicesAPI, err := api.NewStandard(s.VersionReporter(), s.EnvironmentReporter(), s.Logger(), s.metricServicesClient, s.userServicesClient, s.dataFactory, s.dataStore, s.dataDeduplicatorFactory)
	if err != nil {
		return app.ExtError(err, "service", "unable to create data services api")
	}
	s.dataServicesAPI = dataServicesAPI

	s.Logger().Debug("Initializing data services api middleware")

	if err = s.dataServicesAPI.InitializeMiddleware(); err != nil {
		return app.ExtError(err, "service", "unable to initialize data services api middleware")
	}

	s.Logger().Debug("Initializing data services api router")

	if err = s.dataServicesAPI.InitializeRouter(v1.Routes()); err != nil {
		return app.ExtError(err, "service", "unable to initialize data services api router")
	}

	return nil
}

func (s *Standard) initializeDataServicesServer() error {
	s.Logger().Debug("Loading data services server config")

	dataServicesServerConfig := &server.Config{}
	if err := s.ConfigLoader().Load("dataservices_server", dataServicesServerConfig); err != nil {
		return app.ExtError(err, "service", "unable to load data services server config")
	}

	s.Logger().Debug("Creating data services server")

	dataServicesServer, err := server.NewStandard(s.Logger(), s.dataServicesAPI, dataServicesServerConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create data services server")
	}
	s.dataServicesServer = dataServicesServer

	return nil
}
