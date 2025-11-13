package service

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/kelseyhightower/envconfig"

	eventsCommon "github.com/tidepool-org/go-common/events"

	abbottClient "github.com/tidepool-org/platform-plugin-abbott/abbott/client"
	abbottProvider "github.com/tidepool-org/platform-plugin-abbott/abbott/provider"
	abbottWork "github.com/tidepool-org/platform-plugin-abbott/abbott/work"

	confirmationClient "github.com/tidepool-org/hydrophone/client"
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
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
	"github.com/tidepool-org/platform/mailer"
	metricClient "github.com/tidepool-org/platform/metric/client"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/summary"
	summaryClient "github.com/tidepool-org/platform/summary/client"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
	"github.com/tidepool-org/platform/twiist"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
	workService "github.com/tidepool-org/platform/work/service"
	workStoreStructuredMongo "github.com/tidepool-org/platform/work/store/structured/mongo"

	"github.com/tidepool-org/platform/work/service/emailnotificationsprocessor"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient                   *metricClient.Client
	permissionClient               *permissionClient.Client
	dataStore                      *dataStoreMongo.Store
	dataRawStructuredStore         *dataRawStoreStructuredMongo.Store
	dataSourceStructuredStore      *dataSourceStoreStructuredMongo.Store
	syncTaskStore                  *syncTaskMongo.Store
	workStructuredStore            *workStoreStructuredMongo.Store
	dataDeduplicatorFactory        *dataDeduplicatorFactory.Factory
	clinicsClient                  clinics.Client
	dataClient                     *Client
	dataRawClient                  *dataRawService.Client
	dataSourceClient               *dataSourceServiceClient.Client
	mailerClient                   mailer.Mailer
	summaryClient                  *summaryClient.Client
	workClient                     *workService.Client
	abbottClient                   *abbottClient.Client
	userClient                     user.Client
	confirmationClient             confirmationClient.ClientWithResponsesInterface
	workCoordinator                *workService.Coordinator
	userEventsHandler              events.Runner
	twiistServiceAccountAuthorizer *twiist.ServiceAccountAuthorizer
	api                            *api.Standard
	server                         *server.Standard
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
	if err := s.initializeMailerClient(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
		return err
	}
	if err := s.initializeSummaryClient(); err != nil {
		return err
	}
	if err := s.initializeConfirmationClient(); err != nil {
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
	if err := s.initializeTwiistServiceAccountAuthorizer(); err != nil {
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
	s.twiistServiceAccountAuthorizer = nil
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
	s.summaryClient = nil
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
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
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

	s.Logger().Debug("Creating data set delete origin older deduplicator")

	dataSetDeleteOriginOlderDeduplicator, err := dataDeduplicatorDeduplicator.NewDataSetDeleteOriginOlder(dependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create data set delete origin older deduplicator")
	}

	s.Logger().Debug("Creating data set drop hash deduplicator")

	dataSetDropHashDeduplicator, err := dataDeduplicatorDeduplicator.NewDataSetDropHash(dependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create data set drop hash deduplicator")
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
		dataSetDeleteOriginOlderDeduplicator,
		dataSetDropHashDeduplicator,
		noneDeduplicator,
	}

	factory, err := dataDeduplicatorFactory.New(deduplicators)
	if err != nil {
		return errors.Wrap(err, "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = factory

	return nil
}

func (s *Standard) initializeClinicsClient() error {
	s.Logger().Debug("Creating clinics client")

	clnt, err := clinics.NewClient(s.AuthClient())
	if err != nil {
		return errors.Wrap(err, "unable to create clinics client")
	}
	s.clinicsClient = clnt

	return nil
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := NewClient(s.dataStore, s.dataDeduplicatorFactory)
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
	s.Logger().Debug("Creating data source client")

	clnt, err := dataSourceServiceClient.New(s)
	if err != nil {
		return errors.Wrap(err, "unable to create data source client")
	}
	s.dataSourceClient = clnt

	return nil
}

func (s *Standard) initializeMailerClient() error {
	s.Logger().Debug("Initializing mailer client")
	client, err := mailer.Client()
	if err != nil {
		return errors.Wrap(err, "unable to create mailer client")
	}
	s.mailerClient = client
	return nil
}

func (s *Standard) initializeUserClient() error {
	s.Logger().Debug("Initializing user client")
	client, err := userClient.NewDefaultClient(userClient.Params{
		ConfigReporter: s.ConfigReporter(),
		Logger:         s.Logger(),
		UserAgent:      s.UserAgent(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	s.userClient = client
	return nil
}

type confirmationClientConfig struct {
	ServiceAddress string `envconfig:"TIDEPOOL_CONFIRMATION_CLIENT_ADDRESS"`
}

func (c *confirmationClientConfig) Load() error {
	return envconfig.Process("", c)
}

func (s *Standard) initializeConfirmationClient() error {
	s.Logger().Debug("Initializing confirmation client")

	cfg := &confirmationClientConfig{}
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load confirmations client config")
	}

	opts := confirmationClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		token, err := s.AuthClient().ServerSessionToken()
		if err != nil {
			return err
		}

		req.Header.Add(auth.TidepoolSessionTokenHeaderKey, token)
		return nil
	})

	client, err := confirmationClient.NewClientWithResponses(cfg.ServiceAddress, opts)
	if err != nil {
		return errors.Wrap(err, "unable to create confirmations client")
	}
	s.confirmationClient = client

	return nil
}

func (s *Standard) initializeSummaryClient() error {
	s.Logger().Debug("Creating summarizer registry")

	summarizerRegistry := summary.New(
		s.dataStore.NewSummaryRepository().GetStore(),
		s.dataStore.NewBucketsRepository().GetStore(),
		s.dataStore.NewDataRepository(),
		s.dataStore.GetClient(),
	)

	s.Logger().Debug("Creating summary client")

	clnt, err := summaryClient.New(summarizerRegistry)
	if err != nil {
		return errors.Wrap(err, "unable to create summary client")
	}
	s.summaryClient = clnt

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

	// Abbott
	abbottJWKS, err := oauthProvider.NewJWKS(s.ConfigReporter().WithScopes("provider", abbottProvider.ProviderName))
	if err != nil {
		return errors.Wrap(err, "unable to create abbott jwks")
	}
	abbottProviderDependencies := abbottProvider.ProviderDependencies{
		ConfigReporter:        s.ConfigReporter().WithScopes("provider"),
		ProviderSessionClient: s.AuthClient(),
		DataSourceClient:      s.dataSourceClient,
		WorkClient:            s.workClient,
		JWKS:                  abbottJWKS,
	}
	if prvdr, err := abbottProvider.New(abbottProviderDependencies); err != nil {
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
		SummaryClient:           s.summaryClient,
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

	emailDependencies := emailnotificationsprocessor.Dependencies{
		DataSources:   s.dataSourceStructuredStore.NewDataSourcesRepository(),
		Mailer:        s.mailerClient,
		Auth:          s.AuthClient(),
		Users:         s.userClient,
		Clinics:       s.clinicsClient,
		Confirmations: s.confirmationClient,
	}
	emailNotifProcessors, err := emailnotificationsprocessor.NewProcessors(emailDependencies)
	if err != nil {
		return errors.Wrap(err, "unable to create email notifications processor")
	}

	if err = s.workCoordinator.RegisterProcessors(emailNotifProcessors); err != nil {
		return errors.Wrap(err, "unable to register email notifications processor")
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

func (s *Standard) initializeTwiistServiceAccountAuthorizer() error {
	s.Logger().Debug("Initializing twiist service account authorizer")

	twiistServiceAccountAuthorizer, err := twiist.NewServiceAccountAuthorizer()
	if err != nil {
		return errors.Wrap(err, "unable to initialize twiist service account authorizer")
	}
	s.twiistServiceAccountAuthorizer = twiistServiceAccountAuthorizer

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient,
		s.dataRawClient, s.dataSourceClient, s.workClient,
		s.abbottClient, s.twiistServiceAccountAuthorizer)
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
