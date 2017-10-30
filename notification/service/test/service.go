package test

import (
	"github.com/tidepool-org/platform/notification/service"
	"github.com/tidepool-org/platform/notification/store"
	testStore "github.com/tidepool-org/platform/notification/store/test"
	testService "github.com/tidepool-org/platform/service/test"
)

type Service struct {
	*testService.Service
	NotificationStoreInvocations int
	NotificationStoreImpl        *testStore.Store
	StatusInvocations            int
	StatusOutputs                []*service.Status
}

func NewService() *Service {
	return &Service{
		Service:               testService.NewService(),
		NotificationStoreImpl: testStore.NewStore(),
	}
}

func (s *Service) NotificationStore() store.Store {
	s.NotificationStoreInvocations++

	return s.NotificationStoreImpl
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
	return s.NotificationStoreImpl.UnusedOutputsCount() +
		len(s.StatusOutputs)
}
