package service

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/service/api"
	"github.com/tidepool-org/platform/data/service/api/v1"
	dataMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/errors"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient            metricClient.Client
	userClient              userClient.Client
	dataFactory             *factory.Standard
	dataDeduplicatorFactory deduplicator.Factory
	dataStore               *dataMongo.Store
	syncTaskStore           *syncTaskMongo.Store
	api                     *api.Standard
	server                  *server.Standard
}

func NewStandard(prefix string) (*Standard, error) {
	svc, err := service.NewDEPRECATEDService(prefix)
	if err != nil {
		return nil, err
	}

	return &Standard{
		DEPRECATEDService: svc,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.DEPRECATEDService.Initialize(); err != nil {
		return err
	}

	if err := s.initializeMetricClient(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
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
	if err := s.initializeSyncTaskStore(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	if err := s.initializeServer(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.server = nil
	s.api = nil
	if s.syncTaskStore != nil {
		s.syncTaskStore.Close()
		s.syncTaskStore = nil
	}
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
	s.dataDeduplicatorFactory = nil
	s.dataFactory = nil
	s.userClient = nil
	s.metricClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service", "service not initialized")
	}

	return s.server.Serve()
}

func (s *Standard) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	metricClientConfig := client.NewConfig()
	if err := metricClientConfig.Load(s.ConfigReporter().WithScopes("metric", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	metricClient, err := metricClient.NewClient(metricClientConfig, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create metric client")
	}
	s.metricClient = metricClient

	return nil
}

func (s *Standard) initializeUserClient() error {
	s.Logger().Debug("Loading user client config")

	userClientConfig := client.NewConfig()
	if err := userClientConfig.Load(s.ConfigReporter().WithScopes("user", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load user client config")
	}

	s.Logger().Debug("Creating user client")

	userClient, err := userClient.NewClient(userClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user client")
	}
	s.userClient = userClient

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

	s.Logger().Debug("Creating data store")

	dataStore, err := dataMongo.New(s.Logger(), dataStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data store")
	}
	s.dataStore = dataStore

	return nil
}

func (s *Standard) initializeSyncTaskStore() error {
	s.Logger().Debug("Loading sync task store config")

	syncTaskStoreConfig := baseMongo.NewConfig()
	if err := syncTaskStoreConfig.Load(s.ConfigReporter().WithScopes("sync_task", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load sync task store config")
	}

	s.Logger().Debug("Creating sync task store")

	syncTaskStore, err := syncTaskMongo.New(s.Logger(), syncTaskStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create sync task store")
	}
	s.syncTaskStore = syncTaskStore

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s.VersionReporter(), s.Logger(),
		s.AuthClient(), s.metricClient, s.userClient,
		s.dataFactory, s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "service", "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(v1.Routes()); err != nil {
		return errors.Wrap(err, "service", "unable to initialize api router")
	}

	return nil
}

func (s *Standard) initializeServer() error {
	s.Logger().Debug("Loading server config")

	serverConfig := server.NewConfig()
	if err := serverConfig.Load(s.ConfigReporter().WithScopes("server")); err != nil {
		return errors.Wrap(err, "service", "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	newServer, err := server.NewStandard(s.Logger(), s.api, serverConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create server")
	}
	s.server = newServer

	return nil
}
