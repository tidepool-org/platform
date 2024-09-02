package service

import (
	"context"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	eventsCommon "github.com/tidepool-org/go-common/events"
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/client"
	authEvents "github.com/tidepool-org/platform/auth/events"
	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/service/api"
	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	"github.com/tidepool-org/platform/auth/store"
	authMongo "github.com/tidepool-org/platform/auth/store/mongo"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceClient "github.com/tidepool-org/platform/data/source/client"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/events"
	logInternal "github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	providerFactory "github.com/tidepool-org/platform/provider/factory"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/task"
	taskClient "github.com/tidepool-org/platform/task/client"
)

type confirmationClientConfig struct {
	ServiceAddress string `envconfig:"TIDEPOOL_CONFIRMATION_CLIENT_ADDRESS"`
}

func (c *confirmationClientConfig) Load() error {
	return envconfig.Process("", c)
}

type Service struct {
	*serviceService.Service
	domain             string
	authStore          *authMongo.Store
	dataSourceClient   *dataSourceClient.Client
	confirmationClient confirmationClient.ClientWithResponsesInterface
	taskClient         task.Client
	providerFactory    provider.Factory
	authClient         *Client
	userEventsHandler  events.Runner
	deviceCheck        apple.DeviceCheck
	appValidator       *appvalidate.Validator
	partnerSecrets     *appvalidate.PartnerSecrets
}

func New() *Service {
	return &Service{
		Service: serviceService.New(),
	}
}

func (s *Service) Run() error {
	errs := make(chan error)
	go func() {
		errs <- s.userEventsHandler.Run()
	}()
	go func() {
		errs <- s.Service.Run()
	}()

	return <-errs
}

func (s *Service) Initialize(provider application.Provider) error {
	if err := s.Service.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeDomain(); err != nil {
		return err
	}
	if err := s.initializeRouter(); err != nil {
		return err
	}
	if err := s.initializeAuthStore(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
		return err
	}
	if err := s.initializeConfirmationClient(); err != nil {
		return err
	}
	if err := s.initializeTaskClient(); err != nil {
		return err
	}
	if err := s.initializeProviderFactory(); err != nil {
		return err
	}
	if err := s.initializeAuthClient(); err != nil {
		return err
	}
	if err := s.initializeDeviceCheck(); err != nil {
		return err
	}
	if err := s.initializeAppValidate(); err != nil {
		return err
	}
	if err := s.initializePartnerSecrets(); err != nil {
		return err
	}
	return s.initializeUserEventsHandler()
}

func (s *Service) Terminate() {
	s.Service.Terminate()
	s.terminateUserEventsHandler()
	s.terminateAuthClient()
	s.terminateProviderFactory()
	s.terminateTaskClient()
	s.terminateDataSourceClient()
	s.terminateConfirmationClient()
	s.terminateAuthStore()
	s.terminateRouter()
	s.terminateDomain()
}

func (s *Service) Domain() string {
	return s.domain
}

func (s *Service) AuthStore() store.Store {
	if s.authStore == nil {
		return nil
	}
	return s.authStore
}

func (s *Service) DataSourceClient() dataSource.Client {
	return s.dataSourceClient
}

func (s *Service) ConfirmationClient() confirmationClient.ClientWithResponsesInterface {
	return s.confirmationClient
}

func (s *Service) TaskClient() task.Client {
	return s.taskClient
}

func (s *Service) ProviderFactory() provider.Factory {
	return s.providerFactory
}

func (s *Service) DeviceCheck() apple.DeviceCheck {
	return s.deviceCheck
}

func (s *Service) AppValidator() *appvalidate.Validator {
	return s.appValidator
}

func (s *Service) PartnerSecrets() *appvalidate.PartnerSecrets {
	return s.partnerSecrets
}

func (s *Service) Status(ctx context.Context) *service.Status {
	return &service.Status{
		Version: s.VersionReporter().Long(),
	}
}

