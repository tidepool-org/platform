package service

import (
	"context"
	"log"
	"os"

	"github.com/IBM/sarama"
	eventsCommon "github.com/tidepool-org/go-common/events"

	abbottClient "github.com/tidepool-org/platform-plugin-abbott/abbott/client"
	abbottProvider "github.com/tidepool-org/platform-plugin-abbott/abbott/provider"
	abbottWork "github.com/tidepool-org/platform-plugin-abbott/abbott/work"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/clinics"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	dataEvents "github.com/tidepool-org/platform/data/events"
	dataRawService "github.com/tidepool-org/platform/data/raw/service"
	dataRawStoreStructuredMongo "github.com/tidepool-org/platform/data/raw/store/structured/mongo"
	"github.com/tidepool-org/platform/data/service/api"
	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataSourceServiceClient "github.com/tidepool-org/platform/data/source/service/client"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataSourceStoreStructuredMongo "github.com/tidepool-org/platform/data/source/store/structured/mongo"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/events"
	logInternal "github.com/tidepool-org/platform/log"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
	workService "github.com/tidepool-org/platform/work/service"
	workStoreStructuredMongo "github.com/tidepool-org/platform/work/store/structured/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient              *metricClient.Client
	permissionClient          *permissionClient.Client
	dataStore                 *dataStoreMongo.Store
	dataRawStructuredStore    *dataRawStoreStructuredMongo.Store
	dataSourceStructuredStore *dataSourceStoreStructuredMongo.Store
	syncTaskStore             *syncTaskMongo.Store
	workStructuredStore       *workStoreStructuredMongo.Store
	dataDeduplicatorFactory   *dataDeduplicatorFactory.Factory
	clinicsClient             *clinics.Client
	dataClient                *Client
	dataRawClient             *dataRawService.Client
	dataSourceClient          *dataSourceServiceClient.Client
	workClient                *workService.Client
	abbottClient              *abbottClient.Client
	workCoordinator           *workService.Coordinator
	userEventsHandler         events.Runner
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
	if err := s.initializeDataStore(); err != nil {
		return err
	}
	if err := s.initializeDataRawStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeDataSourceStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeSyncTaskStore(); err != nil {
		return err
	}
	if err := s.initializeWorkStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeDataDeduplicatorFactory(); err != nil {
		return err
	}
	if err := s.initializeClinicsClient(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDataRawClient(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
		return err
	}
	if err := s.initializeWorkClient(); err != nil {
		return err
	}
	if err := s.initializeAbbottClient(); err != nil {
		return err
	}
	if err := s.initializeWorkCoordinator(); err != nil {
		return err
	}
	if err := s.initializeUserEventsHandler(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	return s.initializeServer()
}

func (s *Standard) Terminate() {
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			s.Logger().Errorf("Error while terminating the the server: %v", err)
		}
		s.server = nil
	}
	s.api = nil
	if s.userEventsHandler != nil {
		s.Logger().Info("Terminating the userEventsHandler")
		if err := s.userEventsHandler.Terminate(); err != nil {
			s.Logger().Errorf("Error while terminating the userEventsHandler: %v", err)
		}
		s.userEventsHandler = nil
	}
	if s.workCoordinator != nil {
		s.workCoordinator.Stop()
		s.workCoordinator = nil
	}
	s.abbottClient = nil
	s.workClient = nil
	s.dataSourceClient = nil
	s.dataRawClient = nil
	s.dataClient = nil
	s.clinicsClient = nil
	s.dataDeduplicatorFactory = nil
	if s.workStructuredStore != nil {
		s.workStructuredStore.Terminate(context.Background())
		s.workStructuredStore = nil
	}
	if s.syncTaskStore != nil {
		s.syncTaskStore.Terminate(context.Background())
		s.syncTaskStore = nil
	}
	if s.dataSourceStructuredStore != nil {
		s.dataSourceStructuredStore.Terminate(context.Background())
		s.dataSourceStructuredStore = nil
	}
	if s.dataRawStructuredStore != nil {
		s.dataRawStructuredStore.Terminate(context.Background())
		s.dataRawStructuredStore = nil
	}
	if s.dataStore != nil {
		s.dataStore.Terminate(context.Background())
		s.dataStore = nil
	}
	s.permissionClient = nil
	s.metricClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	errs := make(chan error)
	go func() {
		errs <- s.userEventsHandler.Run()
	}()
	go func() {
		errs <- s.server.Serve()
	}()

	return <-errs
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
	reporter := s.ConfigReporter().WithScopes("metric", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
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
	reporter := s.ConfigReporter().WithScopes("permission", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
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

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store DEPRECATED config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data store DEPRECATED config")
	}
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data store")

	str, err := dataStoreMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create data store DEPRECATED")
	}
	s.dataStore = str

	s.Logger().Debug("Ensuring data store DEPRECATED indexes")

	err = s.dataStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data store DEPRECATED indexes")
	}

	return nil
}

func (s *Standard) initializeDataRawStructuredStore() error {
	s.Logger().Debug("Loading data raw structured store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data raw structured store config")
	}

	s.Logger().Debug("Creating data raw structured store")

	str, err := dataRawStoreStructuredMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create data raw structured store")
	}
	s.dataRawStructuredStore = str

	s.Logger().Debug("Ensuring data raw structured store indexes")

	err = s.dataRawStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data raw structured store indexes")
	}

	return nil
}

