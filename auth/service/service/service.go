package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/service/api"
	"github.com/tidepool-org/platform/auth/service/api/v1"
	"github.com/tidepool-org/platform/auth/store"
	authMongo "github.com/tidepool-org/platform/auth/store/mongo"
	dataClient "github.com/tidepool-org/platform/data/client"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	providerFactory "github.com/tidepool-org/platform/provider/factory"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task"
	taskClient "github.com/tidepool-org/platform/task/client"
)

type Service struct {
	*serviceService.Service
	domain          string
	authStore       *authMongo.Store
	dataClient      dataClient.Client
	taskClient      task.Client
	providerFactory provider.Factory
	authClient      *Client
}

func New(prefix string) (*Service, error) {
	svc, err := serviceService.New(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Service: svc,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Service.Initialize(); err != nil {
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
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeTaskClient(); err != nil {
		return err
	}
	if err := s.initializeProviderFactory(); err != nil {
		return err
	}
	return s.initializeAuthClient()
}

func (s *Service) Terminate() {
	s.terminateAuthClient()
	s.terminateProviderFactory()
	s.terminateTaskClient()
	s.terminateDataClient()
	s.terminateAuthStore()
	s.terminateRouter()
	s.terminateDomain()

	s.Service.Terminate()
}

func (s *Service) Domain() string {
	return s.domain
}

func (s *Service) AuthStore() store.Store {
	return s.authStore
}

func (s *Service) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Service) TaskClient() task.Client {
	return s.taskClient
}

func (s *Service) ProviderFactory() provider.Factory {
	return s.providerFactory
}

func (s *Service) Status() *service.Status {
	return &service.Status{
		Version:   s.VersionReporter().Long(),
		AuthStore: s.AuthStore().Status(),
		Server:    s.API().Status(),
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
	routes := []*rest.Route{}

	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}
	routes = append(routes, apiRouter.Routes()...)

	s.Logger().Debug("Creating v1 router")

	v1Router, err := v1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}
	routes = append(routes, v1Router.Routes()...)

	s.Logger().Debug("Initializing router")

	if err = s.API().InitializeRouter(routes...); err != nil {
		return errors.Wrap(err, "unable to initialize router")
	}

	return nil
}

func (s *Service) terminateRouter() {
}

func (s *Service) initializeAuthStore() error {
	s.Logger().Debug("Loading auth store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("auth", "store")); err != nil {
		return errors.Wrap(err, "unable to load auth store config")
	}

	s.Logger().Debug("Creating auth store")

	str, err := authMongo.NewStore(cfg, s.Logger())
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
		s.authStore.Close()

		s.Logger().Debug("Destroying auth store")
		s.authStore = nil
	}
}

func (s *Service) initializeDataClient() error {
	s.Logger().Debug("Loading data client config")

	cfg := platform.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("data", "client")); err != nil {
		return errors.Wrap(err, "unable to load data client config")
	}

	s.Logger().Debug("Creating data client")

	clnt, err := dataClient.New(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Service) terminateDataClient() {
	if s.dataClient != nil {
		s.Logger().Debug("Destroying data client")
		s.dataClient = nil
	}
}

func (s *Service) initializeTaskClient() error {
	s.Logger().Debug("Loading task client config")

	cfg := platform.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("task", "client")); err != nil {
		return errors.Wrap(err, "unable to load task client config")
	}

	s.Logger().Debug("Creating task client")

	clnt, err := taskClient.New(cfg)
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

	if prvdr, prvdrErr := dexcomProvider.New(s.ConfigReporter().WithScopes("provider"), s.DataClient(), s.TaskClient()); prvdrErr != nil {
		s.Logger().Warn("Unable to create dexcom provider")
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
	if err := cfg.Load(s.ConfigReporter().WithScopes("auth", "client", "external")); err != nil {
		return errors.Wrap(err, "unable to load auth client config")
	}

	s.Logger().Debug("Creating auth client")

	clnt, err := NewClient(cfg, s.Name(), s.Logger(), s.AuthStore(), s.ProviderFactory())
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
