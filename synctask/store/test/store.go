package test

import (
	testStore "github.com/tidepool-org/platform/store/test"
	"github.com/tidepool-org/platform/synctask/store"
)

type Store struct {
	*testStore.Store
	NewSyncTaskSessionInvocations int
	NewSyncTaskSessionImpl        *SyncTaskSession
}

func NewStore() *Store {
	return &Store{
		Store: testStore.NewStore(),
		NewSyncTaskSessionImpl: NewSyncTaskSession(),
	}
}

func (s *Store) NewSyncTaskSession() store.SyncTaskSession {
	s.NewSyncTaskSessionInvocations++
	return s.NewSyncTaskSessionImpl
}

func (s *Store) Expectations() {
	s.Store.Expectations()
	s.NewSyncTaskSessionImpl.Expectations()
}
