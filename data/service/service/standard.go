package service

import (
	"github.com/tidepool-org/platform/application"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	"github.com/tidepool-org/platform/data/service/api"
	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataSourceServiceClient "github.com/tidepool-org/platform/data/source/service/client"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataSourceStoreStructuredMongo "github.com/tidepool-org/platform/data/source/store/structured/mongo"
	dataStoreDEPRECATEDMongo "github.com/tidepool-org/platform/data/storeDEPRECATED/mongo"
	"github.com/tidepool-org/platform/errors"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient              *metricClient.Client
	permissionClient          *permissionClient.Client
	dataDeduplicatorFactory   *dataDeduplicatorFactory.Factory
	dataStoreDEPRECATED       *dataStoreDEPRECATEDMongo.Store
	dataSourceStructuredStore *dataSourceStoreStructuredMongo.Store
	syncTaskStore             *syncTaskMongo.Store
	dataClient                *Client
	dataSourceClient          *dataSourceServiceClient.Client
	api                       *api.Standard
	server                    *server.Standard
}

func NewStandard() *Standard {
	return &Standard{
		DEPRECATEDService: service.NewDEPRECATEDService(),
	}
}

func (s *Standard) Initialize(provider application.Provider) error {
	if err := s.DEPRECATEDService.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeMetricClient(); err != nil {
		return err
	}
	if err := s.initializePermissionClient(); err != nil {
		return err
	}
	if err := s.initializeDataDeduplicatorFactory(); err != nil {
		return err
	}
	if err := s.initializeDataStoreDEPRECATED(); err != nil {
		return err
	}
	if err := s.initializeDataSourceStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeSyncTaskStore(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
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
	if s.dataSourceStructuredStore != nil {
		s.dataSourceStructuredStore.Close()
		s.dataSourceStructuredStore = nil
	}
	if s.dataStoreDEPRECATED != nil {
		s.dataStoreDEPRECATED.Close()
		s.dataStoreDEPRECATED = nil
	}
	s.dataDeduplicatorFactory = nil
	s.permissionClient = nil
	s.metricClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	return s.server.Serve()
}

func (s *Standard) PermissionClient() permission.Client {
	return s.permissionClient
}

func (s *Standard) DataSourceStructuredStore() dataSourceStoreStructured.Store {
	return s.dataSourceStructuredStore
}

func (s *Standard) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	if err := cfg.Load(s.ConfigReporter().WithScopes("metric", "client")); err != nil {
		return errors.Wrap(err, "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	clnt, err := metricClient.New(cfg, platform.AuthorizeAsUser, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "unable to create metric client")
	}
	s.metricClient = clnt

	return nil
}

func (s *Standard) initializePermissionClient() error {
	s.Logger().Debug("Loading permission client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	if err := cfg.Load(s.ConfigReporter().WithScopes("permission", "client")); err != nil {
		return errors.Wrap(err, "unable to load permission client config")
	}

	s.Logger().Debug("Creating permission client")

	clnt, err := permissionClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create permission client")
	}
	s.permissionClient = clnt

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating device deactivate hash deduplicator")

	deviceDeactivateHashDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceDeactivateHash()
	if err != nil {
		return errors.Wrap(err, "unable to create device deactivate hash deduplicator")
	}

	s.Logger().Debug("Creating device truncate data set deduplicator")

	deviceTruncateDataSetDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet()
	if err != nil {
		return errors.Wrap(err, "unable to create device truncate data set deduplicator")
	}

	s.Logger().Debug("Creating data set delete origin deduplicator")

	dataSetDeleteOriginDeduplicator, err := dataDeduplicatorDeduplicator.NewDataSetDeleteOrigin()
	if err != nil {
		return errors.Wrap(err, "unable to create data set delete origin deduplicator")
	}

	s.Logger().Debug("Creating none deduplicator")

	noneDeduplicator, err := dataDeduplicatorDeduplicator.NewNone()
	if err != nil {
		return errors.Wrap(err, "unable to create none deduplicator")
	}

	s.Logger().Debug("Creating data deduplicator factory")

	deduplicators := []dataDeduplicatorFactory.Deduplicator{
		deviceDeactivateHashDeduplicator,
		deviceTruncateDataSetDeduplicator,
		dataSetDeleteOriginDeduplicator,
		noneDeduplicator,
	}

	factory, err := dataDeduplicatorFactory.New(deduplicators)
	if err != nil {
		return errors.Wrap(err, "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = factory

	return nil
}

func (s *Standard) initializeDataStoreDEPRECATED() error {
	s.Logger().Debug("Loading data store DEPRECATED config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data store DEPRECATED config")
	}

	s.Logger().Debug("Creating data store")

	str, err := dataStoreDEPRECATEDMongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create data store DEPRECATED")
	}
	s.dataStoreDEPRECATED = str

	s.Logger().Debug("Ensuring data store DEPRECATED indexes")

	err = s.dataStoreDEPRECATED.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data store DEPRECATED indexes")
	}

	return nil
}

func (s *Standard) initializeDataSourceStructuredStore() error {
	s.Logger().Debug("Loading data source structured store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("data_source", "store")); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data source structured store")

	str, err := dataSourceStoreStructuredMongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create data source structured store")
	}
	s.dataSourceStructuredStore = str

	s.Logger().Debug("Ensuring data source structured store indexes")

	err = s.dataSourceStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data source structured store indexes")
	}

	return nil
}

func (s *Standard) initializeSyncTaskStore() error {
	s.Logger().Debug("Loading sync task store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("sync_task", "store")); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}

	s.Logger().Debug("Creating sync task store")

	str, err := syncTaskMongo.NewStore(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create sync task store")
	}
	s.syncTaskStore = str

	return nil
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := NewClient(s.dataStoreDEPRECATED)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Standard) initializeDataSourceClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := dataSourceServiceClient.New(s)
	if err != nil {
		return errors.Wrap(err, "unable to create source data client")
	}
	s.dataSourceClient = clnt

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStoreDEPRECATED, s.syncTaskStore, s.dataClient, s.dataSourceClient)
	if err != nil {
		return errors.Wrap(err, "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(dataServiceApiV1.Routes()); err != nil {
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
