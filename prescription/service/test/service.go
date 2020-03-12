package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"

	"github.com/tidepool-org/platform/prescription/service"
	serviceTest "github.com/tidepool-org/platform/service/test"
)

type Service struct {
	*serviceTest.Service
	DomainInvocations            int
	DomainOutputs                []string
	StatusInvocations            int
	StatusOutputs                []*service.Status
	PrescriptionStoreInvocations int
	PrescriptionOutputs          []*mongo.Store
}

func NewService() *Service {
	return &Service{
		Service: serviceTest.NewService(),
	}
}

func (s *Service) Domain() string {
	s.DomainInvocations++

	gomega.Expect(s.DomainOutputs).ToNot(gomega.BeEmpty())

	output := s.DomainOutputs[0]
	s.DomainOutputs = s.DomainOutputs[1:]
	return output
}

func (s *Service) Status() *service.Status {
	s.StatusInvocations++

	gomega.Expect(s.StatusOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Service) PrescriptionStore() store.Store {
	s.PrescriptionStoreInvocations++

	gomega.Expect(s.PrescriptionOutputs).ToNot(gomega.BeEmpty())

	output := s.PrescriptionOutputs[0]
	s.PrescriptionOutputs = s.PrescriptionOutputs[1:]
	return output
}

func (s *Service) Expectations() {
	s.Service.Expectations()
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
