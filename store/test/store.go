package test

import "github.com/tidepool-org/platform/test"

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

	if len(s.IsClosedOutputs) == 0 {
		panic("Unexpected invocation of IsClosed on Store")
	}

	output := s.IsClosedOutputs[0]
	s.IsClosedOutputs = s.IsClosedOutputs[1:]
	return output
}

func (s *Store) Close() {
	s.CloseInvocations++
}

func (s *Store) Status() interface{} {
	s.StatusInvocations++

	if len(s.StatusOutputs) == 0 {
		panic("Unexpected invocation of Status on Store")
	}

	output := s.StatusOutputs[0]
	s.StatusOutputs = s.StatusOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.IsClosedOutputs) +
		len(s.StatusOutputs)
}
