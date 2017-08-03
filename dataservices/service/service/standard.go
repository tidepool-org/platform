package service

import (
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/factory"
	dataMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	"github.com/tidepool-org/platform/dataservices/service/api"
	"github.com/tidepool-org/platform/dataservices/service/api/v1"
	"github.com/tidepool-org/platform/errors"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/service/server"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	taskMongo "github.com/tidepool-org/platform/task/store/mongo"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Standard struct {
	*service.Service
	metricServicesClient    *metricservicesClient.Standard
	userServicesClient      *userservicesClient.Standard
	dataFactory             *factory.Standard
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               *dataMongo.Store
	taskStore               *taskMongo.Store
	dataServicesAPI         *api.Standard
	dataServicesServer      *server.Standard
}

func NewStandard() (*Standard, error) {
	svc, err := service.New("dataservices", "TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Standard{
		Service: svc,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.Service.Initialize(); err != nil {
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
	if err := s.initializeDataDeduplicatorFactory(); err != nil {
		return err
	}
	if err := s.initializeDataStore(); err != nil {
		return err
	}
	if err := s.initializeTaskStore(); err != nil {
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
	if s.taskStore != nil {
		s.taskStore.Close()
		s.taskStore = nil
	}
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
	s.dataDeduplicatorFactory = nil
	s.dataFactory = nil
	if s.userServicesClient != nil {
		s.userServicesClient.Close()
		s.userServicesClient = nil
	}
	s.metricServicesClient = nil

	s.Service.Terminate()
}

func (s *Standard) Run() error {
	if s.dataServicesServer == nil {
		return errors.New("service", "service not initialized")
	}

	return s.dataServicesServer.Serve()
}

func (s *Standard) initializeMetricServicesClient() error {
	s.Logger().Debug("Loading metric services client config")

	metricServicesClientConfig := metricservicesClient.NewConfig()
	if err := metricServicesClientConfig.Load(s.ConfigReporter().WithScopes("metricservices", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load metric services client config")
	}

	s.Logger().Debug("Creating metric services client")

	metricServicesClient, err := metricservicesClient.NewStandard(s.VersionReporter(), s.Name(), metricServicesClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create metric services client")
	}
	s.metricServicesClient = metricServicesClient

	return nil
}

func (s *Standard) initializeUserServicesClient() error {
	s.Logger().Debug("Loading user services client config")

	userServicesClientConfig := userservicesClient.NewConfig()
	if err := userServicesClientConfig.Load(s.ConfigReporter().WithScopes("userservices", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load user services client config")
	}

	s.Logger().Debug("Creating user services client")

	userServicesClient, err := userservicesClient.NewStandard(s.Logger(), s.Name(), userServicesClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user services client")
	}
	s.userServicesClient = userServicesClient

	s.Logger().Debug("Starting user services client")
	if err = s.userServicesClient.Start(); err != nil {
		return errors.Wrap(err, "service", "unable to start user services client")
	}

	return nil
}

func (s *Standard) initializeDataFactory() error {
	s.Logger().Debug("Creating data factory")

	dataFactory, err := factory.NewStandard()
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data factory")
	}
	s.dataFactory = dataFactory

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating truncate data deduplicator factory")

	truncateDeduplicatorFactory, err := deduplicator.NewTruncateFactory()
	if err != nil {
		return errors.Wrap(err, "service", "unable to create truncate data deduplicator factory")
	}

	s.Logger().Debug("Creating hash-deactivate-old data deduplicator factory")

	hashDeactivateOldDeduplicatorFactory, err := deduplicator.NewHashDeactivateOldFactory()
	if err != nil {
		return errors.Wrap(err, "service", "unable to create hash-deactivate-old data deduplicator factory")
	}

	s.Logger().Debug("Creating data deduplicator factory")

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
		hashDeactivateOldDeduplicatorFactory,
	}

	dataDeduplicatorFactory, err := deduplicator.NewDelegateFactory(factories)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = dataDeduplicatorFactory

	return nil
}

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store config")

	dataStoreConfig := baseMongo.NewConfig()
	if err := dataStoreConfig.Load(s.ConfigReporter().WithScopes("data", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load data store config")
	}
	dataStoreConfig.Collection = "deviceData"

	s.Logger().Debug("Creating data store")

	dataStore, err := dataMongo.New(s.Logger(), dataStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data store")
	}
	s.dataStore = dataStore

	return nil
}

func (s *Standard) initializeTaskStore() error {
	s.Logger().Debug("Loading task store config")

	taskStoreConfig := baseMongo.NewConfig()
	if err := taskStoreConfig.Load(s.ConfigReporter().WithScopes("task", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load task store config")
	}
	taskStoreConfig.Collection = "syncTasks"

	s.Logger().Debug("Creating task store")

	taskStore, err := taskMongo.New(s.Logger(), taskStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create task store")
	}
	s.taskStore = taskStore

	return nil
}

func (s *Standard) initializeDataServicesAPI() error {
	s.Logger().Debug("Creating data services api")

	dataServicesAPI, err := api.NewStandard(s.VersionReporter(), s.Logger(),
		s.metricServicesClient, s.userServicesClient,
		s.dataFactory, s.dataDeduplicatorFactory,
		s.dataStore, s.taskStore)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data services api")
	}
	s.dataServicesAPI = dataServicesAPI

	s.Logger().Debug("Initializing data services api middleware")

	if err = s.dataServicesAPI.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "service", "unable to initialize data services api middleware")
	}

	s.Logger().Debug("Configuring data services api middleware headers")

	s.dataServicesAPI.HeaderMiddleware().AddHeaderFieldFunc(
		dataservicesClient.TidepoolAuthenticationTokenHeaderName, middleware.NewMD5FieldFunc("authenticationTokenMD5"))

	s.Logger().Debug("Initializing data services api router")

	if err = s.dataServicesAPI.InitializeRouter(v1.Routes()); err != nil {
		return errors.Wrap(err, "service", "unable to initialize data services api router")
	}

	return nil
}

func (s *Standard) initializeDataServicesServer() error {
	s.Logger().Debug("Loading data services server config")

	dataServicesServerConfig := server.NewConfig()
	if err := dataServicesServerConfig.Load(s.ConfigReporter().WithScopes(s.Name(), "server")); err != nil {
		return errors.Wrap(err, "service", "unable to load data services server config")
	}

	s.Logger().Debug("Creating data services server")

	dataServicesServer, err := server.NewStandard(s.Logger(), s.dataServicesAPI, dataServicesServerConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data services server")
	}
	s.dataServicesServer = dataServicesServer

	return nil
}
