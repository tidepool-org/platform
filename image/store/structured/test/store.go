package test

import imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"

type Store struct {
	NewSessionInvocations int
	NewSessionStub        func() imageStoreStructured.Session
	NewSessionOutputs     []imageStoreStructured.Session
	NewSessionOutput      *imageStoreStructured.Session
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewSession() imageStoreStructured.Session {
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
