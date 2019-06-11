package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/blob"
	blobClient "github.com/tidepool-org/platform/blob/client"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	confirmationStoreMongo "github.com/tidepool-org/platform/confirmation/store/mongo"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceClient "github.com/tidepool-org/platform/data/source/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageClient "github.com/tidepool-org/platform/image/client"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	messageStore "github.com/tidepool-org/platform/message/store"
	messageStoreMongo "github.com/tidepool-org/platform/message/store/mongo"
	"github.com/tidepool-org/platform/metric"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	permissionStoreMongo "github.com/tidepool-org/platform/permission/store/mongo"
	"github.com/tidepool-org/platform/platform"
	profileStoreStructured "github.com/tidepool-org/platform/profile/store/structured"
	profileStoreStructuredMongo "github.com/tidepool-org/platform/profile/store/structured/mongo"
	serviceApi "github.com/tidepool-org/platform/service/api"
	serviceService "github.com/tidepool-org/platform/service/service"
	sessionStore "github.com/tidepool-org/platform/session/store"
	sessionStoreMongo "github.com/tidepool-org/platform/session/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"
	userServiceApiV1 "github.com/tidepool-org/platform/user/service/api/v1"
	userServiceClient "github.com/tidepool-org/platform/user/service/client"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
	userStoreStructuredMongo "github.com/tidepool-org/platform/user/store/structured/mongo"
)

// TODO: We really should not have direct access of these other stores, but short
// of implementing a master delete in each of the legacy services or creating six brand
// new services, this is a reasonable HACK. Once the fates of the legacy services are
// determined then the necessary changes can be made.

type Service struct {
	*serviceService.Authenticated
	blobClient          *blobClient.Client
	dataClient          *dataClient.ClientImpl
	dataSourceClient    *dataSourceClient.Client
	imageClient         *imageClient.Client
	metricClient        *metricClient.Client
	permissionClient    *permissionClient.Client
	confirmationStore   *confirmationStoreMongo.Store
	messageStore        *messageStoreMongo.Store
	permissionStore     *permissionStoreMongo.Store
	profileStore        *profileStoreStructuredMongo.Store
	sessionStore        *sessionStoreMongo.Store
	userStructuredStore *userStoreStructuredMongo.Store
	passwordHasher      *PasswordHasher
	userClient          *userServiceClient.Client
}

func New() *Service {
	return &Service{
		Authenticated: serviceService.NewAuthenticated(),
	}
}

