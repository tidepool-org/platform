package test

import (
	testService "github.com/tidepool-org/platform/service/test"
	"github.com/tidepool-org/platform/task/service"
	"github.com/tidepool-org/platform/task/store"
	testStore "github.com/tidepool-org/platform/task/store/test"
)

type Service struct {
	*testService.Service
	TaskStoreInvocations int
	TaskStoreImpl        *testStore.Store
	StatusInvocations    int
	StatusOutputs        []*service.Status
}

func NewService() *Service {
	return &Service{
		Service:       testService.NewService(),
		TaskStoreImpl: testStore.NewStore(),
	}
}

func (s *Service) TaskStore() store.Store {
	s.TaskStoreInvocations++

	return s.TaskStoreImpl
}

func (s *Service) Status() *service.Status {
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
