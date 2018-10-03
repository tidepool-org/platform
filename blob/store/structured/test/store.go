package test

import blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"

type Store struct {
	NewSessionInvocations int
	NewSessionStub        func() blobStoreStructured.Session
	NewSessionOutputs     []blobStoreStructured.Session
	NewSessionOutput      *blobStoreStructured.Session
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewSession() blobStoreStructured.Session {
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
