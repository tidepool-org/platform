package service

import (
	"context"
	"os"
	"strings"

	"github.com/mdblp/go-common/clients/mongo"
	"github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/application"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	"github.com/tidepool-org/platform/data/service/api"
	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/errors"
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
	permissionClient        *permissionClient.Client
	dataDeduplicatorFactory *dataDeduplicatorFactory.Factory
	dataStore               *dataStoreMongo.Stores
	syncTaskStore           *syncTaskMongo.Store
	dataClient              *Client
	api                     *api.Standard
	server                  *server.Standard
}

var logrusLogger = logrus.New()

func NewStandard() *Standard {
	return &Standard{
		DEPRECATEDService: service.NewDEPRECATEDService(),
	}
}

func (s *Standard) Initialize(provider application.Provider) error {
	if err := s.DEPRECATEDService.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializePermissionClient(); err != nil {
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
		s.syncTaskStore.Terminate(context.Background())
		s.syncTaskStore = nil
	}

	if s.dataStore != nil {
		s.dataStore.Terminate(context.Background())
		s.dataStore = nil
		if s.dataStore.BucketStore != nil {
			s.dataStore.BucketStore.Close()
			s.dataStore = nil
		}
	}
	s.dataDeduplicatorFactory = nil
	s.permissionClient = nil

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

func (s *Standard) initializePermissionClient() error {
	s.Logger().Debug("Loading permission client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("permission", "client")
	if err := cfg.Load(reporter); err != nil {
		return errors.Wrap(err, "unable to load permission client config")
	}

	s.Logger().Debug("Creating permission client")
	clnt, err := permissionClient.New(cfg)
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

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data store config")
	}
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data store")

	// Temporary hack
	// new logger configuration required due to go common
	logrusLogger.Out = os.Stdout
	logrusLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	envLogLevel, _ := getenvStr("LOG_LEVEL")
	logLevel, err := logrus.ParseLevel(envLogLevel)
	if err != nil {
		logLevel = logrus.WarnLevel
	}

	logrusLogger.SetLevel(logLevel)
	// report method name
	logrusLogger.SetReportCaller(true)
	var mongoDbReadConfig = &mongo.Config{}
	mongoDbReadConfig.FromEnv()
	mongoDbReadConfig.Database = "data_read"

	migrateConfig := dataStoreMongo.BucketMigrationConfig{
		DataTypesArchived:     getArchivedDataTypesEnv(),
		DataTypesBucketed:     getBucketsDataTypesEnv(),
		DataTypesKeptInLegacy: getKeptInLegacyDataTypesEnv(),
	}

	str, err := dataStoreMongo.NewStores(cfg, mongoDbReadConfig, s.Logger(), logrusLogger, migrateConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	s.dataStore = str

	return nil
}

func (s *Standard) initializeSyncTaskStore() error {
	s.Logger().Debug("Loading sync task store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("sync_task", "store")); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}

	s.Logger().Debug("Creating sync task store")

	str, err := syncTaskMongo.NewStore(cfg)
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

	newAPI, err := api.NewStandard(s, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient)
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

func getenvStr(key string) (string, error) {
	ErrEnvVarEmpty := errors.New("getenv: environment variable empty")
	v := os.Getenv(key)
	if v == "" {
		logrusLogger.Debug("environment variable empty")
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func getArchivedDataTypesEnv() []string {
	s, err := getenvStr("ARCHIVED_DATA_TYPES")
	if err != nil {
		logrusLogger.Warn("environment variable ARCHIVED_DATA_TYPES not exported, set empty by default")
		return []string{}
	}
	if s != "" {
		dataTypes := strings.Split(s, ",")
		return dataTypes
	}
	return []string{}
}

func getBucketsDataTypesEnv() []string {
	s, err := getenvStr("BUCKETED_DATA_TYPES")
	if err != nil {
		logrusLogger.Warn("environment variable BUCKETED_DATA_TYPES not exported, set empty by default")
		return []string{}
	}
	if s != "" {
		dataTypes := strings.Split(s, ",")
		return dataTypes
	}
	return []string{}
}

func getKeptInLegacyDataTypesEnv() []string {
	s, err := getenvStr("KEPT_IN_LEGACY_DATA_TYPES")
	if err != nil {
		logrusLogger.Warn("environment variable KEPT_IN_LEGACY_DATA_TYPES not exported, set empty by default")
		return []string{}
	}
	if s != "" {
		dataTypes := strings.Split(s, ",")
		return dataTypes
	}

	return []string{}
}
