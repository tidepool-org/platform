package test

import "github.com/tidepool-org/platform/synctask/store"

type Store struct {
	NewSyncTaskRepositoryInvocations int
	NewSyncTaskRepositoryImpl        *SyncTaskRepository
}

func NewStore() *Store {
	return &Store{
		NewSyncTaskRepositoryImpl: NewSyncTaskRepository(),
	}
}

func (s *Store) NewSyncTaskRepository() store.SyncTaskRepository {
	s.NewSyncTaskRepositoryInvocations++
	return s.NewSyncTaskRepositoryImpl
}

func (s *Store) Expectations() {
	s.NewSyncTaskRepositoryImpl.Expectations()
}