func (s *Service) initializeDomain() error {
	s.Logger().Debug("Initializing domain")

	domain := s.ConfigReporter().GetWithDefault("domain", "")
	if domain == "" {
		return errors.New("domain is missing")
	}
	s.domain = domain

	return nil
}

func (s *Service) terminateDomain() {
	if s.domain != "" {
		s.Logger().Debug("Terminating domain")
		s.domain = ""
	}
}

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}

	s.Logger().Debug("Creating v1 router")

	v1Router, err := authServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}

	s.Logger().Debug("Initializing routers")

	if err = s.API().InitializeRouters(apiRouter, v1Router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (s *Service) terminateRouter() {
}

func (s *Service) initializeAuthStore() error {
	s.Logger().Debug("Loading auth store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load auth store config")
	}

	s.Logger().Debug("Creating auth store")

	str, err := authMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create auth store")
	}
	s.authStore = str

	s.Logger().Debug("Ensuring auth store indexes")

	err = s.authStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure auth store indexes")
	}

	return nil
}

func (s *Service) terminateAuthStore() {
	if s.authStore != nil {
		s.Logger().Debug("Closing auth store")
		s.authStore.Terminate(context.Background())

		s.Logger().Debug("Destroying auth store")
		s.authStore = nil
	}
}

func (s *Service) initializeDataSourceClient() error {
	s.Logger().Debug("Loading data source client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("data_source", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load data source client config")
	}

	s.Logger().Debug("Creating data source client")

	clnt, err := dataSourceClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create data source client")
	}
	s.dataSourceClient = clnt

	return nil
}

func (s *Service) terminateDataSourceClient() {
	if s.dataSourceClient != nil {
		s.Logger().Debug("Destroying data source client")
		s.dataSourceClient = nil
	}
}

func (s *Service) initializeConfirmationClient() error {
	s.Logger().Debug("Loading confirmation client config")

	cfg := &confirmationClientConfig{}
	if err := cfg.Load(); err != nil {
		return err
	}

	opts := confirmationClient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		token, err := s.authClient.ServerSessionToken()
		if err != nil {
			return err
		}

		req.Header.Add(auth.TidepoolSessionTokenHeaderKey, token)
		return nil
	})

	clnt, err := confirmationClient.NewClientWithResponses(cfg.ServiceAddress, opts)
	if err != nil {
		return err
	}
	s.confirmationClient = clnt

	return nil
}

func (s *Service) terminateConfirmationClient() {
	if s.confirmationClient != nil {
		s.Logger().Debug("Destroying confirmation client")
		s.confirmationClient = nil
	}
}

func (s *Service) initializeTaskClient() error {
	s.Logger().Debug("Loading task client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("task", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load task client config")
	}

	s.Logger().Debug("Creating task client")

	clnt, err := taskClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create task client")
	}
	s.taskClient = clnt

	return nil
}

func (s *Service) terminateTaskClient() {
	if s.taskClient != nil {
		s.Logger().Debug("Destroying task client")
		s.taskClient = nil
	}
}

func (s *Service) initializeProviderFactory() error {
	s.Logger().Debug("Creating provider factory")

	prvdrFctry, err := providerFactory.New()
	if err != nil {
		return errors.Wrap(err, "unable to create provider factory")
	}

	s.providerFactory = prvdrFctry

	if prvdr, prvdrErr := dexcomProvider.New(s.ConfigReporter().WithScopes("provider"), s.DataSourceClient(), s.TaskClient()); prvdrErr != nil {
		s.Logger().WithError(prvdrErr).Warn("Unable to create dexcom provider")
	} else if prvdrErr = prvdrFctry.Add(prvdr); prvdrErr != nil {
		return errors.Wrap(prvdrErr, "unable to add dexcom provider")
	}

	return nil
}

func (s *Service) terminateProviderFactory() {
	if s.providerFactory != nil {
		s.Logger().Debug("Destroying provider factory")
		s.providerFactory = nil
	}
}

