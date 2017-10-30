package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type Source struct {
	*test.Mock
	ParameterInvocations     int
	ParameterOutput          *string
	ParameterOutputs         []string
	PointerInvocations       int
	PointerOutput            *string
	PointerOutputs           []string
	WithReferenceInvocations int
	WithReferenceInputs      []string
	WithReferenceOutput      *structure.Source
	WithReferenceOutputs     []structure.Source
}

func NewSource() *Source {
	return &Source{
		Mock: test.NewMock(),
	}
}

func (s *Source) Parameter() string {
	s.ParameterInvocations++

	if s.ParameterOutput != nil {
		return *s.ParameterOutput
	}

	gomega.Expect(s.ParameterOutputs).ToNot(gomega.BeEmpty())

	output := s.ParameterOutputs[0]
	s.ParameterOutputs = s.ParameterOutputs[1:]
	return output
}

func (s *Source) Pointer() string {
	s.PointerInvocations++

	if s.PointerOutput != nil {
		return *s.PointerOutput
	}

	gomega.Expect(s.PointerOutputs).ToNot(gomega.BeEmpty())

	output := s.PointerOutputs[0]
	s.PointerOutputs = s.PointerOutputs[1:]
	return output
}

func (s *Source) WithReference(reference string) structure.Source {
	s.WithReferenceInvocations++

	if s.WithReferenceOutput != nil {
		return *s.WithReferenceOutput
	}

	s.WithReferenceInputs = append(s.WithReferenceInputs, reference)

	gomega.Expect(s.WithReferenceOutputs).ToNot(gomega.BeEmpty())

	output := s.WithReferenceOutputs[0]
	s.WithReferenceOutputs = s.WithReferenceOutputs[1:]
	return output
}

func (s *Source) Expectations() {
	s.Mock.Expectations()
	gomega.Expect(s.ParameterOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.PointerOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.WithReferenceOutputs).To(gomega.BeEmpty())
}
