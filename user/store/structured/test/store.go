package test

import userStoreStructured "github.com/tidepool-org/platform/user/store/structured"

type Store struct {
	NewSessionInvocations int
	NewSessionStub        func() userStoreStructured.Session
	NewSessionOutputs     []userStoreStructured.Session
	NewSessionOutput      *userStoreStructured.Session
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewSession() userStoreStructured.Session {
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
