package taskstoretest

import (
	"context"

	"github.com/tidepool-org/platform/task/store"
)

type Store struct {
	NewTaskRepositoryInvocations int
	NewTaskRepositoryOutputs     []store.TaskRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewTaskRepository() store.TaskRepository {
	s.NewTaskRepositoryInvocations++

	if len(s.NewTaskRepositoryOutputs) == 0 {
		panic("Unexpected invocation of NewTaskRepository on Store")
	}

	output := s.NewTaskRepositoryOutputs[0]
	s.NewTaskRepositoryOutputs = s.NewTaskRepositoryOutputs[1:]
	return output
}

func (s *Store) WithTypeFilter(typeFilter string) store.Store {
	return s
}

func (s *Store) Terminate(ctx context.Context) error {
	return nil
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewTaskRepositoryOutputs)
}
