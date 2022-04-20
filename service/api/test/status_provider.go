package test

import "context"

type StatusProvider struct {
	StatusInvocations int
	StatusStub        func() interface{}
	StatusOutputs     []interface{}
	StatusOutput      *interface{}
}

func NewStatusProvider() *StatusProvider {
	return &StatusProvider{}
}

func (s *StatusProvider) Status(ctx context.Context) interface{} {
	s.StatusInvocations++
	if s.StatusStub != nil {
		return s.StatusStub()
	}
	if len(s.StatusOutputs) > 0 {
		output := s.StatusOutputs[0]
		s.StatusOutputs = s.StatusOutputs[1:]
		return output
	}
	if s.StatusOutput != nil {
		return *s.StatusOutput
	}
	panic("Status has no output")
}

func (s *StatusProvider) AssertOutputsEmpty() {
	if len(s.StatusOutputs) > 0 {
		panic("StatusOutputs is not empty")
	}
}