func (s *Standard) initializeDataSourceStructuredStore() error {
	s.Logger().Debug("Loading data source structured store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data source structured store")

	str, err := dataSourceStoreStructuredMongo.NewStore(cfg)
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

func (s *Standard) initializeWorkStructuredStore() error {
	s.Logger().Debug("Loading work structured store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load work structured store config")
	}

	s.Logger().Debug("Creating work structured store")

	str, err := workStoreStructuredMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create work structured store")
	}
	s.workStructuredStore = str

	s.Logger().Debug("Ensuring work structured store indexes")

	err = s.workStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure work structured store indexes")
	}

	return nil
}

func (s *Standard) initializeClinicsClient() error {
	s.Logger().Debug("Creating clinics client")

	clnt, err := clinics.NewClient(s.AuthClient())
	if err != nil {
		return errors.Wrap(err, "unable to create clinics client")
	}
	s.clinicsClient = &clnt

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating device deactivate hash deduplicator")

	dataRepository := s.dataStore.NewDataRepository()
	dependencies := dataDeduplicatorDeduplicator.Dependencies{
		DataSetStore: dataRepository,
		DataStore:    dataRepository,
	}

	deviceDeactivateHashDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceDeactivateHash(dependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create device deactivate hash deduplicator")
	}

	s.Logger().Debug("Creating device truncate data set deduplicator")

	deviceTruncateDataSetDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet(dependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create device truncate data set deduplicator")
	}

	s.Logger().Debug("Creating data set delete origin deduplicator")

	dataSetDeleteOriginDeduplicator, err := dataDeduplicatorDeduplicator.NewDataSetDeleteOrigin(dependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create data set delete origin deduplicator")
	}

	s.Logger().Debug("Creating none deduplicator")

	noneDeduplicator, err := dataDeduplicatorDeduplicator.NewNone(dependencies)
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

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := NewClient(s.Logger(), s.dataStore, s.dataDeduplicatorFactory)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Standard) initializeDataRawClient() error {
	s.Logger().Debug("Creating data raw client")

	clnt, err := dataRawService.NewClient(s.dataRawStructuredStore)
	if err != nil {
		return errors.Wrap(err, "unable to create data raw client")
	}
	s.dataRawClient = clnt

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

func (s *Standard) initializeWorkClient() error {
	s.Logger().Debug("Creating work client")

	clnt, err := workService.NewClient(s.workStructuredStore)
	if err != nil {
		return errors.Wrap(err, "unable to create work client")
	}
	s.workClient = clnt

	return nil
}

func (s *Standard) initializeAbbottClient() error {
	s.Logger().Debug("Loading abbott provider")

	abbottProviderDependencies := abbottProvider.ProviderDependencies{
		ConfigReporter:        s.ConfigReporter().WithScopes("provider"),
		ProviderSessionClient: s.AuthClient(),
		DataSourceClient:      s.dataSourceClient,
		WorkClient:            s.workClient,
	}
	if prvdr, err := abbottProvider.NewProvider(abbottProviderDependencies); err != nil {
		s.Logger().Warn("Unable to create abbott provider")
	} else {
		s.Logger().Debug("Loading abbott client config")

		cfg := abbottClient.NewConfig()
		cfg.UserAgent = s.UserAgent()
		reporter := s.ConfigReporter().WithScopes("abbott", "client")
		if err = cfg.LoadFromConfigReporter(reporter); err != nil {
			return errors.Wrap(err, "unable to load abbott client config")
		}

		s.Logger().Debug("Creating abbott client")

		abbottClientDependencies := abbottClient.ClientDependencies{
			Config:            cfg,
			TokenSourceSource: prvdr,
		}
		clnt, clntErr := abbottClient.NewClient(abbottClientDependencies)
		if clntErr != nil {
			return errors.Wrap(clntErr, "unable to create abbott client")
		}
		s.abbottClient = clnt
	}

	return nil
}

func (s *Standard) initializeWorkCoordinator() error {
	s.Logger().Debug("Creating work coordinator")

	coordinator, err := workService.NewCoordinator(s.Logger(), s.AuthClient(), s.workClient)
	if err != nil {
		return errors.Wrap(err, "unable to create work coordinator")
	}
	s.workCoordinator = coordinator

	s.Logger().Debug("Creating abbott processors")

	abbottProcessorDependencies := abbottWork.ProcessorDependencies{
		DataDeduplicatorFactory: s.dataDeduplicatorFactory,
		DataSetClient:           s.dataClient,
		DataSourceClient:        s.dataSourceStructuredStore.NewDataSourcesRepository(),
		DataStoreClient:         s.dataStore,
		ProviderSessionClient:   s.AuthClient(),
		DataRawClient:           s.dataRawClient,
		AbbottClient:            s.abbottClient,
		WorkClient:              s.workClient,
	}
	abbottProcessors, err := abbottWork.NewProcessors(abbottProcessorDependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create abbott processors")
	}

	s.Logger().Debug("Registering abbott processors")

	if err = s.workCoordinator.RegisterProcessors(abbottProcessors); err != nil {
		return errors.Wrap(err, "unable to register abbott processors")
	}

	s.Logger().Debug("Starting work coordinator")

	s.workCoordinator.Start()

	return nil
}

func (s *Standard) initializeUserEventsHandler() error {
	s.Logger().Debug("Initializing user events handler")

	sarama.Logger = log.New(os.Stdout, "SARAMA ", log.LstdFlags|log.Lshortfile)

	ctx := logInternal.NewContextWithLogger(context.Background(), s.Logger())
	handler := dataEvents.NewUserDataDeletionHandler(ctx, s.dataStore, s.dataSourceStructuredStore)
	handlers := []eventsCommon.EventHandler{handler}
	runner := events.NewRunner(handlers)
	if err := runner.Initialize(); err != nil {
		return errors.Wrap(err, "unable to initialize user events handler runner")
	}
	s.userEventsHandler = runner

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient,
		s.dataRawClient, s.dataSourceClient, s.workClient)
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
