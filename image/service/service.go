package service

import (
	"context"

	awsSdkGoAwsSession "github.com/aws/aws-sdk-go/aws/session"

	"github.com/tidepool-org/platform/application"
	awsApi "github.com/tidepool-org/platform/aws/api"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	imageServiceApiV1 "github.com/tidepool-org/platform/image/service/api/v1"
	imageServiceClient "github.com/tidepool-org/platform/image/service/client"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreStructuredMongo "github.com/tidepool-org/platform/image/store/structured/mongo"
	imageStoreUnstructured "github.com/tidepool-org/platform/image/store/unstructured"
	imageTransform "github.com/tidepool-org/platform/image/transform"
	serviceApi "github.com/tidepool-org/platform/service/api"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeUnstructuredFactory "github.com/tidepool-org/platform/store/unstructured/factory"
)

type Service struct {
	*serviceService.Authenticated
	imageStructuredStore      *imageStoreStructuredMongo.Store
	imageUnstructuredStore    *imageStoreUnstructured.StoreImpl
	imageTransformer          *imageTransform.TransformerImpl
	imageMultipartFormDecoder *imageMultipart.FormDecoderImpl
	imageClient               *imageServiceClient.Client
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

	if err := s.initializeImageStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeImageUnstructuredStore(); err != nil {
		return err
	}
	if err := s.initializeImageTransformer(); err != nil {
		return err
	}
	if err := s.initializeImageMultipartFormDecoder(); err != nil {
		return err
	}
	if err := s.initializeImageClient(); err != nil {
		return err
	}
	return s.initializeRouter()
}

func (s *Service) Terminate() {
	s.terminateRouter()
	s.terminateImageClient()
	s.terminateImageMultipartFormDecoder()
	s.terminateImageTransformer()
	s.terminateImageUnstructuredStore()
	s.terminateImageStructuredStore()

	s.Authenticated.Terminate()
}

func (s *Service) Status(ctx context.Context) interface{} {
	return &status{
		Version: s.VersionReporter().Long(),
		Server:  s.API().Status(),
		Store:   s.imageStructuredStore.Status(ctx),
	}
}

func (s *Service) ImageStructuredStore() imageStoreStructured.Store {
	return s.imageStructuredStore
}

func (s *Service) ImageUnstructuredStore() imageStoreUnstructured.Store {
	return s.imageUnstructuredStore
}

func (s *Service) ImageTransformer() imageTransform.Transformer {
	return s.imageTransformer
}

func (s *Service) ImageMultipartFormDecoder() imageMultipart.FormDecoder {
	return s.imageMultipartFormDecoder
}

func (s *Service) ImageClient() image.Client {
	return s.imageClient
}

func (s *Service) initializeImageStructuredStore() error {
	s.Logger().Debug("Loading image structured store config")

	config := storeStructuredMongo.NewConfig(nil)
	if err := config.Load(); err != nil {
		return errors.Wrap(err, "unable to load image structured store config")
	}

	s.Logger().Debug("Creating image structured store")

	params := storeStructuredMongo.Params{DatabaseConfig: config}
	imageStructuredStore, err := imageStoreStructuredMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create image structured store")
	}
	s.imageStructuredStore = imageStructuredStore

	s.Logger().Debug("Ensuring image structured store indexes")

	err = s.imageStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure image structured store indexes")
	}

	return nil
}

func (s *Service) terminateImageStructuredStore() {
	if s.imageStructuredStore != nil {
		s.Logger().Debug("Closing image structured store")
		s.imageStructuredStore.Terminate(nil)

		s.Logger().Debug("Destroying image structured store")
		s.imageStructuredStore = nil
	}
}

func (s *Service) initializeImageUnstructuredStore() error {
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

	s.Logger().Debug("Creating image unstructured store")

	imageUnstructuredStore, err := imageStoreUnstructured.NewStore(unstructuredStore)
	if err != nil {
		return errors.Wrap(err, "unable to create image unstructured store")
	}
	s.imageUnstructuredStore = imageUnstructuredStore

	return nil
}

func (s *Service) terminateImageUnstructuredStore() {
	if s.imageUnstructuredStore != nil {
		s.Logger().Debug("Destroying image unstructured store")
		s.imageUnstructuredStore = nil
	}
}

func (s *Service) initializeImageTransformer() error {
	s.Logger().Debug("Creating image transformer")

	s.imageTransformer = imageTransform.NewTransformer()

	return nil
}

func (s *Service) terminateImageTransformer() {
	if s.imageTransformer != nil {
		s.Logger().Debug("Destroying image transformer")
		s.imageTransformer = nil
	}
}

func (s *Service) initializeImageMultipartFormDecoder() error {
	s.Logger().Debug("Creating image multipart form decoder")

	s.imageMultipartFormDecoder = imageMultipart.NewFormDecoder()

	return nil
}

func (s *Service) terminateImageMultipartFormDecoder() {
	if s.imageMultipartFormDecoder != nil {
		s.Logger().Debug("Destroying image multipart form decoder")
		s.imageMultipartFormDecoder = nil
	}
}

func (s *Service) initializeImageClient() error {
	s.Logger().Debug("Creating image client")

	client, err := imageServiceClient.New(s)
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

func (s *Service) initializeRouter() error {
	s.Logger().Debug("Creating status router")

	statusRouter, err := serviceApi.NewStatusRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create status router")
	}

	s.Logger().Debug("Creating image service api v1 router")

	router, err := imageServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create image service api v1 router")
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
