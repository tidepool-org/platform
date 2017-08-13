package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	serviceService "github.com/tidepool-org/platform/service/service"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/service/api"
	"github.com/tidepool-org/platform/task/service/api/v1"
	"github.com/tidepool-org/platform/task/store"
	taskMongo "github.com/tidepool-org/platform/task/store/mongo"
)

type Service struct {
	*serviceService.Service
	taskStore *taskMongo.Store
}

func New(prefix string) (*Service, error) {
	svc, err := serviceService.New(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Service: svc,
	}, nil
}

func (s *Service) Initialize() error {
	if err := s.Service.Initialize(); err != nil {
		return err
	}

	if err := s.initializeRouter(); err != nil {
		return err
	}
	if err := s.initializeTaskStore(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Terminate() {
	if s.taskStore != nil {
		s.taskStore.Close()
		s.taskStore = nil
	}

	s.Service.Terminate()
}

func (s *Service) TaskStore() store.Store {
	return s.taskStore
}

func (s *Service) Status() *task.Status {
	return &task.Status{
		Version:   s.VersionReporter().Long(),
		TaskStore: s.TaskStore().GetStatus(),
		Server:    s.API().Status(),
	}
}

func (s *Service) initializeRouter() error {
	routes := []*rest.Route{}

	s.Logger().Debug("Creating api router")

	apiRouter, err := api.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create api router")
	}
	routes = append(routes, apiRouter.Routes()...)

	s.Logger().Debug("Creating v1 router")

	v1Router, err := v1.NewRouter(s)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create v1 router")
	}
	routes = append(routes, v1Router.Routes()...)

	s.Logger().Debug("Initializing router")

	if err = s.API().InitializeRouter(routes...); err != nil {
		return errors.Wrap(err, "service", "unable to initialize router")
	}

	return nil
}

func (s *Service) initializeTaskStore() error {
	s.Logger().Debug("Loading task store config")

	cfg := baseMongo.NewConfig()
	if err := cfg.Load(s.ConfigReporter().WithScopes("task", "store")); err != nil {
		return errors.Wrap(err, "service", "unable to load task store config")
	}
	cfg.Collection = "tasks"

	s.Logger().Debug("Creating task store")

	str, err := taskMongo.New(s.Logger(), cfg)
	if err != nil {
		return errors.Wrap(err, "service", "unable to create task store")
	}
	s.taskStore = str

	return nil
}
