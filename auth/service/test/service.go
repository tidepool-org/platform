package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/store"
	testStore "github.com/tidepool-org/platform/auth/store/test"
	testService "github.com/tidepool-org/platform/service/test"
)

type Service struct {
	*testService.Service
	AuthStoreInvocations int
	AuthStoreImpl        *testStore.Store
	StatusInvocations    int
	StatusOutputs        []*auth.Status
}

func NewService() *Service {
	return &Service{
		Service:       testService.NewService(),
		AuthStoreImpl: testStore.NewStore(),
	}
}

func (s *Service) AuthStore() store.Store {
	s.AuthStoreInvocations++

	return s.AuthStoreImpl
}

func (s *Service) Status() *auth.Status {
	s.StatusInvocations++

	if len(s.StatusOutputs) == 0 {
		panic("Unexpected invocation of Status on Service")
	}

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Service) UnusedOutputsCount() int {
	return s.AuthStoreImpl.UnusedOutputsCount() +
		len(s.StatusOutputs)
}
