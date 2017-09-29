package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type Source struct {
	*test.Mock
	SourceInvocations        int
	SourceOutputs            []*errors.Source
	WithReferenceInvocations int
	WithReferenceInputs      []string
	WithReferenceOutputs     []structure.Source
}

func NewSource() *Source {
	return &Source{
		Mock: test.NewMock(),
	}
}

func (s *Source) Source() *errors.Source {
	s.SourceInvocations++

	gomega.Expect(s.SourceOutputs).ToNot(gomega.BeEmpty())

	output := s.SourceOutputs[0]
	s.SourceOutputs = s.SourceOutputs[1:]
	return output
}

func (s *Source) WithReference(reference string) structure.Source {
	s.WithReferenceInvocations++

	s.WithReferenceInputs = append(s.WithReferenceInputs, reference)

	gomega.Expect(s.WithReferenceOutputs).ToNot(gomega.BeEmpty())

	output := s.WithReferenceOutputs[0]
	s.WithReferenceOutputs = s.WithReferenceOutputs[1:]
	return output
}

func (s *Source) Expectations() {
	s.Mock.Expectations()
	gomega.Expect(s.SourceOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.WithReferenceOutputs).To(gomega.BeEmpty())
}
