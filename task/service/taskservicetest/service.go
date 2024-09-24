package taskservicetest

import (
	"context"

	serviceTest "github.com/tidepool-org/platform/service/test"
	"github.com/tidepool-org/platform/task"
	taskService "github.com/tidepool-org/platform/task/service"
	taskStore "github.com/tidepool-org/platform/task/store"
	"github.com/tidepool-org/platform/task/store/taskstoretest"
	"github.com/tidepool-org/platform/task/tasktest"
)

type Service struct {
	*serviceTest.Service
	TaskStoreInvocations  int
	TaskStoreImpl         *taskstoretest.Store
	TaskClientInvocations int
	TaskClientImpl        *tasktest.Client
	StatusInvocations     int
	StatusOutputs         []*taskService.Status
}

func NewService() *Service {
	return &Service{
		Service:        serviceTest.NewService(),
		TaskStoreImpl:  taskstoretest.NewStore(),
		TaskClientImpl: tasktest.NewClient(),
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
