package test

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notification/store"
	testStore "github.com/tidepool-org/platform/store/test"
)

type Store struct {
	*testStore.Store
	NewSessionInvocations int
	NewSessionInputs      []log.Logger
	NewSessionOutputs     []store.StoreSession
}

func NewStore() *Store {
	return &Store{
		Store: testStore.NewStore(),
	}
}

func (s *Store) NewSession(lgr log.Logger) store.StoreSession {
	s.NewSessionInvocations++

	s.NewSessionInputs = append(s.NewSessionInputs, lgr)

	if len(s.NewSessionOutputs) == 0 {
		panic("Unexpected invocation of NewSession on Store")
	}

	output := s.NewSessionOutputs[0]
	s.NewSessionOutputs = s.NewSessionOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewSessionOutputs)
}
