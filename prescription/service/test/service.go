package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription"

	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"

	"github.com/tidepool-org/platform/prescription/service"
	serviceTest "github.com/tidepool-org/platform/service/test"
)

type Service struct {
	*serviceTest.Service
	DomainInvocations             int
	DomainOutputs                 []string
	StatusInvocations             int
	StatusOutputs                 []*service.Status
	PrescriptionStoreInvocations  int
	PrescriptionStoreOutputs      []*mongo.Store
	PrescriptionClientInvocations int
	PrescriptionClientOutputs     []prescription.Client
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

	gomega.Expect(s.PrescriptionStoreOutputs).ToNot(gomega.BeEmpty())

	output := s.PrescriptionStoreOutputs[0]
	s.PrescriptionStoreOutputs = s.PrescriptionStoreOutputs[1:]
	return output
}

func (s *Service) PrescriptionClient() prescription.Client {
	s.PrescriptionClientInvocations++

	gomega.Expect(s.PrescriptionClientOutputs).ToNot(gomega.BeEmpty())

	output := s.PrescriptionClientOutputs[0]
	s.PrescriptionClientOutputs = s.PrescriptionClientOutputs[1:]
	return output
}

func (s *Service) Expectations() {
	s.Service.Expectations()
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
