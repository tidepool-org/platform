package test

import (
	"github.com/tidepool-org/platform/notification/store"
	testStore "github.com/tidepool-org/platform/store/test"
)

type Store struct {
	*testStore.Store
	NewNotificationsSessionInvocations int
	NewNotificationsSessionOutputs     []store.NotificationsSession
}

func NewStore() *Store {
	return &Store{
		Store: testStore.NewStore(),
	}
}

func (s *Store) NewNotificationsSession() store.NotificationsSession {
	s.NewNotificationsSessionInvocations++

	if len(s.NewNotificationsSessionOutputs) == 0 {
		panic("Unexpected invocation of NewNotificationsSession on Store")
	}

	output := s.NewNotificationsSessionOutputs[0]
	s.NewNotificationsSessionOutputs = s.NewNotificationsSessionOutputs[1:]
	return output
}

func (s *Store) UnusedOutputsCount() int {
	return len(s.NewNotificationsSessionOutputs)
}
