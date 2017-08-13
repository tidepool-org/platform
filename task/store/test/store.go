package test

import (
	"github.com/tidepool-org/platform/log"
	testStore "github.com/tidepool-org/platform/store/test"
	"github.com/tidepool-org/platform/task/store"
)

type Store struct {
	*testStore.Store
	NewTasksSessionInvocations int
	NewTasksSessionInputs      []log.Logger
	NewTasksSessionOutputs     []store.TasksSession
}

func NewStore() *Store {
	return &Store{
		Store: testStore.NewStore(),
	}
}

func (s *Store) NewTasksSession(lgr log.Logger) store.TasksSession {
	s.NewTasksSessionInvocations++

	s.NewTasksSessionInputs = append(s.NewTasksSessionInputs, lgr)

	if len(s.NewTasksSessionOutputs) == 0 {
		panic("Unexpected invocation of NewTasksSession on Store")
	}

	output := s.NewTasksSessionOutputs[0]
	s.NewTasksSessionOutputs = s.NewTasksSessionOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewTasksSessionOutputs)
}
