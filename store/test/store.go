package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type Store struct {
	*test.Mock
	IsClosedInvocations int
	IsClosedOutputs     []bool
	CloseInvocations    int
	StatusInvocations   int
	StatusOutputs       []interface{}
}

func NewStore() *Store {
	return &Store{
		Mock: test.NewMock(),
	}
}

func (s *Store) IsClosed() bool {
	s.IsClosedInvocations++

	gomega.Expect(s.IsClosedOutputs).ToNot(gomega.BeEmpty())

	output := s.IsClosedOutputs[0]
	s.IsClosedOutputs = s.IsClosedOutputs[1:]
	return output
}

func (s *Store) Close() {
	s.CloseInvocations++
}

func (s *Store) Status() interface{} {
	s.StatusInvocations++

	gomega.Expect(s.StatusOutputs).ToNot(gomega.BeEmpty())

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.IsClosedOutputs) +
		len(s.StatusOutputs)
}

func (s *Store) Expectations() {
	s.Mock.Expectations()
	gomega.Expect(s.IsClosedOutputs).To(gomega.BeEmpty())
	gomega.Expect(s.StatusOutputs).To(gomega.BeEmpty())
}
