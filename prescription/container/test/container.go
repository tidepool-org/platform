package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription/status"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/prescription"

	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"
)

type Container struct {
	StatusReporterInvocations      int
	StatusReporterOutputs          []status.Reporter
	PrescriptionStoreInvocations   int
	PrescriptionStoreOutputs       []*mongo.Store
	PrescriptionServiceInvocations int
	PrescriptionServiceOutputs     []prescription.Service
	UserClientInvocations          int
	UserClientOutputs              []user.Client
	InitializeInvocations          int
	InitializeOutputs              []error
}

func NewContainer() *Container {
	return &Container{}
}

func (s *Container) Initialize() error {
	s.InitializeInvocations++

	gomega.Expect(s.InitializeOutputs).ToNot(gomega.BeEmpty())

	output := s.InitializeOutputs[0]
	s.InitializeOutputs = s.InitializeOutputs[1:]
	return output
}

func (s *Container) StatusReporter() status.Reporter {
	s.StatusReporterInvocations++

	gomega.Expect(s.StatusReporterOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusReporterOutputs[0]
	s.StatusReporterOutputs = s.StatusReporterOutputs[1:]
	return output
}

func (s *Container) PrescriptionStore() store.Store {
	s.PrescriptionStoreInvocations++

	gomega.Expect(s.PrescriptionStoreOutputs).ToNot(gomega.BeEmpty())

	output := s.PrescriptionStoreOutputs[0]
	s.PrescriptionStoreOutputs = s.PrescriptionStoreOutputs[1:]
	return output
}

func (s *Container) PrescriptionService() prescription.Service {
	s.PrescriptionServiceInvocations++

	gomega.Expect(s.PrescriptionServiceInvocations).ToNot(gomega.BeEmpty())

	output := s.PrescriptionServiceOutputs[0]
	s.PrescriptionServiceOutputs = s.PrescriptionServiceOutputs[1:]
	return output
}

func (s *Container) UserClient() user.Client {
	s.UserClientInvocations++

	gomega.Expect(s.UserClientOutputs).ToNot(gomega.BeEmpty())

	output := s.UserClientOutputs[0]
	s.UserClientOutputs = s.UserClientOutputs[1:]
	return output
}

func (s *Container) Expectations() {
	gomega.Expect(s.PrescriptionStoreOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.PrescriptionServiceOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.StatusReporterOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.UserClientOutputs).To(gomega.BeEmpty())
}
