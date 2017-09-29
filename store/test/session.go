package test

import (
	"github.com/onsi/gomega"
	"github.com/tidepool-org/platform/test"
)

type Session struct {
	*test.Mock
	IsClosedInvocations      int
	IsClosedOutputs          []bool
	CloseInvocations         int
	EnsureIndexesInvocations int
	EnsureIndexesOutputs     []error
}

func NewSession() *Session {
	return &Session{
		Mock: test.NewMock(),
	}
}

func (s *Session) IsClosed() bool {
	s.IsClosedInvocations++

	gomega.Expect(s.IsClosedOutputs).ToNot(gomega.BeEmpty())

	output := s.IsClosedOutputs[0]
	s.IsClosedOutputs = s.IsClosedOutputs[1:]
	return output
}

func (s *Session) Close() {
	s.CloseInvocations++
}

func (s *Session) EnsureIndexes() error {
	s.EnsureIndexesInvocations++

	gomega.Expect(s.EnsureIndexesOutputs).ToNot(gomega.BeEmpty())

	output := s.EnsureIndexesOutputs[0]
	s.EnsureIndexesOutputs = s.EnsureIndexesOutputs[1:]
	return output
}

func (s *Session) UnusedOutputsCount() int {
	return len(s.IsClosedOutputs) +
		len(s.EnsureIndexesOutputs)
}

func (s *Session) Expectations() {
	s.Mock.Expectations()
	gomega.Expect(s.IsClosedOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.EnsureIndexesOutputs).To(gomega.BeEmpty())
}
