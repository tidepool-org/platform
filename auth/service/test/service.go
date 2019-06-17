package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/store"
	authStoreTest "github.com/tidepool-org/platform/auth/store/test"
	"github.com/tidepool-org/platform/provider"
	providerTest "github.com/tidepool-org/platform/provider/test"
	serviceTest "github.com/tidepool-org/platform/service/test"
	"github.com/tidepool-org/platform/task"
	taskTest "github.com/tidepool-org/platform/task/test"
)

type Service struct {
	*serviceTest.Service
	DomainInvocations          int
	DomainOutputs              []string
	AuthStoreInvocations       int
	AuthStoreImpl              *authStoreTest.Store
	ProviderFactoryInvocations int
	ProviderFactoryImpl        *providerTest.Factory
	TaskClientInvocations      int
	TaskClientImpl             *taskTest.Client
	StatusInvocations          int
	StatusOutputs              []*service.Status
}

func NewService() *Service {
	return &Service{
		Service:             serviceTest.NewService(),
		AuthStoreImpl:       authStoreTest.NewStore(),
		ProviderFactoryImpl: providerTest.NewFactory(),
		TaskClientImpl:      taskTest.NewClient(),
	}
}

func (s *Service) Domain() string {
	s.DomainInvocations++

	gomega.Expect(s.DomainOutputs).ToNot(gomega.BeEmpty())

	output := s.DomainOutputs[0]
	s.DomainOutputs = s.DomainOutputs[1:]
	return output
}

func (s *Service) AuthStore() store.Store {
	s.AuthStoreInvocations++

	return s.AuthStoreImpl
}

func (s *Service) ProviderFactory() provider.Factory {
	s.ProviderFactoryInvocations++

	return s.ProviderFactoryImpl
}

func (s *Service) TaskClient() task.Client {
	s.TaskClientInvocations++

	return s.TaskClientImpl
}

func (s *Service) Status() *service.Status {
	s.StatusInvocations++

	gomega.Expect(s.StatusOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Service) Expectations() {
	s.Service.Expectations()
	s.AuthStoreImpl.Expectations()
	s.ProviderFactoryImpl.Expectations()
	s.TaskClientImpl.Expectations()
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
