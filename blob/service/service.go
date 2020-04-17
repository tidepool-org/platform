package service

import (
	awsSdkGoAwsSession "github.com/aws/aws-sdk-go/aws/session"

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
	blobStructuredStore   *blobStoreStructuredMongo.Store
	blobUnstructuredStore *blobStoreUnstructured.StoreImpl
	blobClient            *blobServiceClient.Client
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

	if err := s.initializeBlobStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeBlobUnstructuredStore(); err != nil {
		return err
	}
	if err := s.initializeBlobClient(); err != nil {
		return err
	}
	return s.initializeRouter()
}

func (s *Service) Terminate() {
	s.terminateRouter()
	s.terminateBlobClient()
	s.terminateBlobUnstructuredStore()
	s.terminateBlobStructuredStore()

	s.Authenticated.Terminate()
}

func (s *Service) Status() interface{} {
	return &status{
		Version: s.VersionReporter().Long(),
		Server:  s.API().Status(),
		Store:   s.blobStructuredStore.Status(),
	}
}

func (s *Service) BlobStructuredStore() blobStoreStructured.Store {
	return s.blobStructuredStore
}

func (s *Service) BlobUnstructuredStore() blobStoreUnstructured.Store {
	return s.blobUnstructuredStore
}

func (s *Service) BlobClient() blob.Client {
	return s.blobClient
}

func (s *Service) initializeBlobStructuredStore() error {
	s.Logger().Debug("Loading blob structured store config")

	config := storeStructuredMongo.NewConfig()
	if err := config.Load(s.ConfigReporter().WithScopes("structured", "store")); err != nil {
		return errors.Wrap(err, "unable to load blob structured store config")
	}

	s.Logger().Debug("Creating blob structured store")

	blobStructuredStore, err := blobStoreStructuredMongo.NewStore(config, s.Logger())
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
		s.blobStructuredStore.Close()

		s.Logger().Debug("Destroying blob structured store")
		s.blobStructuredStore = nil
	}
}

func (s *Service) initializeBlobUnstructuredStore() error {
	s.Logger().Debug("Creating aws session")

	session, err := awsSdkGoAwsSession.NewSession() // FUTURE: Session pooling
	if err != nil {
		return errors.Wrap(err, "unable to create aws session")
	}

	s.Logger().Debug("Creating aws session")

	api, err := awsApi.New(session)
	if err != nil {
		return errors.Wrap(err, "unable to create aws api")
	}

	s.Logger().Debug("Creating unstructured store")

	unstructuredStore, err := storeUnstructuredFactory.NewStore(s.ConfigReporter().WithScopes("unstructured", "store"), api)
	if err != nil {
		return errors.Wrap(err, "unable to create unstructured store")
	}

	s.Logger().Debug("Creating blob unstructured store")

	blobUnstructuredStore, err := blobStoreUnstructured.NewStore(unstructuredStore)
	if err != nil {
		return errors.Wrap(err, "unable to create blob unstructured store")
	}
	s.blobUnstructuredStore = blobUnstructuredStore

	return nil
}

func (s *Service) terminateBlobUnstructuredStore() {
	if s.blobUnstructuredStore != nil {
		s.Logger().Debug("Destroying blob unstructured store")
		s.blobUnstructuredStore = nil
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
