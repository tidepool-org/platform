package test

import "github.com/tidepool-org/platform/synctask/store"

type Store struct {
	NewSyncTaskSessionInvocations int
	NewSyncTaskSessionImpl        *SyncTaskSession
}

func NewStore() *Store {
	return &Store{
		NewSyncTaskSessionImpl: NewSyncTaskSession(),
	}
}

func (s *Store) NewSyncTaskSession() store.SyncTaskSession {
	s.NewSyncTaskSessionInvocations++
	return s.NewSyncTaskSessionImpl
}

func (s *Store) Expectations() {
	s.NewSyncTaskSessionImpl.Expectations()
}
