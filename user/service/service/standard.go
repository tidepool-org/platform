package service

import (
	confirmationMongo "github.com/tidepool-org/platform/confirmation/store/mongo"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageMongo "github.com/tidepool-org/platform/message/store/mongo"
	"github.com/tidepool-org/platform/metric"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionMongo "github.com/tidepool-org/platform/permission/store/mongo"
	"github.com/tidepool-org/platform/platform"
	profileMongo "github.com/tidepool-org/platform/profile/store/mongo"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	sessionMongo "github.com/tidepool-org/platform/session/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
	"github.com/tidepool-org/platform/user/service/api"
	"github.com/tidepool-org/platform/user/service/api/v1"
	userMongo "github.com/tidepool-org/platform/user/store/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	dataClient        dataClient.Client
	metricClient      metric.Client
	userClient        user.Client
	confirmationStore *confirmationMongo.Store
	messageStore      *messageMongo.Store
	permissionStore   *permissionMongo.Store
	profileStore      *profileMongo.Store
	sessionStore      *sessionMongo.Store
	userStore         *userMongo.Store
	api               *api.Standard
	server            *server.Standard
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

	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeMetricClient(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
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
	s.userClient = nil
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

	clnt, err := dataClient.New(cfg)
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
	cfg.UserAgent = s.UserAgent()
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

func (s *Standard) initializeConfirmationStore() error {
	s.Logger().Debug("Loading confirmation store config")

	confirmationStoreConfig := baseMongo.NewConfig()
	if err := confirmationStoreConfig.Load(s.ConfigReporter().WithScopes("confirmation", "store")); err != nil {
		return errors.Wrap(err, "unable to load confirmation store config")
	}

	s.Logger().Debug("Creating confirmation store")

	confirmationStore, err := confirmationMongo.New(confirmationStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create confirmation store")
	}
	s.confirmationStore = confirmationStore

	return nil
}

func (s *Standard) initializeMessageStore() error {
	s.Logger().Debug("Loading message store config")

	messageStoreConfig := baseMongo.NewConfig()
	if err := messageStoreConfig.Load(s.ConfigReporter().WithScopes("message", "store")); err != nil {
		return errors.Wrap(err, "unable to load message store config")
	}

	s.Logger().Debug("Creating message store")

	messageStore, err := messageMongo.New(messageStoreConfig, s.Logger())
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

	permissionStore, err := permissionMongo.New(permissionStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create permission store")
	}
	s.permissionStore = permissionStore

	return nil
}

func (s *Standard) initializeProfileStore() error {
	s.Logger().Debug("Loading profile store config")

	profileStoreConfig := baseMongo.NewConfig()
	if err := profileStoreConfig.Load(s.ConfigReporter().WithScopes("profile", "store")); err != nil {
		return errors.Wrap(err, "unable to load profile store config")
	}

	s.Logger().Debug("Creating profile store")

	profileStore, err := profileMongo.New(profileStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create profile store")
	}
	s.profileStore = profileStore

	return nil
}

func (s *Standard) initializeSessionStore() error {
	s.Logger().Debug("Loading session store config")

	sessionStoreConfig := baseMongo.NewConfig()
	if err := sessionStoreConfig.Load(s.ConfigReporter().WithScopes("session", "store")); err != nil {
		return errors.Wrap(err, "unable to load session store config")
	}

	s.Logger().Debug("Creating session store")

	sessionStore, err := sessionMongo.New(sessionStoreConfig, s.Logger())
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

	userStore, err := userMongo.New(userStoreConfig, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create user store")
	}
	s.userStore = userStore

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.dataClient, s.metricClient, s.userClient,
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
