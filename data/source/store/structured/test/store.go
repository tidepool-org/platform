package test

import dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"

type Store struct {
	NewSessionInvocations int
	NewSessionStub        func() dataSourceStoreStructured.DataRepository
	NewSessionOutputs     []dataSourceStoreStructured.DataRepository
	NewSessionOutput      *dataSourceStoreStructured.DataRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewDataRepository() dataSourceStoreStructured.DataRepository {
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
	panic("NewDataRepository has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewSessionOutputs) > 0 {
		panic("NewSessionOutputs is not empty")
	}
}