func (s *Service) initializeAuthClient() error {
	s.Logger().Debug("Loading auth client config")

	cfg := client.NewExternalConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("auth", "client", "external")
	loader := client.NewExternalConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load auth client config")
	}

	s.Logger().Debug("Creating auth client")

	clnt, err := NewClient(cfg, platform.AuthorizeAsService, s.Name(), s.Logger(), s.AuthStore(), s.ProviderFactory())
	if err != nil {
		return errors.Wrap(err, "unable to create auth client")
	}
	s.authClient = clnt

	s.Logger().Debug("Starting auth client")

	if err = s.authClient.Start(); err != nil {
		return errors.Wrap(err, "unable to start auth client")
	}

	s.SetAuthClient(s.authClient)

	return nil
}

func (s *Service) terminateAuthClient() {
	if s.authClient != nil {
		s.Logger().Debug("Closing auth client")
		s.authClient.Close()

		s.Logger().Debug("Destroying auth client")
		s.authClient = nil

		s.SetAuthClient(nil)
	}
}

func (s *Service) initializeUserEventsHandler() error {
	s.Logger().Debug("Initializing user events handler")

	ctx := logInternal.NewContextWithLogger(context.Background(), s.Logger())
	handler := authEvents.NewUserDataDeletionHandler(ctx, s.authClient)
	handlers := []eventsCommon.EventHandler{handler}
	runner := events.NewRunner(handlers)

	if err := runner.Initialize(); err != nil {
		return errors.Wrap(err, "unable to initialize events runner")
	}
	s.userEventsHandler = runner

	return nil
}

func (s *Service) initializeDeviceCheck() error {
	s.Logger().Debug("Initializing device check")

	cfg := apple.NewDeviceCheckConfig()
	if err := cfg.Load(); err != nil {
		s.Logger().Errorf("error loading device check config: %v", err)
		return err
	}

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}
	s.deviceCheck = apple.NewDeviceCheck(cfg, httpClient)

	return nil
}

func (s *Service) initializeAppValidate() error {
	s.Logger().Debug("Initializing app validate")
	cfg, err := appvalidate.NewValidatorConfig()
	if err != nil {
		return err
	}
	s.Logger().Infof("Initialized AppValidate with: %#v", *cfg)
	authStore := s.AuthStore()
	if authStore == nil {
		return errors.New("auth store should be initialized before app validate")
	}
	validator, err := appvalidate.NewValidator(authStore.NewAppValidateRepository(), appvalidate.NewChallengeGenerator(), *cfg)
	if err != nil {
		return err
	}
	s.appValidator = validator
	return nil
}

func (s *Service) initializePartnerSecrets() error {
	s.Logger().Debug("Initializing partner secrets")
	var err error
	var coastalSecrets *appvalidate.CoastalSecrets
	var palmtreeSecrets *appvalidate.PalmTreeSecrets

	// We are OK with partner secrets being missing so we only log any errors.
	coastalCfg, err := appvalidate.NewCoastalSecretsConfig()
	if err != nil {
		s.Logger().Warnf("error initializing Coastal config: %v", err)
	} else {
		coastalSecrets, err = appvalidate.NewCoastalSecrets(*coastalCfg)
		if err != nil {
			s.Logger().Warnf("error initializing Coastal secrets: %v", err)
		}
	}

	palmtreeCfg, err := appvalidate.NewPalmTreeSecretsConfig()
	if err != nil {
		s.Logger().Warnf("error initializing Palmtree config: %v", err)
	} else {
		palmtreeSecrets, err = appvalidate.NewPalmTreeSecrets(*palmtreeCfg)
		if err != nil {
			s.Logger().Warnf("error initializing Palmtree secrets: %v", err)
		}
	}
	s.partnerSecrets = appvalidate.NewPartnerSecrets(coastalSecrets, palmtreeSecrets)
	return nil
}

func (s *Service) terminateUserEventsHandler() {
	if s.userEventsHandler != nil {
		s.Logger().Info("Terminating the userEventsHandler")
		if err := s.userEventsHandler.Terminate(); err != nil {
			s.Logger().Errorf("Error while terminating the userEventsHandler: %v", err)
		}
		s.userEventsHandler = nil
	}
}
