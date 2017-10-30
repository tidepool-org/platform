package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/client"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/dexcom"
	dexcomClient "github.com/tidepool-org/platform/dexcom/client"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/queue"
	"github.com/tidepool-org/platform/task/service"
	"github.com/tidepool-org/platform/task/service/api"
	"github.com/tidepool-org/platform/task/service/api/v1"
	"github.com/tidepool-org/platform/task/store"
	taskMongo "github.com/tidepool-org/platform/task/store/mongo"
)

type Service struct {
	*serviceService.Authenticated
	taskStore    *taskMongo.Store
	taskClient   *Client
	dataClient   dataClient.Client
	dexcomClient dexcom.Client
	taskQueue    *queue.Queue
}

func New(prefix string) (*Service, error) {
	authenticated, err := serviceService.NewAuthenticated(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Authenticated: authenticated,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Authenticated.Initialize(); err != nil {
		return err
	}

	if err := s.initializeTaskStore(); err != nil {
		return err
	}
	if err := s.initializeTaskClient(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDexcomClient(); err != nil {
		return err
	}
	if err := s.initializeTaskQueue(); err != nil {
		return err
	}
	return s.initializeRouter()
}

func (s *Service) Terminate() {
	s.terminateRouter()
	s.terminateTaskQueue()
	s.terminateDexcomClient()
	s.terminateDataClient()
	s.terminateTaskClient()
	s.terminateTaskStore()

	s.Authenticated.Terminate()
}

func (s *Service) TaskStore() store.Store {
	return s.taskStore
}

func (s *Service) TaskClient() task.Client {
	return s.taskClient
}

func (s *Service) Status() *service.Status {
	return &service.Status{
		Version:   s.VersionReporter().Long(),
		TaskStore: s.TaskStore().Status(),
		Server:    s.API().Status(),
	}
}

func (s *Service) initializeTaskStore() error {
	s.Logger().Debug("Loading task store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("task", "store")); err != nil {
		return errors.Wrap(err, "unable to load task store config")
	}

	s.Logger().Debug("Creating task store")

	taskStore, err := taskMongo.New(cfg, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create task store")
	}
	s.taskStore = taskStore

	return nil
}

func (s *Service) terminateTaskStore() {
	if s.taskStore != nil {
		s.Logger().Debug("Closing task store")
		s.taskStore.Close()

		s.Logger().Debug("Destroying task store")
		s.taskStore = nil
	}
}

func (s *Service) initializeTaskClient() error {
	s.Logger().Debug("Creating task client")

	clnt, err := NewClient(s.TaskStore())
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

func (s *Service) initializeDexcomClient() error {
	s.Logger().Debug("Loading dexcom provider")

	dxcmPrvdr, err := dexcomProvider.New(s.ConfigReporter().WithScopes("provider"), s.dataClient, s.TaskClient())
	if err != nil {
		s.Logger().Warn("Unable to create dexcom provider")
	} else {
		s.Logger().Debug("Loading dexcom client config")

		cfg := client.NewConfig()
		if err = cfg.Load(s.ConfigReporter().WithScopes("dexcom", "client")); err != nil {
			return errors.Wrap(err, "unable to load dexcom client config")
		}

		s.Logger().Debug("Creating dexcom client")

		clnt, clntErr := dexcomClient.New(cfg, dxcmPrvdr)
		if clntErr != nil {
			return errors.Wrap(clntErr, "unable to create dexcom client")
		}
		s.dexcomClient = clnt
	}

	return nil
}

func (s *Service) terminateDexcomClient() {
	if s.dexcomClient != nil {
		s.Logger().Debug("Destroying dexcom client")
		s.dexcomClient = nil
	}
}

func (s *Service) initializeTaskQueue() error {
	s.Logger().Debug("Loading task queue config")

	cfg := queue.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("task", "queue")); err != nil {
		return errors.Wrap(err, "unable to load task queue config")
	}

	s.Logger().Debug("Creating task queue")

	taskQueue, err := queue.New(cfg, s.Logger(), s.TaskStore())
	if err != nil {
		return errors.Wrap(err, "unable to create task queue")
	}

	s.taskQueue = taskQueue

	if s.dexcomClient != nil {
		s.Logger().Debug("Creating dexcom fetch runner")

		rnnr, rnnrErr := dexcomFetch.NewRunner(s.Logger(), s.VersionReporter(), s.AuthClient(), s.dataClient, s.dexcomClient)
		if rnnrErr != nil {
			return errors.Wrap(rnnrErr, "unable to create dexcom fetch runner")
		}

		taskQueue.RegisterRunner(rnnr)
	}

	s.Logger().Debug("Starting task queue")

	s.taskQueue.Start()

	return nil
}

func (s *Service) terminateTaskQueue() {
	if s.taskQueue != nil {
		s.Logger().Debug("Stopping task queue")
		s.taskQueue.Stop()

		s.Logger().Debug("Destroying task queue")
		s.taskQueue = nil
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
