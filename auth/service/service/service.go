package service

import (
	"context"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/customerio"
	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/oura/jotform"
	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/user"

	userClient "github.com/tidepool-org/platform/user/client"

	"github.com/kelseyhightower/envconfig"

	eventsCommon "github.com/tidepool-org/go-common/events"
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	abbottProvider "github.com/tidepool-org/platform-plugin-abbott/abbott/provider"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authEvents "github.com/tidepool-org/platform/auth/events"
	authService "github.com/tidepool-org/platform/auth/service"
	authServiceApi "github.com/tidepool-org/platform/auth/service/api"
	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	authStore "github.com/tidepool-org/platform/auth/store"
	authStoreMongo "github.com/tidepool-org/platform/auth/store/mongo"
	"github.com/tidepool-org/platform/consent"
	consentApiV1 "github.com/tidepool-org/platform/consent/api/v1"
	consentLoader "github.com/tidepool-org/platform/consent/loader"
	consentService "github.com/tidepool-org/platform/consent/service"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceClient "github.com/tidepool-org/platform/data/source/client"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/events"
	jotformAPI "github.com/tidepool-org/platform/oura/jotform/api"
	shopifyAPI "github.com/tidepool-org/platform/oura/shopify/api"
	shopifyClient "github.com/tidepool-org/platform/oura/shopify/client"

	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	providerFactory "github.com/tidepool-org/platform/provider/factory"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/task"
	taskClient "github.com/tidepool-org/platform/task/client"
	"github.com/tidepool-org/platform/twiist"
	twiistProvider "github.com/tidepool-org/platform/twiist/provider"
	"github.com/tidepool-org/platform/work"
	workService "github.com/tidepool-org/platform/work/service"
	workStoreStructuredMongo "github.com/tidepool-org/platform/work/store/structured/mongo"
)

type confirmationClientConfig struct {
	ServiceAddress string `envconfig:"TIDEPOOL_CONFIRMATION_CLIENT_ADDRESS"`
}

func (c *confirmationClientConfig) Load() error {
	return envconfig.Process("", c)
}

