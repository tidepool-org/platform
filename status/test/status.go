package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription/status"
)

type Reporter struct {
	StatusInvocations int
	StatusOutputs     []*status.Status
}

func NewReporter() *Reporter {
	return &Reporter{}
}

func (s *Reporter) Status() *status.Status {
	s.StatusInvocations++

	gomega.Expect(s.StatusOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Reporter) Expectations() {
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
