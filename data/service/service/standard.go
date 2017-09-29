package service

import (
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/service/api"
	"github.com/tidepool-org/platform/data/service/api/v1"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataStoreDEPRECATEDMongo "github.com/tidepool-org/platform/data/storeDEPRECATED/mongo"
	"github.com/tidepool-org/platform/errors"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
	userClient "github.com/tidepool-org/platform/user/client"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient            *metricClient.Client
	userClient              *userClient.Client
	dataFactory             *factory.Standard
	dataDeduplicatorFactory deduplicator.Factory
	dataStoreDEPRECATED     *dataStoreDEPRECATEDMongo.Store
	dataStore               *dataStoreMongo.Store
	syncTaskStore           *syncTaskMongo.Store
	dataClient              *Client
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
	if err := s.initializeDataStoreDEPRECATED(); err != nil {
		return err
	}
	if err := s.initializeDataStore(); err != nil {
		return err
	}
	if err := s.initializeSyncTaskStore(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	return s.initializeServer()
}

func (s *Standard) Terminate() {
	s.server = nil
	s.api = nil
	s.dataClient = nil
	if s.syncTaskStore != nil {
		s.syncTaskStore.Close()
		s.syncTaskStore = nil
	}
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
	if s.dataStoreDEPRECATED != nil {
		s.dataStoreDEPRECATED.Close()
		s.dataStoreDEPRECATED = nil
	}
	s.dataDeduplicatorFactory = nil
	s.dataFactory = nil
	s.userClient = nil
	s.metricClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	return s.server.Serve()
}

func (s *Standard) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	cfg := platform.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("metric", "client")); err != nil {
		return errors.Wrap(err, "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	clnt, err := metricClient.New(cfg, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "unable to create metric client")
	}
	s.metricClient = clnt

	return nil
}

func (s *Standard) initializeUserClient() error {
	s.Logger().Debug("Loading user client config")

	cfg := platform.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("user", "client")); err != nil {
		return errors.Wrap(err, "unable to load user client config")
	}

	s.Logger().Debug("Creating user client")

	clnt, err := userClient.New(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	s.userClient = clnt

	return nil
}

func (s *Standard) initializeDataFactory() error {
	s.Logger().Debug("Creating data factory")

	dataFactory, err := factory.NewStandard()
	if err != nil {
		return errors.Wrap(err, "unable to create data factory")
	}
	s.dataFactory = dataFactory

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating truncate data deduplicator factory")

	truncateDeduplicatorFactory, err := deduplicator.NewTruncateFactory()
	if err != nil {
		return errors.Wrap(err, "unable to create truncate data deduplicator factory")
	}

	s.Logger().Debug("Creating hash-deactivate-old data deduplicator factory")

	hashDeactivateOldDeduplicatorFactory, err := deduplicator.NewHashDeactivateOldFactory()
	if err != nil {
		return errors.Wrap(err, "unable to create hash-deactivate-old data deduplicator factory")
	}

	s.Logger().Debug("Creating continuous data deduplicator factory")

	continuousDeduplicatorFactory, err := deduplicator.NewContinuousFactory()
	if err != nil {
		return errors.Wrap(err, "unable to create continuous data deduplicator factory")
	}

	s.Logger().Debug("Creating data deduplicator factory")

	factories := []deduplicator.Factory{
		truncateDeduplicatorFactory,
		hashDeactivateOldDeduplicatorFactory,
		continuousDeduplicatorFactory,
	}

	dataDeduplicatorFactory, err := deduplicator.NewDelegateFactory(factories)
	if err != nil {
		return errors.Wrap(err, "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = dataDeduplicatorFactory

	return nil
}

func (s *Standard) initializeDataStoreDEPRECATED() error {
	s.Logger().Debug("Loading data store DEPRECATED config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data store DEPRECATED config")
	}

	s.Logger().Debug("Creating data store")

	str, err := dataStoreDEPRECATEDMongo.New(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create data store DEPRECATED")
	}
	s.dataStoreDEPRECATED = str

	return nil
}

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data store config")
	}

	s.Logger().Debug("Creating data store")

	str, err := dataStoreMongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	s.dataStore = str

	return nil
}

func (s *Standard) initializeSyncTaskStore() error {
	s.Logger().Debug("Loading sync task store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("sync_task", "store")); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}

	s.Logger().Debug("Creating sync task store")

	str, err := syncTaskMongo.New(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create sync task store")
	}
	s.syncTaskStore = str

	return nil
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := NewClient(s.dataStore)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.metricClient, s.userClient,
		s.dataFactory, s.dataDeduplicatorFactory, s.dataStore,
		s.dataStoreDEPRECATED, s.syncTaskStore, s.dataClient)
	if err != nil {
		return errors.Wrap(err, "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(v1.Routes()); err != nil {
		return errors.Wrap(err, "unable to initialize api router")
	}

	return nil
}

func (s *Standard) initializeServer() error {
	s.Logger().Debug("Loading server config")

	serverConfig := server.NewConfig()
	if err := serverConfig.Load(s.ConfigReporter().WithScopes("server")); err != nil {
		return errors.Wrap(err, "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	newServer, err := server.NewStandard(serverConfig, s.Logger(), s.api)
	if err != nil {
		return errors.Wrap(err, "unable to create server")
	}
	s.server = newServer

	return nil
}
