package test

import (
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/log"
	testStore "github.com/tidepool-org/platform/store/test"
)

type Store struct {
	*testStore.Store
	NewAuthsSessionInvocations int
	NewAuthsSessionInputs      []log.Logger
	NewAuthsSessionOutputs     []store.AuthsSession
}

func NewStore() *Store {
	return &Store{
		Store: testStore.NewStore(),
	}
}

func (s *Store) NewAuthsSession(lgr log.Logger) store.AuthsSession {
	s.NewAuthsSessionInvocations++

	s.NewAuthsSessionInputs = append(s.NewAuthsSessionInputs, lgr)

	if len(s.NewAuthsSessionOutputs) == 0 {
		panic("Unexpected invocation of NewAuthsSession on Store")
	}

	output := s.NewAuthsSessionOutputs[0]
	s.NewAuthsSessionOutputs = s.NewAuthsSessionOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewAuthsSessionOutputs)
}