func (s *Service) Initialize(provider application.Provider) error {
	if err := s.Authenticated.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeBlobClient(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
		return err
	}
	if err := s.initializeImageClient(); err != nil {
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
	if err := s.initializeUserStructuredStore(); err != nil {
		return err
	}
	if err := s.initializePasswordHasher(); err != nil {
		return err
	}
	if err := s.initializeUserClient(); err != nil {
		return err
	}
	return s.initializeRouter()
}

func (s *Service) Terminate() {
	s.terminateRouter()
	s.terminateUserClient()
	s.terminatePasswordHasher()
	s.terminateUserStructuredStore()
	s.terminateSessionStore()
	s.terminateProfileStore()
	s.terminatePermissionStore()
	s.terminateMessageStore()
	s.terminateConfirmationStore()
	s.terminatePermissionClient()
	s.terminateMetricClient()
	s.terminateImageClient()
	s.terminateDataSourceClient()
	s.terminateDataClient()
	s.terminateBlobClient()

	s.Authenticated.Terminate()
}

func (s *Service) Status() interface{} {
	return &status{
		Version: s.VersionReporter().Long(),
		Server:  s.API().Status(),
		Store:   s.userStructuredStore.Status(),
	}
}

func (s *Service) BlobClient() blob.Client {
	return s.blobClient
}

func (s *Service) DataClient() dataClient.Client {
	return s.dataClient
}

func (s *Service) DataSourceClient() dataSource.Client {
	return s.dataSourceClient
}

func (s *Service) ImageClient() image.Client {
	return s.imageClient
}

func (s *Service) MetricClient() metric.Client {
	return s.metricClient
}

func (s *Service) PermissionClient() permission.Client {
	return s.permissionClient
}

func (s *Service) ConfirmationStore() confirmationStore.Store {
	return s.confirmationStore
}

func (s *Service) MessageStore() messageStore.Store {
	return s.messageStore
}

func (s *Service) PermissionStore() permissionStore.Store {
	return s.permissionStore
}

func (s *Service) ProfileStore() profileStoreStructured.Store {
	return s.profileStore
}

func (s *Service) SessionStore() sessionStore.Store {
	return s.sessionStore
}

func (s *Service) UserStructuredStore() userStoreStructured.Store {
	return s.userStructuredStore
}

func (s *Service) PasswordHasher() userServiceClient.PasswordHasher {
	return s.passwordHasher
}

func (s *Service) UserClient() user.Client {
	return s.userClient
}

func (s *Service) initializeBlobClient() error {
	s.Logger().Debug("Loading blob client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("blob", "client")); err != nil {
		return errors.Wrap(err, "unable to load blob client config")
	}

	s.Logger().Debug("Creating blob client")

	client, err := blobClient.New(config, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create blob client")
	}
	s.blobClient = client

	return nil
}

func (s *Service) terminateBlobClient() {
	if s.blobClient != nil {
		s.Logger().Debug("Destroying blob client")
		s.blobClient = nil
	}
}

func (s *Service) initializeDataClient() error {
	s.Logger().Debug("Loading data client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("data", "client")); err != nil {
		return errors.Wrap(err, "unable to load data client config")
	}

	s.Logger().Debug("Creating data client")

	client, err := dataClient.New(config, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = client

	return nil
}

func (s *Service) terminateDataClient() {
	if s.dataClient != nil {
		s.Logger().Debug("Destroying data client")
		s.dataClient = nil
	}
}

func (s *Service) initializeDataSourceClient() error {
	s.Logger().Debug("Loading data source client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("data_source", "client")); err != nil {
		return errors.Wrap(err, "unable to load data source client config")
	}

	s.Logger().Debug("Creating data source client")

	client, err := dataSourceClient.New(config, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create data source client")
	}
	s.dataSourceClient = client

	return nil
}

func (s *Service) terminateDataSourceClient() {
	if s.dataSourceClient != nil {
		s.Logger().Debug("Destroying data source client")
		s.dataSourceClient = nil
	}
}

func (s *Service) initializeImageClient() error {
	s.Logger().Debug("Loading image client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("image", "client")); err != nil {
		return errors.Wrap(err, "unable to load image client config")
	}

	s.Logger().Debug("Creating image client")

	client, err := imageClient.New(config, platform.AuthorizeAsService, imageMultipart.NewFormEncoder())
	if err != nil {
		return errors.Wrap(err, "unable to create image client")
	}
	s.imageClient = client

	return nil
}

func (s *Service) terminateImageClient() {
	if s.imageClient != nil {
		s.Logger().Debug("Destroying image client")
		s.imageClient = nil
	}
}

func (s *Service) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("metric", "client")); err != nil {
		return errors.Wrap(err, "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	client, err := metricClient.New(config, platform.AuthorizeAsUser, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "unable to create metric client")
	}
	s.metricClient = client

	return nil
}

func (s *Service) terminateMetricClient() {
	if s.metricClient != nil {
		s.Logger().Debug("Destroying metric client")
		s.metricClient = nil
	}
}

func (s *Service) initializePermissionClient() error {
	s.Logger().Debug("Loading permission client config")

	config := platform.NewConfig()
	config.UserAgent = s.UserAgent()
	if err := config.Load(s.ConfigReporter().WithScopes("permission", "client")); err != nil {
		return errors.Wrap(err, "unable to load permission client config")
	}

	s.Logger().Debug("Creating permission client")

	client, err := permissionClient.New(config, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create permission client")
	}
	s.permissionClient = client

	return nil
}

func (s *Service) terminatePermissionClient() {
	if s.permissionClient != nil {
		s.Logger().Debug("Destroying permission client")
		s.permissionClient = nil
	}
}

func (s *Service) initializeConfirmationStore() error {
	s.Logger().Debug("Loading confirmation store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("confirmation", "store")); err != nil {
		return errors.Wrap(err, "unable to load confirmation store config")
	}

	s.Logger().Debug("Creating confirmation store")

	store, err := confirmationStoreMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create confirmation store")
	}
	s.confirmationStore = store

	return nil
}

func (s *Service) terminateConfirmationStore() {
	if s.confirmationStore != nil {
		s.Logger().Debug("Closing confirmation store")
		s.confirmationStore.Close()

		s.Logger().Debug("Destroying confirmation store")
		s.confirmationStore = nil
	}
}

func (s *Service) initializeMessageStore() error {
	s.Logger().Debug("Loading message store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("message", "store")); err != nil {
		return errors.Wrap(err, "unable to load message store config")
	}

	s.Logger().Debug("Creating message store")

	store, err := messageStoreMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create message store")
	}
	s.messageStore = store

	return nil
}

