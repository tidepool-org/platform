package test

import (
	"github.com/tidepool-org/platform/notification/store"
)

type Store struct {
	NewNotificationsRepositoryInvocations int
	NewNotificationsRepositoryOutputs     []store.NotificationsRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewNotificationsRepository() store.NotificationsRepository {
	s.NewNotificationsRepositoryInvocations++

	if len(s.NewNotificationsRepositoryOutputs) == 0 {
		panic("Unexpected invocation of NewNotificationsRepository on Store")
	}

	output := s.NewNotificationsRepositoryOutputs[0]
	s.NewNotificationsRepositoryOutputs = s.NewNotificationsRepositoryOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewNotificationsRepositoryOutputs)
}
