package test

import (
	"context"

	serviceTest "github.com/tidepool-org/platform/service/test"
	"github.com/tidepool-org/platform/task"
	taskService "github.com/tidepool-org/platform/task/service"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreTest "github.com/tidepool-org/platform/task/store/test"
	taskTest "github.com/tidepool-org/platform/task/test"
)

type Service struct {
	*serviceTest.Service
	TaskStoreInvocations  int
	TaskStoreImpl         *taskStoreTest.Store
	TaskClientInvocations int
	TaskClientImpl        *taskTest.Client
	StatusInvocations     int
	StatusOutputs         []*taskService.Status
}

func NewService() *Service {
	return &Service{
		Service:        serviceTest.NewService(),
		TaskStoreImpl:  taskStoreTest.NewStore(),
		TaskClientImpl: taskTest.NewClient(),
	}
}

func (s *Service) TaskStore() taskStore.Store {
	s.TaskStoreInvocations++

	return s.TaskStoreImpl
}

func (s *Service) TaskClient() task.Client {
	s.TaskClientInvocations++

	return s.TaskClientImpl
}

func (s *Service) Status(ctx context.Context) *taskService.Status {
	s.StatusInvocations++

	if len(s.StatusOutputs) == 0 {
		panic("Unexpected invocation of Status on Service")
	}

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Service) UnusedOutputsCount() int {
	return s.TaskStoreImpl.UnusedOutputsCount() +
		len(s.StatusOutputs)
}
