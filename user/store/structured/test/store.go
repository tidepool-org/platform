package test

import userStoreStructured "github.com/tidepool-org/platform/user/store/structured"

type Store struct {
	NewRepositoryInvocations int
	NewRepositoryStub        func() userStoreStructured.UserRepository
	NewRepositoryOutputs     []userStoreStructured.UserRepository
	NewRepositoryOutput      *userStoreStructured.UserRepository
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) NewUserRepository() userStoreStructured.UserRepository {
	s.NewRepositoryInvocations++
	if s.NewRepositoryStub != nil {
		return s.NewRepositoryStub()
	}
	if len(s.NewRepositoryOutputs) > 0 {
		output := s.NewRepositoryOutputs[0]
		s.NewRepositoryOutputs = s.NewRepositoryOutputs[1:]
		return output
	}
	if s.NewRepositoryOutput != nil {
		return *s.NewRepositoryOutput
	}
	panic("NewRepository has no output")
}

func (s *Store) AssertOutputsEmpty() {
	if len(s.NewRepositoryOutputs) > 0 {
		panic("NewRepositoryOutputs is not empty")
	}
}
