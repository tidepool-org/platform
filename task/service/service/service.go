package service

import (
	"context"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/clinics"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceClient "github.com/tidepool-org/platform/data/source/client"
	"github.com/tidepool-org/platform/dexcom"
	dexcomClient "github.com/tidepool-org/platform/dexcom/client"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	dexcomProvider "github.com/tidepool-org/platform/dexcom/provider"
	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	serviceService "github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	summaryTask "github.com/tidepool-org/platform/summary/task"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/queue"
	"github.com/tidepool-org/platform/task/service"
	"github.com/tidepool-org/platform/task/service/api"
	taskServiceApiV1 "github.com/tidepool-org/platform/task/service/api/v1"
	"github.com/tidepool-org/platform/task/store"
	taskMongo "github.com/tidepool-org/platform/task/store/mongo"
)

type Service struct {
	*serviceService.Authenticated
	taskStore        store.Store
	taskClient       *Client
	dataClient       dataClient.Client
	dataSourceClient dataSource.Client
	dexcomClient     dexcom.Client
	taskQueue        queue.Queue
	clinicsClient    clinics.Client
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

	if err := s.initializeTaskStore(); err != nil {
		return err
	}
	if err := s.initializeTaskClient(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
		return err
	}
	if err := s.initializeDexcomClient(); err != nil {
		return err
	}
	if err := s.initializeClinicsClient(); err != nil {
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
	s.terminateClinicsClient()
	s.terminateDexcomClient()
	s.terminateDataSourceClient()
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

func (s *Service) Status(ctx context.Context) *service.Status {
	return &service.Status{
		Version: s.VersionReporter().Long(),
	}
}

func (s *Service) initializeTaskStore() error {
	s.Logger().Debug("Loading task store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load task store config")
	}

	s.Logger().Debug("Creating task store")

	taskStore, err := taskMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create task store")
	}
	s.taskStore = taskStore

	s.Logger().Debug("Ensuring task store indexes")

	err = taskStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure task store indexes")
	}

	err = taskStore.EnsureDefaultTasks()
	if err != nil {
		return errors.Wrap(err, "unable to ensure task store contains default tasks")
	}

	return nil
}

func (s *Service) terminateTaskStore() {
	if s.taskStore != nil {
		s.Logger().Debug("Closing task store")
		s.taskStore.Terminate(context.Background())

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
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("data", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
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

func (s *Service) terminateDataClient() {
	if s.dataClient != nil {
		s.Logger().Debug("Destroying data client")
		s.dataClient = nil
	}
}

func (s *Service) initializeDataSourceClient() error {
	s.Logger().Debug("Loading data source client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("data_source", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load data source client config")
	}

	s.Logger().Debug("Creating data source client")

	clnt, err := dataSourceClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create data source client")
	}
	s.dataSourceClient = clnt

	return nil
}

func (s *Service) terminateDataSourceClient() {
	if s.dataSourceClient != nil {
		s.Logger().Debug("Destroying data source client")
		s.dataSourceClient = nil
	}
}

func (s *Service) initializeDexcomClient() error {
	s.Logger().Debug("Loading dexcom provider")

	dxcmPrvdr, err := dexcomProvider.New(s.ConfigReporter().WithScopes("provider"), s.dataSourceClient, s.TaskClient())
	if err != nil {
		s.Logger().Warn("Unable to create dexcom provider")
	} else {
		s.Logger().Debug("Loading dexcom client config")

		cfg := client.NewConfig()
		cfg.UserAgent = s.UserAgent()
		reporter := s.ConfigReporter().WithScopes("dexcom", "client")
		loader := client.NewConfigReporterLoader(reporter)
		if err = cfg.Load(loader); err != nil {
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

func (s *Service) initializeClinicsClient() error {
	s.Logger().Debug("Creating clinics client")

	clnt, err := clinics.NewClient(s.AuthClient())
	if err != nil {
		return errors.Wrap(err, "unable to create clinics client")
	}
	s.clinicsClient = clnt

	return nil
}

func (s *Service) terminateClinicsClient() {
	if s.clinicsClient != nil {
		s.Logger().Debug("Destroying clinics client")
		s.clinicsClient = nil
	}
}

func (s *Service) initializeTaskQueue() error {
	s.Logger().Debug("Loading task queue config")

	cfg := queue.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("task", "queue")); err != nil {
		return errors.Wrap(err, "unable to load task queue config")
	}

	s.Logger().Debug("Creating task queue")

	taskQueue, err := queue.NewMultiQueue(cfg, s.Logger(), s.TaskStore())
	if err != nil {
		return errors.Wrap(err, "unable to create task queue")
	}

	s.taskQueue = taskQueue

	var runners []queue.Runner

	if s.dexcomClient != nil {
		s.Logger().Debug("Creating dexcom fetch runner")

		rnnr, rnnrErr := dexcomFetch.NewRunner(s.AuthClient(), s.dataClient, s.dataSourceClient, s.dexcomClient)
		if rnnrErr != nil {
			return errors.Wrap(rnnrErr, "unable to create dexcom fetch runner")
		}

		runners = append(runners, rnnr)
	}

	summaryRunners, err := summaryTask.NewSummaryRunners(s.AuthClient(), s.dataClient, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create summary runners")
	}
	runners = append(runners, summaryRunners...)

	s.Logger().Debug("Creating ehr reconcile runner")

	ehrReconcileRnnr, err := reconcile.NewRunner(s.AuthClient(), s.clinicsClient, s.taskClient, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create ehr reconcile runner")
	}
	runners = append(runners, ehrReconcileRnnr)

	s.Logger().Debug("Creating ehr sync runner")

	ehrSyncRnnr, err := sync.NewRunner(s.clinicsClient, s.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create ehr sync runner")
	}
	runners = append(runners, ehrSyncRnnr)

	for _, r := range runners {
		r := r
		if err := taskQueue.RegisterRunner(r); err != nil {
			return errors.Wrapf(err, "unable to register runner %s", r.GetRunnerType())
		}
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
	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create api router")
	}

	s.Logger().Debug("Creating v1 router")

	v1Router, err := taskServiceApiV1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "unable to create v1 router")
	}

	s.Logger().Debug("Initializing routers")

	if err = s.API().InitializeRouters(apiRouter, v1Router); err != nil {
		return errors.Wrap(err, "unable to initialize routers")
	}

	return nil
}

func (s *Service) terminateRouter() {
}
