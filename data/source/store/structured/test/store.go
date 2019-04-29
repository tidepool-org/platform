package test

import dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"

type Store struct {
	NewSessionInvocations int
	NewSessionStub        func() dataSourceStoreStructured.Session
	NewSessionOutputs     []dataSourceStoreStructured.Session
	NewSessionOutput      *dataSourceStoreStructured.Session
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewSession() dataSourceStoreStructured.Session {
	s.NewSessionInvocations++
	if s.NewSessionStub != nil {
		return s.NewSessionStub()
	}
	if len(s.NewSessionOutputs) > 0 {
		output := s.NewSessionOutputs[0]
		s.NewSessionOutputs = s.NewSessionOutputs[1:]
		return output
	}
	if s.NewSessionOutput != nil {
		return *s.NewSessionOutput
	}
	panic("NewSession has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewSessionOutputs) > 0 {
		panic("NewSessionOutputs is not empty")
	}
}
