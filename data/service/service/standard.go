package service

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/mdblp/go-db/mongo"
	"github.com/sirupsen/logrus"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data/service/api"
	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const DEFAULT_MINIMAL_YEAR = 2015

type Standard struct {
	*service.Service
	permissionClient *permissionClient.Client
	dataStore        *dataStoreMongo.Stores
	api              *api.Standard
	server           *server.Standard
}

var logrusLogger = logrus.New()

func NewStandard() *Standard {
	return &Standard{
		Service: service.NewService(),
	}
}

func (s *Standard) Initialize(provider application.Provider) error {
	if err := s.Service.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializePermissionClient(); err != nil {
		return err
	}
	if err := s.initializeDataStore(); err != nil {
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

	if s.dataStore != nil {
		s.dataStore.Terminate(context.Background())
		s.dataStore = nil
		if s.dataStore.BucketStore != nil {
			s.dataStore.BucketStore.Close()
			s.dataStore = nil
		}
	}
	s.permissionClient = nil

	s.Service.Terminate()
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
	s.Logger().Debug("Creating permission client")
	clnt := permissionClient.New()
	s.permissionClient = clnt

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

	str, err := dataStoreMongo.NewStores(cfg, mongoDbReadConfig, logrusLogger, migrateConfig, getMinimalYearSupportedForData())
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	s.dataStore = str

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.permissionClient, s.dataStore)
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

func getMinimalYearSupportedForData() int {
	s, err := getenvStr("MINIMAL_YEAR_SUPPORTED_FOR_DATA")
	if err != nil {
		logrusLogger.Warnf("environment variable MINIMAL_YEAR_SUPPORTED_FOR_DATA not exported, set %d by default", DEFAULT_MINIMAL_YEAR)
		return DEFAULT_MINIMAL_YEAR
	}
	if s != "" {
		intVar, err := strconv.Atoi(s)
		if err != nil {
			logrusLogger.Warnf("environment variable MINIMAL_YEAR_SUPPORTED_FOR_DATA=%s is not an integer, set to %d by default", s, DEFAULT_MINIMAL_YEAR)
			return DEFAULT_MINIMAL_YEAR
		}
		return intVar
	}
	logrusLogger.Warnf("environment variable MINIMAL_YEAR_SUPPORTED_FOR_DATA is empty, set to %d by default", DEFAULT_MINIMAL_YEAR)
	return DEFAULT_MINIMAL_YEAR
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
