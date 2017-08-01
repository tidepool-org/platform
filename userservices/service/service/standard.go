package service

import (
	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	"github.com/tidepool-org/platform/errors"
	messageMongo "github.com/tidepool-org/platform/message/store/mongo"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	notificationMongo "github.com/tidepool-org/platform/notification/store/mongo"
	permissionMongo "github.com/tidepool-org/platform/permission/store/mongo"
	profileMongo "github.com/tidepool-org/platform/profile/store/mongo"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	sessionMongo "github.com/tidepool-org/platform/session/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	userMongo "github.com/tidepool-org/platform/user/store/mongo"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service/api"
	"github.com/tidepool-org/platform/userservices/service/api/v1"
)

type Standard struct {
	*service.Standard
	metricServicesClient *metricservicesClient.Standard
	userServicesClient   *userservicesClient.Standard
	dataServicesClient   *dataservicesClient.Standard
	messageStore         *messageMongo.Store
	notificationStore    *notificationMongo.Store
	permissionStore      *permissionMongo.Store
	profileStore         *profileMongo.Store
	sessionStore         *sessionMongo.Store
	userStore            *userMongo.Store
	userServicesAPI      *api.Standard
	userServicesServer   *server.Standard
}

func NewStandard() (*Standard, error) {
	standard, err := service.NewStandard("userservices", "TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard: standard,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.Standard.Initialize(); err != nil {
		return err
	}

	if err := s.initializeMetricServicesClient(); err != nil {
		return err
	}
	if err := s.initializeUserServicesClient(); err != nil {
		return err
	}
	if err := s.initializeDataServicesClient(); err != nil {
		return err
	}
	if err := s.initializeMessageStore(); err != nil {
		return err
	}
	if err := s.initializeNotificationStore(); err != nil {
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
	if err := s.initializeUserServicesAPI(); err != nil {
		return err
	}
	if err := s.initializeUserServicesServer(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.userServicesServer = nil
	s.userServicesAPI = nil
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
	if s.notificationStore != nil {
		s.notificationStore.Close()
		s.notificationStore = nil
	}
	if s.messageStore != nil {
		s.messageStore.Close()
		s.messageStore = nil
	}
	s.dataServicesClient = nil
	if s.userServicesClient != nil {
		s.userServicesClient.Close()
		s.userServicesClient = nil
	}
	s.metricServicesClient = nil

	s.Standard.Terminate()
}

func (s *Standard) Run() error {
	if s.userServicesServer == nil {
		return errors.New("service", "service not initialized")
	}

	return s.userServicesServer.Serve()
}

func (s *Standard) initializeMetricServicesClient() error {
	s.Logger().Debug("Loading metric services client config")

	metricServicesClientConfig := metricservicesClient.NewConfig()
	if err := metricServicesClientConfig.Load(s.ConfigReporter().WithScopes("metricservices", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load metric services client config")
	}

	s.Logger().Debug("Creating metric services client")

	metricServicesClient, err := metricservicesClient.NewStandard(s.VersionReporter(), s.Name(), metricServicesClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create metric services client")
	}
	s.metricServicesClient = metricServicesClient

	return nil
}

func (s *Standard) initializeUserServicesClient() error {
	s.Logger().Debug("Loading user services client config")

	userServicesClientConfig := userservicesClient.NewConfig()
	if err := userServicesClientConfig.Load(s.ConfigReporter().WithScopes("userservices", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load user services client config")
	}

	s.Logger().Debug("Creating user services client")

	userServicesClient, err := userservicesClient.NewStandard(s.Logger(), s.Name(), userServicesClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user services client")
	}
	s.userServicesClient = userServicesClient

	s.Logger().Debug("Starting user services client")

	if err = s.userServicesClient.Start(); err != nil {
		return errors.Wrap(err, "service", "unable to start user services client")
	}

	return nil
}

func (s *Standard) initializeDataServicesClient() error {
	s.Logger().Debug("Loading data services client config")

	dataServicesClientConfig := dataservicesClient.NewConfig()
	if err := dataServicesClientConfig.Load(s.ConfigReporter().WithScopes("dataservices", "client")); err != nil {
		return errors.Wrap(err, "service", "unable to load data services client config")
	}

	s.Logger().Debug("Creating data services client")

	dataServicesClient, err := dataservicesClient.NewStandard(dataServicesClientConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create data services client")
	}
	s.dataServicesClient = dataServicesClient

	return nil
}

func (s *Standard) initializeMessageStore() error {
	s.Logger().Debug("Loading message store config")

	messageStoreConfig := baseMongo.NewConfig()
	if err := messageStoreConfig.Load(s.ConfigReporter().WithScopes("message", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load message store config")
	}
	messageStoreConfig.Collection = "messages"

	s.Logger().Debug("Creating message store")

	messageStore, err := messageMongo.New(s.Logger(), messageStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create message store")
	}
	s.messageStore = messageStore

	return nil
}

func (s *Standard) initializeNotificationStore() error {
	s.Logger().Debug("Loading notification store config")

	notificationStoreConfig := baseMongo.NewConfig()
	if err := notificationStoreConfig.Load(s.ConfigReporter().WithScopes("notification", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load notification store config")
	}
	notificationStoreConfig.Collection = "confirmations"

	s.Logger().Debug("Creating notification store")

	notificationStore, err := notificationMongo.New(s.Logger(), notificationStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create notification store")
	}
	s.notificationStore = notificationStore

	return nil
}

func (s *Standard) initializePermissionStore() error {
	s.Logger().Debug("Loading permission store config")

	permissionStoreConfig := permissionMongo.NewConfig()
	if err := permissionStoreConfig.Load(s.ConfigReporter().WithScopes("permission", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load permission store config")
	}
	permissionStoreConfig.Collection = "perms"

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
	profileStoreConfig.Collection = "seagull"

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
	sessionStoreConfig.Collection = "tokens"

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
	userStoreConfig.Collection = "users"

	s.Logger().Debug("Creating user store")

	userStore, err := userMongo.New(s.Logger(), userStoreConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user store")
	}
	s.userStore = userStore

	return nil
}

func (s *Standard) initializeUserServicesAPI() error {
	s.Logger().Debug("Creating user services api")

	userServicesAPI, err := api.NewStandard(s.VersionReporter(), s.EnvironmentReporter(), s.Logger(),
		s.metricServicesClient, s.userServicesClient, s.dataServicesClient,
		s.messageStore, s.notificationStore, s.permissionStore, s.profileStore, s.sessionStore, s.userStore)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user services api")
	}
	s.userServicesAPI = userServicesAPI

	s.Logger().Debug("Initializing user services api middleware")

	if err = s.userServicesAPI.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "service", "unable to initialize user services api middleware")
	}

	s.Logger().Debug("Configuring user services api middleware headers")

	s.userServicesAPI.HeaderMiddleware().AddHeaderFieldFunc(
		userservicesClient.TidepoolAuthenticationTokenHeaderName, middleware.NewMD5FieldFunc("authenticationTokenMD5"))

	s.Logger().Debug("Initializing user services api router")

	if err = s.userServicesAPI.InitializeRouter(v1.Routes()); err != nil {
		return errors.Wrap(err, "service", "unable to initialize user services api router")
	}

	return nil
}

func (s *Standard) initializeUserServicesServer() error {
	s.Logger().Debug("Loading user services server config")

	userServicesServerConfig := server.NewConfig()
	if err := userServicesServerConfig.Load(s.ConfigReporter().WithScopes("userservices", "server")); err != nil {
		return errors.Wrap(err, "service", "unable to load user services server config")
	}

	s.Logger().Debug("Creating user services server")

	userServicesServer, err := server.NewStandard(s.Logger(), s.userServicesAPI, userServicesServerConfig)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create user services server")
	}
	s.userServicesServer = userServicesServer

	return nil
}
