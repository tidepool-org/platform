package test

import (
	storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"
	"github.com/tidepool-org/platform/task/store"
)

type Store struct {
	*storeStructuredTest.Store
	NewTaskSessionInvocations int
	NewTaskSessionOutputs     []store.TaskSession
}

func NewStore() *Store {
	return &Store{
		Store: storeStructuredTest.NewStore(),
	}
}

func (s *Store) NewTaskSession() store.TaskSession {
	s.NewTaskSessionInvocations++

	if len(s.NewTaskSessionOutputs) == 0 {
		panic("Unexpected invocation of NewTaskSession on Store")
	}

	output := s.NewTaskSessionOutputs[0]
	s.NewTaskSessionOutputs = s.NewTaskSessionOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewTaskSessionOutputs)
}
