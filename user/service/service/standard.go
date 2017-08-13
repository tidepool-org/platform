package service

import (
	"github.com/tidepool-org/platform/client"
	confirmationMongo "github.com/tidepool-org/platform/confirmation/store/mongo"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageMongo "github.com/tidepool-org/platform/message/store/mongo"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionMongo "github.com/tidepool-org/platform/permission/store/mongo"
	profileMongo "github.com/tidepool-org/platform/profile/store/mongo"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	sessionMongo "github.com/tidepool-org/platform/session/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	userClient "github.com/tidepool-org/platform/user/client"
	"github.com/tidepool-org/platform/user/service/api"
	"github.com/tidepool-org/platform/user/service/api/v1"
	userMongo "github.com/tidepool-org/platform/user/store/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	dataClient        dataClient.Client
	metricClient      metricClient.Client
	userClient        userClient.Client
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
	if err := s.initializeServer(); err != nil {
		return err
	}

	return nil
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
		return errors.New("service", "service not initialized")
	}

	return s.server.Serve()
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Loading data client config")

	dataClientConfig := client.NewConfig()
	if err := dataClientConfig.Load(s.ConfigReporter().WithScopes("data", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load data client config")
	}

	s.Logger().Debug("Creating data client")

	dataClient, err := dataClient.NewClient(dataClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data client")
	}
	s.dataClient = dataClient

	return nil
}

func (s *Standard) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	metricClientConfig := client.NewConfig()
	if err := metricClientConfig.Load(s.ConfigReporter().WithScopes("metric", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	metricClient, err := metricClient.NewClient(metricClientConfig, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "service", "unable to create metric client")
	}
	s.metricClient = metricClient

	return nil
}

func (s *Standard) initializeUserClient() error {
	s.Logger().Debug("Loading user client config")

	userClientConfig := client.NewConfig()
	if err := userClientConfig.Load(s.ConfigReporter().WithScopes("user", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load user client config")
	}

	s.Logger().Debug("Creating user client")

	userClient, err := userClient.NewClient(userClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user client")
	}
	s.userClient = userClient

	return nil
}

func (s *Standard) initializeConfirmationStore() error {
	s.Logger().Debug("Loading confirmation store config")

	confirmationStoreConfig := baseMongo.NewConfig()
	if err := confirmationStoreConfig.Load(s.ConfigReporter().WithScopes("confirmation", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load confirmation store config")
	}

	s.Logger().Debug("Creating confirmation store")

	confirmationStore, err := confirmationMongo.New(s.Logger(), confirmationStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create confirmation store")
	}
	s.confirmationStore = confirmationStore

	return nil
}

func (s *Standard) initializeMessageStore() error {
	s.Logger().Debug("Loading message store config")

	messageStoreConfig := baseMongo.NewConfig()
	if err := messageStoreConfig.Load(s.ConfigReporter().WithScopes("message", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load message store config")
	}

	s.Logger().Debug("Creating message store")

	messageStore, err := messageMongo.New(s.Logger(), messageStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create message store")
	}
	s.messageStore = messageStore

	return nil
}

func (s *Standard) initializePermissionStore() error {
	s.Logger().Debug("Loading permission store config")

	permissionStoreConfig := permissionMongo.NewConfig()
	if err := permissionStoreConfig.Load(s.ConfigReporter().WithScopes("permission", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load permission store config")
	}

	s.Logger().Debug("Creating permission store")

	permissionStore, err := permissionMongo.New(s.Logger(), permissionStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create permission store")
	}
	s.permissionStore = permissionStore

	return nil
}

func (s *Standard) initializeProfileStore() error {
	s.Logger().Debug("Loading profile store config")

	profileStoreConfig := baseMongo.NewConfig()
	if err := profileStoreConfig.Load(s.ConfigReporter().WithScopes("profile", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load profile store config")
	}

	s.Logger().Debug("Creating profile store")

	profileStore, err := profileMongo.New(s.Logger(), profileStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create profile store")
	}
	s.profileStore = profileStore

	return nil
}

func (s *Standard) initializeSessionStore() error {
	s.Logger().Debug("Loading session store config")

	sessionStoreConfig := baseMongo.NewConfig()
	if err := sessionStoreConfig.Load(s.ConfigReporter().WithScopes("session", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load session store config")
	}

	s.Logger().Debug("Creating session store")

	sessionStore, err := sessionMongo.New(s.Logger(), sessionStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create session store")
	}
	s.sessionStore = sessionStore

	return nil
}

func (s *Standard) initializeUserStore() error {
	s.Logger().Debug("Loading user store config")

	userStoreConfig := userMongo.NewConfig()
	if err := userStoreConfig.Load(s.ConfigReporter().WithScopes("user", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load user store config")
	}

	s.Logger().Debug("Creating user store")

	userStore, err := userMongo.New(s.Logger(), userStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user store")
	}
	s.userStore = userStore

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s.VersionReporter(), s.Logger(),
		s.AuthClient(), s.dataClient, s.metricClient, s.userClient,
		s.confirmationStore, s.messageStore, s.permissionStore, s.profileStore, s.sessionStore, s.userStore)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "service", "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(v1.Routes()); err != nil {
		return errors.Wrap(err, "service", "unable to initialize api router")
	}

	return nil
}

func (s *Standard) initializeServer() error {
	s.Logger().Debug("Loading server config")

	serverConfig := server.NewConfig()
	if err := serverConfig.Load(s.ConfigReporter().WithScopes(s.Name(), "server")); err != nil {
		return errors.Wrap(err, "service", "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	newServer, err := server.NewStandard(s.Logger(), s.api, serverConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create server")
	}
	s.server = newServer

	return nil
}
