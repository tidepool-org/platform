package test

import (
	"context"

	"github.com/onsi/gomega"
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth/service"
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/task"

	authStoreTest "github.com/tidepool-org/platform/auth/store/test"
	providerTest "github.com/tidepool-org/platform/provider/test"
	serviceTest "github.com/tidepool-org/platform/service/test"
	taskTest "github.com/tidepool-org/platform/task/test"
)

type Service struct {
	*serviceTest.Service
	DomainInvocations               int
	DomainOutputs                   []string
	AuthStoreInvocations            int
	AuthStoreImpl                   *authStoreTest.Store
	ProviderFactoryInvocations      int
	ProviderFactoryImpl             *providerTest.Factory
	TaskClientInvocations           int
	TaskClientImpl                  *taskTest.Client
	ConfirmationClientInvocations   int
	ConfirmationClientImpl          confirmationClient.ClientWithResponsesInterface
	DeviceCheckInvocations          int
	DeviceCheckImpl                 apple.DeviceCheck
	StatusInvocations               int
	StatusOutputs                   []*service.Status
	AppvalidateValidatorInvocations int
	AppvalidateValidatorImpl        *appvalidate.Validator
	PartnerSecretsInvocations       int
	PartnerSecretsImpl              *appvalidate.PartnerSecrets
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

func (s *Service) ConfirmationClient() confirmationClient.ClientWithResponsesInterface {
	s.ConfirmationClientInvocations++

	return s.ConfirmationClientImpl
}

func (s *Service) DeviceCheck() apple.DeviceCheck {
	s.DeviceCheckInvocations++

	return s.DeviceCheckImpl
}

func (s *Service) Status(ctx context.Context) *service.Status {
	s.StatusInvocations++

	gomega.Expect(s.StatusOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Service) AppValidator() *appvalidate.Validator {
	s.AppvalidateValidatorInvocations++

	return s.AppvalidateValidatorImpl
}

func (s *Service) PartnerSecrets() *appvalidate.PartnerSecrets {
	s.PartnerSecretsInvocations++

	return s.PartnerSecretsImpl
}

func (s *Service) Expectations() {
	s.Service.Expectations()
	s.AuthStoreImpl.Expectations()
	s.ProviderFactoryImpl.Expectations()
	s.TaskClientImpl.Expectations()
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