func (s *Service) terminateMessageStore() {
	if s.messageStore != nil {
		s.Logger().Debug("Closing message store")
		s.messageStore.Close()

		s.Logger().Debug("Destroying message store")
		s.messageStore = nil
	}
}

func (s *Service) initializePermissionStore() error {
	s.Logger().Debug("Loading permission store config")

	config := permissionStoreMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("permission", "store")); err != nil {
		return errors.Wrap(err, "unable to load permission store config")
	}

	s.Logger().Debug("Creating permission store")

	store, err := permissionStoreMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create permission store")
	}
	s.permissionStore = store

	return nil
}

func (s *Service) terminatePermissionStore() {
	if s.permissionStore != nil {
		s.Logger().Debug("Closing permission store")
		s.permissionStore.Close()

		s.Logger().Debug("Destroying permission store")
		s.permissionStore = nil
	}
}

func (s *Service) initializeProfileStore() error {
	s.Logger().Debug("Loading profile store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("profile", "store")); err != nil {
		return errors.Wrap(err, "unable to load profile store config")
	}

	s.Logger().Debug("Creating profile store")

	store, err := profileStoreStructuredMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create profile store")
	}
	s.profileStore = store

	return nil
}

func (s *Service) terminateProfileStore() {
	if s.profileStore != nil {
		s.Logger().Debug("Closing profile store")
		s.profileStore.Close()

		s.Logger().Debug("Destroying profile store")
		s.profileStore = nil
	}
}

func (s *Service) initializeSessionStore() error {
	s.Logger().Debug("Loading session store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("session", "store")); err != nil {
		return errors.Wrap(err, "unable to load session store config")
	}

	s.Logger().Debug("Creating session store")

	store, err := sessionStoreMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create session store")
	}
	s.sessionStore = store

	return nil
}

func (s *Service) terminateSessionStore() {
	if s.sessionStore != nil {
		s.Logger().Debug("Closing session store")
		s.sessionStore.Close()

		s.Logger().Debug("Destroying session store")
		s.sessionStore = nil
	}
}

func (s *Service) initializeUserStructuredStore() error {
	s.Logger().Debug("Loading user structured store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("user", "store")); err != nil {
		return errors.Wrap(err, "unable to load user structured store config")
	}

	s.Logger().Debug("Creating user structured store")

	userStructuredStore, err := userStoreStructuredMongo.NewStore(config, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create user structured store")
	}
	s.userStructuredStore = userStructuredStore

	return nil
}

func (s *Service) terminateUserStructuredStore() {
	if s.userStructuredStore != nil {
		s.Logger().Debug("Closing user structured store")
		s.userStructuredStore.Close()

		s.Logger().Debug("Destroying user structured store")
		s.userStructuredStore = nil
	}
}

func (s *Service) initializePasswordHasher() error {
	s.Logger().Debug("Loading password hasher config")

	config := NewPasswordHasherConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("user", "store", "password")); err != nil {
		return errors.Wrap(err, "unable to load password hasher config")
	}

	s.Logger().Debug("Creating password hasher")

	passwordHasher, err := NewPasswordHasher(config)
	if err != nil {
		return errors.Wrap(err, "unable to create password hasher")
	}
	s.passwordHasher = passwordHasher

	return nil
}

func (s *Service) terminatePasswordHasher() {
	if s.passwordHasher != nil {
		s.Logger().Debug("Destroying password hasher")
		s.passwordHasher = nil
	}
}

func (s *Service) initializeUserClient() error {
	s.Logger().Debug("Creating user client")

	client, err := userServiceClient.New(s)
	if err != nil {
		return errors.Wrap(err, "unable to create user client")
	}
	s.userClient = client

	return nil
}

func (s *Service) terminateUserClient() {
	if s.userClient != nil {
		s.Logger().Debug("Destroying user client")
		s.userClient = nil
	}
}

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating status router")

	statusRouter, err := serviceApi.NewStatusRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create status router")
	}

	s.Logger().Debug("Creating user service api v1 router")

	router, err := userServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create user service api v1 router")
	}

	s.Logger().Debug("Initializing routers")

	if err = s.API().InitializeRouters(statusRouter, router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (s *Service) terminateRouter() {
}

type status struct {
	Version string      `json:"version,omitempty"`
	Server  interface{} `json:"server,omitempty"`
	Store   interface{} `json:"store,omitempty"`
}
