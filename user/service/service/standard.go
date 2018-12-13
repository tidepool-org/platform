package service

import (
	"github.com/tidepool-org/platform/application"
	confirmationMongo "github.com/tidepool-org/platform/confirmation/store/mongo"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageMongo "github.com/tidepool-org/platform/message/store/mongo"
	"github.com/tidepool-org/platform/metric"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	permissionMongo "github.com/tidepool-org/platform/permission/store/mongo"
	"github.com/tidepool-org/platform/platform"
	profileMongo "github.com/tidepool-org/platform/profile/store/mongo"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	sessionMongo "github.com/tidepool-org/platform/session/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user/service/api"
	userServiceApiV1 "github.com/tidepool-org/platform/user/service/api/v1"
	userMongo "github.com/tidepool-org/platform/user/store/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	dataClient        dataClient.Client
	metricClient      metric.Client
	permissionClient  permission.Client
	confirmationStore *confirmationMongo.Store
	messageStore      *messageMongo.Store
	permissionStore   *permissionMongo.Store
	profileStore      *profileMongo.Store
	sessionStore      *sessionMongo.Store
	userStore         *userMongo.Store
	api               *api.Standard
	server            *server.Standard
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

	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeMetricClient(); err != nil {
		return err
	}
	if err := s.initializePermissionClient(); err != nil {
		return err
	}
	if err := s.initializeConfirmationStore(); err != nil {
		return err
	}
	if err := s.initializeMessageStore(); err != nil {
		return err
	}
	if err := s.initializePermissionStore(); err != nil {
		return err
	}
	if err := s.initializeProfileStore(); err != nil {
		return err
	}
	if err := s.initializeSessionStore(); err != nil {
		return err
	}
	if err := s.initializeUserStore(); err != nil {
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
	if s.userStore != nil {
		s.userStore.Close()
		s.userStore = nil
	}
	if s.sessionStore != nil {
		s.sessionStore.Close()
		s.sessionStore = nil
	}
	if s.profileStore != nil {
		s.profileStore.Close()
		s.profileStore = nil
	}
	if s.permissionStore != nil {
		s.permissionStore.Close()
		s.permissionStore = nil
	}
	if s.messageStore != nil {
		s.messageStore.Close()
		s.messageStore = nil
	}
	if s.confirmationStore != nil {
		s.confirmationStore.Close()
		s.confirmationStore = nil
	}
	s.permissionClient = nil
	s.metricClient = nil
	s.dataClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	return s.server.Serve()
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Loading data client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	if err := cfg.Load(s.ConfigReporter().WithScopes("data", "client")); err != nil {
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

func (s *Standard) initializeConfirmationStore() error {
	s.Logger().Debug("Loading confirmation store config")

	confirmationStoreConfig := storeStructuredMongo.NewConfig()
	if err := confirmationStoreConfig.Load(s.ConfigReporter().WithScopes("confirmation", "store")); err != nil {
		return errors.Wrap(err, "unable to load confirmation store config")
	}

	s.Logger().Debug("Creating confirmation store")

	confirmationStore, err := confirmationMongo.NewStore(confirmationStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create confirmation store")
	}
	s.confirmationStore = confirmationStore

	return nil
}

func (s *Standard) initializeMessageStore() error {
	s.Logger().Debug("Loading message store config")

	messageStoreConfig := storeStructuredMongo.NewConfig()
	if err := messageStoreConfig.Load(s.ConfigReporter().WithScopes("message", "store")); err != nil {
		return errors.Wrap(err, "unable to load message store config")
	}

	s.Logger().Debug("Creating message store")

	messageStore, err := messageMongo.NewStore(messageStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create message store")
	}
	s.messageStore = messageStore

	return nil
}

func (s *Standard) initializePermissionStore() error {
	s.Logger().Debug("Loading permission store config")

	permissionStoreConfig := permissionMongo.NewConfig()
	if err := permissionStoreConfig.Load(s.ConfigReporter().WithScopes("permission", "store")); err != nil {
		return errors.Wrap(err, "unable to load permission store config")
	}

	s.Logger().Debug("Creating permission store")

	permissionStore, err := permissionMongo.NewStore(permissionStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create permission store")
	}
	s.permissionStore = permissionStore

	return nil
}

func (s *Standard) initializeProfileStore() error {
	s.Logger().Debug("Loading profile store config")

	profileStoreConfig := storeStructuredMongo.NewConfig()
	if err := profileStoreConfig.Load(s.ConfigReporter().WithScopes("profile", "store")); err != nil {
		return errors.Wrap(err, "unable to load profile store config")
	}

	s.Logger().Debug("Creating profile store")

	profileStore, err := profileMongo.NewStore(profileStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create profile store")
	}
	s.profileStore = profileStore

	return nil
}

func (s *Standard) initializeSessionStore() error {
	s.Logger().Debug("Loading session store config")

	sessionStoreConfig := storeStructuredMongo.NewConfig()
	if err := sessionStoreConfig.Load(s.ConfigReporter().WithScopes("session", "store")); err != nil {
		return errors.Wrap(err, "unable to load session store config")
	}

	s.Logger().Debug("Creating session store")

	sessionStore, err := sessionMongo.NewStore(sessionStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create session store")
	}
	s.sessionStore = sessionStore

	return nil
}

func (s *Standard) initializeUserStore() error {
	s.Logger().Debug("Loading user store config")

	userStoreConfig := userMongo.NewConfig()
	if err := userStoreConfig.Load(s.ConfigReporter().WithScopes("user", "store")); err != nil {
		return errors.Wrap(err, "unable to load user store config")
	}

	s.Logger().Debug("Creating user store")

	userStore, err := userMongo.NewStore(userStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create user store")
	}
	s.userStore = userStore

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.dataClient, s.metricClient, s.permissionClient,
		s.confirmationStore, s.messageStore, s.permissionStore, s.profileStore, s.sessionStore, s.userStore)
	if err != nil {
		return errors.Wrap(err, "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(userServiceApiV1.Routes()); err != nil {
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