type Service struct {
	*serviceService.Service
	domain                         string
	appValidator                   *appvalidate.Validator
	authClient                     *Client
	authStore                      *authStoreMongo.Store
	confirmationClient             confirmationClient.ClientWithResponsesInterface
	customerIOClient               *customerio.Client
	dataClient                     dataClient.Client
	dataSourceClient               *dataSourceClient.Client
	deviceCheck                    apple.DeviceCheck
	partnerSecrets                 *appvalidate.PartnerSecrets
	providerFactory                *providerFactory.Factory
	shopifyClient                  shopify.Client
	taskClient                     task.Client
	twiistServiceAccountAuthorizer auth.ServiceAccountAuthorizer
	userEventsHandler              events.Runner
	userClient                     user.Client
	consentService                 consent.Service
	workClient                     *workService.Client
	workStructuredStore            *workStoreStructuredMongo.Store
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
	if err := s.initializeAuthStore(); err != nil {
		return err
	}
	if err := s.initializeWorkStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
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
	if err := s.initializeWorkClient(); err != nil {
		return err
	}
	if err := s.initializeProviderFactory(); err != nil {
		return err
	}
	if err := s.initializeAuthClient(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
		return err
	}
	if err := s.initializeConsentService(); err != nil {
		return err
	}
	if err := s.initializeProviders(); err != nil {
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
	if err := s.initializeTwiistServiceAccountAuthorizer(); err != nil {
		return err
	}
	if err := s.initializeCustomerIOClient(); err != nil {
		return err
	}
	if err := s.initializeShopifyClient(); err != nil {
		return err
	}
	if err := s.initializeRouter(); err != nil {
		return err
	}
	return s.initializeUserEventsHandler()
}

func (s *Service) Terminate() {
	s.Service.Terminate()
	s.terminateUserEventsHandler()
	s.terminateAuthClient()
	s.terminateProviderFactory()
	s.terminateWorkClient()
	s.terminateTaskClient()
	s.terminateDataSourceClient()
	s.terminateConfirmationClient()
	s.terminateWorkStructuredStore()
	s.terminateAuthStore()
	s.terminateRouter()
	s.terminateDomain()
}

func (s *Service) Domain() string {
	return s.domain
}

func (s *Service) AuthStore() authStore.Store {
	if s.authStore == nil {
		return nil
	}
	return s.authStore
}

func (s *Service) AuthServiceClient() authService.Client {
	return s.authClient
}

func (s *Service) DataSourceClient() dataSource.Client {
	return s.dataSourceClient
}

func (s *Service) ConfirmationClient() confirmationClient.ClientWithResponsesInterface {
	return s.confirmationClient
}

func (s *Service) ConsentService() consent.Service {
	return s.consentService
}

func (s *Service) TaskClient() task.Client {
	return s.taskClient
}

func (s *Service) WorkClient() work.Client {
	return s.workClient
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

func (s *Service) TwiistServiceAccountAuthorizer() auth.ServiceAccountAuthorizer {
	return s.twiistServiceAccountAuthorizer
}

func (s *Service) Status(ctx context.Context) *authService.Status {
	return &authService.Status{
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

func (s *Service) initializeUserClient() error {
	s.Logger().Debug("Initializing user client")
	var err error
	s.userClient, err = userClient.NewDefaultClient(userClient.Params{
		ConfigReporter: s.ConfigReporter(),
		Logger:         s.Logger(),
		UserAgent:      s.UserAgent(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	return nil
}

func (s *Service) initializeShopifyClient() error {
	shopifyConfig := shopify.ClientConfig{}
	if err := envconfig.Process("", &shopifyConfig); err != nil {
		return errors.Wrap(err, "unable to load shopify config")
	}

	var err error
	s.shopifyClient, err = shopifyClient.New(context.Background(), shopifyConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create shopify client")
	}

	return nil
}
func (s *Service) initializeCustomerIOClient() error {
	customerIOConfig := customerio.Config{}
	if err := envconfig.Process("", &customerIOConfig); err != nil {
		return errors.Wrap(err, "unable to load customerio config")
	}

	var err error
	s.customerIOClient, err = customerio.NewClient(customerIOConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create customerio client")
	}

	return nil
}

func (s *Service) createJotformRouter() (*jotformAPI.Router, error) {
	jotformConfig := jotform.Config{}
	if err := envconfig.Process("", &jotformConfig); err != nil {
		return nil, errors.Wrap(err, "unable to load jotform config")
	}

	webhookProcessor, err := jotform.NewWebhookProcessor(jotformConfig, s.Logger(), s.consentService, s.customerIOClient, s.userClient, s.shopifyClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create jotform webhook processor")
	}

	jotformRouter, err := jotformAPI.NewRouter(webhookProcessor)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create jotform router")
	}

	return jotformRouter, nil
}

func (s *Service) createShopifyRouter() (*shopifyAPI.Router, error) {
	fulfillmentEventProcessor, err := shopify.NewFulfillmentEventProcessor(s.Logger(), s.customerIOClient, s.shopifyClient, s.AuthServiceClient())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create fulfillment event processor")
	}

	ordersCreateEventProcessor, err := shopify.NewOrdersCreateEventProcessor(s.Logger(), s.customerIOClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create orders create event processor")
	}

	shopifyRouter, err := shopifyAPI.NewRouter(fulfillmentEventProcessor, ordersCreateEventProcessor)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create shopify router")
	}

	return shopifyRouter, nil
}

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating api router")

	apiRouter, err := authServiceApi.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}

	s.Logger().Debug("Creating v1 router")

	v1Router, err := authServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}

	s.Logger().Debug("Creating consent router")

	consentV1Router, err := consentApiV1.NewRouter(s.consentService)
	if err != nil {
		return errors.Wrap(err, "unable to create consent router")
	}

	s.Logger().Debug("Creating jotform router")

	jotformRouter, err := s.createJotformRouter()
	if err != nil {
		return errors.Wrap(err, "unable to create jotform router")
	}

	s.Logger().Debug("Creating shopify router")

	shopifyRouter, err := s.createShopifyRouter()
	if err != nil {
		return errors.Wrap(err, "unable to create shopify router")
	}

	s.Logger().Debug("Initializing routers")

	if err = s.API().InitializeRouters(apiRouter, v1Router, consentV1Router, jotformRouter, shopifyRouter); err != nil {
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

	str, err := authStoreMongo.NewStore(cfg)
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

func (s *Service) initializeWorkStructuredStore() error {
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

func (s *Service) initializeConsentService() error {
	s.Logger().Debug("Initializing big data donation project sharer")

	bddpSharerConfig := consentService.BigDataDonationProjectConfig{}
	if err := envconfig.Process("", &bddpSharerConfig); err != nil {
		return errors.Wrap(err, "unable to load bddp sharer config")
	}

	bddpSharer, err := consentService.NewBigDataDonationProjectSharer(bddpSharerConfig, s.AuthClient())
	if err != nil {
		return errors.Wrap(err, "unable to create bddp sharer")
	}

	s.Logger().Debug("Initializing mailer")
	mailr, err := mailer.Client()
	if err != nil {
		return errors.Wrap(err, "unable to create mailer")
	}

	s.Logger().Debug("Initializing consent mailer")
	consentMailer, err := consentService.NewConsentMailer(mailr, s.userClient, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create consent mailer")
	}

	s.Logger().Debug("Initializing consent service")
	s.consentService = consentService.NewConsentService(consentMailer, bddpSharer, s.authStore.NewConsentRepository(), s.authStore.NewConsentRecordRepository(), s.authStore.Store.GetClient(), s.Logger())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return consentLoader.SeedConsents(ctx, s.Logger(), s.consentService)
}

func (s *Service) terminateWorkStructuredStore() {
	if s.workStructuredStore != nil {
		s.Logger().Debug("Closing work structured store")
		s.workStructuredStore.Terminate(context.Background())

		s.Logger().Debug("Destroying work structured store")
		s.workStructuredStore = nil
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

func (s *Service) initializeDataClient() error {
	s.Logger().Debug("Loading data client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("data", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load data client config")
	}

	s.Logger().Debug("Creating data client")

	clnt, err := dataClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Service) terminateDataSourceClient() {
	if s.dataSourceClient != nil {
		s.Logger().Debug("Destroying data source client")
		s.dataSourceClient = nil
	}
}

func (s *Service) initializeWorkClient() error {
	s.Logger().Debug("Creating work client")

	clnt, err := workService.NewClient(s.workStructuredStore)
	if err != nil {
		return errors.Wrap(err, "unable to create work client")
	}
	s.workClient = clnt

	return nil
}

func (s *Service) terminateWorkClient() {
	if s.workClient != nil {
		s.Logger().Debug("Destroying work client")
		s.workClient = nil
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

	cfg := authClient.NewExternalConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("auth", "client", "external")
	loader := authClient.NewExternalConfigReporterLoader(reporter)
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

func (s *Service) initializeProviders() error {

	configReporter := s.ConfigReporter().WithScopes("provider")

	// Abbott
	abbottJWKS, err := oauthProvider.NewJWKS(configReporter.WithScopes(abbottProvider.ProviderName))
	if err != nil {
		return errors.Wrap(err, "unable to create abbott jwks")
	}
	abbottProviderDependencies := abbottProvider.ProviderDependencies{
		ConfigReporter:        configReporter,
		ProviderSessionClient: s.AuthClient(),
		DataSourceClient:      s.DataSourceClient(),
		WorkClient:            s.workClient,
		JWKS:                  abbottJWKS,
	}
	if prvdr, prvdrErr := abbottProvider.New(abbottProviderDependencies); prvdrErr != nil || prvdr == nil {
		s.Logger().WithError(prvdrErr).Warn("Unable to create abbott provider")
	} else if prvdrErr = s.providerFactory.Add(prvdr); prvdrErr != nil {
		return errors.Wrap(prvdrErr, "unable to add abbott provider")
	}

	// Dexcom
	if prvdr, prvdrErr := dexcomProvider.New(configReporter, s.DataSourceClient(), s.TaskClient()); prvdrErr != nil || prvdr == nil {
		s.Logger().WithError(prvdrErr).Warn("Unable to create dexcom provider")
	} else if prvdrErr = s.providerFactory.Add(prvdr); prvdrErr != nil {
		return errors.Wrap(prvdrErr, "unable to add dexcom provider")
	}

	// twiist
	twiistJWKS, err := oauthProvider.NewJWKS(configReporter.WithScopes(twiistProvider.ProviderName))
	if err != nil {
		return errors.Wrap(err, "unable to create twiist jwks")
	}
	twiistProviderDependencies := twiistProvider.ProviderDependencies{
		ConfigReporter:        configReporter,
		ProviderSessionClient: s.AuthClient(),
		DataSourceClient:      s.DataSourceClient(),
		DataSetClient:         s.dataClient,
		JWKS:                  twiistJWKS,
	}
	if prvdr, prvdrErr := twiistProvider.New(twiistProviderDependencies); prvdrErr != nil || prvdr == nil {
		s.Logger().WithError(prvdrErr).Warn("Unable to create twiist provider")
	} else if prvdrErr = s.providerFactory.Add(prvdr); prvdrErr != nil {
		return errors.Wrap(prvdrErr, "unable to add twiist provider")
	}

	return nil
}

func (s *Service) initializeUserEventsHandler() error {
	s.Logger().Debug("Initializing user events handler")

	ctx := log.NewContextWithLogger(context.Background(), s.Logger())
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
	var palmTreeSecrets *appvalidate.PalmTreeSecrets

	// We are OK with partner secrets being missing so we only log any errors.
	coastalConfig, err := appvalidate.NewCoastalSecretsConfig(s.Logger())
	if err != nil {
		s.Logger().Warnf("error initializing coastal config: %v", err)
	} else {
		coastalSecrets, err = appvalidate.NewCoastalSecrets(s.Logger(), *coastalConfig)
		if err != nil {
			s.Logger().Warnf("error initializing coastal secrets: %v", err)
		}
	}

	palmTreeConfig, err := appvalidate.NewPalmTreeSecretsConfig(s.Logger())
	if err != nil {
		s.Logger().Warnf("error initializing palm tree config: %v", err)
	} else {
		palmTreeSecrets, err = appvalidate.NewPalmTreeSecrets(s.Logger(), *palmTreeConfig)
		if err != nil {
			s.Logger().Warnf("error initializing palm tree secrets: %v", err)
		}
	}
	s.partnerSecrets = appvalidate.NewPartnerSecrets(coastalSecrets, palmTreeSecrets)
	return nil
}

func (s *Service) initializeTwiistServiceAccountAuthorizer() error {
	s.Logger().Debug("Initializing twiist service account authorizer")

	twiistServiceAccountAuthorizer, err := twiist.NewServiceAccountAuthorizer()
	if err != nil {
		return errors.Wrap(err, "unable to initialize twiist service account authorizer")
	}
	s.twiistServiceAccountAuthorizer = twiistServiceAccountAuthorizer

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
