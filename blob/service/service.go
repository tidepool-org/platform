package service

import (
	"context"

	awsSdkGoAwsSession "github.com/aws/aws-sdk-go/aws/session"
	eventsCommon "github.com/tidepool-org/go-common/events"

	blobEvents "github.com/tidepool-org/platform/blob/events"
	"github.com/tidepool-org/platform/events"
	logInternal "github.com/tidepool-org/platform/log"

	"github.com/tidepool-org/platform/application"
	awsApi "github.com/tidepool-org/platform/aws/api"
	"github.com/tidepool-org/platform/blob"
	blobServiceApiV1 "github.com/tidepool-org/platform/blob/service/api/v1"
	blobServiceClient "github.com/tidepool-org/platform/blob/service/client"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	"github.com/tidepool-org/platform/errors"
	serviceApi "github.com/tidepool-org/platform/service/api"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeUnstructuredFactory "github.com/tidepool-org/platform/store/unstructured/factory"
)

type Service struct {
	*serviceService.Authenticated
	blobStructuredStore         *blobStoreStructuredMongo.Store
	blobUnstructuredStore       *blobStoreUnstructured.StoreImpl
	deviceLogsUnstructuredStore *blobStoreUnstructured.StoreImpl
	blobClient                  *blobServiceClient.Client
	userEventsHandler           events.Runner
}

func New() *Service {
	return &Service{
		Authenticated: serviceService.NewAuthenticated(),
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
	if err := s.Authenticated.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeBlobStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeBlobUnstructuredStore(); err != nil {
		return err
	}
	if err := s.initializeDeviceLogsUnstructuredStore(); err != nil {
		return err
	}
	if err := s.initializeBlobClient(); err != nil {
		return err
	}
	if err := s.initializeUserEventsHandler(); err != nil {
		return err
	}
	return s.initializeRouter()
}

func (s *Service) Terminate() {
	s.Authenticated.Terminate()
	s.terminateUserEventsHandler()
	s.terminateRouter()
	s.terminateBlobClient()
	s.terminateBlobUnstructuredStore()
	s.terminateDeviceLogsUnstructuredStore()
	s.terminateBlobStructuredStore()
}

func (s *Service) Status(ctx context.Context) interface{} {
	return &status{
		Version: s.VersionReporter().Long(),
	}
}

func (s *Service) BlobStructuredStore() blobStoreStructured.Store {
	return s.blobStructuredStore
}

func (s *Service) BlobUnstructuredStore() blobStoreUnstructured.Store {
	return s.blobUnstructuredStore
}

func (s *Service) DeviceLogsUnstructuredStore() blobStoreUnstructured.Store {
	return s.deviceLogsUnstructuredStore
}

func (s *Service) BlobClient() blob.Client {
	return s.blobClient
}

func (s *Service) initializeBlobStructuredStore() error {
	s.Logger().Debug("Loading blob structured store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(); err != nil {
		return errors.Wrap(err, "unable to load blob structured store config")
	}

	s.Logger().Debug("Creating blob structured store")

	blobStructuredStore, err := blobStoreStructuredMongo.NewStore(config)
	if err != nil {
		return errors.Wrap(err, "unable to create blob structured store")
	}
	s.blobStructuredStore = blobStructuredStore

	s.Logger().Debug("Ensuring blob structured store indexes")

	err = s.blobStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure blob structured store indexes")
	}

	return nil
}

func (s *Service) terminateBlobStructuredStore() {
	if s.blobStructuredStore != nil {
		s.Logger().Debug("Closing blob structured store")
		s.blobStructuredStore.Terminate(context.Background())

		s.Logger().Debug("Destroying blob structured store")
		s.blobStructuredStore = nil
	}
}

func (s *Service) getAWSUnstructuredStore(bucketGroup *string) (*blobStoreUnstructured.StoreImpl, error) {
	s.Logger().Debug("Creating aws session")

	session, err := awsSdkGoAwsSession.NewSession() // FUTURE: Session pooling
	if err != nil {
		return nil, errors.Wrap(err, "unable to create aws session")
	}

	s.Logger().Debug("Creating aws session")

	api, err := awsApi.New(session)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create aws api")
	}

	s.Logger().Debug("Creating unstructured store")

	configReporter := s.ConfigReporter().WithScopes("unstructured", "store", *bucketGroup)
	unstructuredStore, err := storeUnstructuredFactory.NewStore(configReporter, api)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create unstructured store")
	}

	s.Logger().Debug("Creating blob unstructured store")

	blobUnstructuredStore, err := blobStoreUnstructured.NewStore(unstructuredStore)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create blob unstructured store")
	}

	return blobUnstructuredStore, nil
}

func (s *Service) initializeBlobUnstructuredStore() error {
	blobs := "blobs"
	store, err := s.getAWSUnstructuredStore(&blobs)
	if err != nil {
		return err
	}
	s.blobUnstructuredStore = store
	return nil
}

func (s *Service) initializeDeviceLogsUnstructuredStore() error {
	deviceLogs := "device_logs"
	store, err := s.getAWSUnstructuredStore(&deviceLogs)
	if err != nil {
		return err
	}
	s.deviceLogsUnstructuredStore = store
	return nil
}

func (s *Service) terminateBlobUnstructuredStore() {
	if s.blobUnstructuredStore != nil {
		s.Logger().Debug("Destroying blob unstructured store")
		s.blobUnstructuredStore = nil
	}
}

func (s *Service) terminateDeviceLogsUnstructuredStore() {
	if s.deviceLogsUnstructuredStore != nil {
		s.Logger().Debug("Destroying device logs unstructured store")
		s.deviceLogsUnstructuredStore = nil
	}
}

func (s *Service) initializeUserEventsHandler() error {
	s.Logger().Debug("Initializing user events handler")

	ctx := logInternal.NewContextWithLogger(context.Background(), s.Logger())
	handler := blobEvents.NewUserDataDeletionHandler(ctx, s.blobClient)
	handlers := []eventsCommon.EventHandler{handler}
	runner := events.NewRunner(handlers)

	if err := runner.Initialize(); err != nil {
		return errors.Wrap(err, "unable to initialize events runner")
	}
	s.userEventsHandler = runner

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

func (s *Service) initializeBlobClient() error {
	s.Logger().Debug("Creating blob client")

	client, err := blobServiceClient.New(s)
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

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating status router")

	statusRouter, err := serviceApi.NewStatusRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create status router")
	}

	s.Logger().Debug("Creating blob service api v1 router")

	router, err := blobServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create blob service api v1 router")
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
	Version string
	Server  interface{}
	Store   interface{}
}
