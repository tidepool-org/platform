package test

import (
	"github.com/tidepool-org/platform/auth/store"
)

type Store struct {
	NewProviderSessionRepositoryInvocations int
	NewProviderSessionRepositoryImpl        *ProviderSessionRepository
	NewRestrictedTokenRepositoryInvocations int
	NewRestrictedTokenRepositoryImpl        *RestrictedTokenRepository
}

func NewStore() *Store {
	return &Store{
		NewProviderSessionRepositoryImpl: NewProviderSessionRepository(),
		NewRestrictedTokenRepositoryImpl: NewRestrictedTokenRepository(),
	}
}

func (s *Store) NewProviderSessionRepository() store.ProviderSessionRepository {
	s.NewProviderSessionRepositoryInvocations++
	return s.NewProviderSessionRepositoryImpl
}

func (s *Store) NewRestrictedTokenRepository() store.RestrictedTokenRepository {
	s.NewRestrictedTokenRepositoryInvocations++
	return s.NewRestrictedTokenRepositoryImpl
}

func (s *Store) Expectations() {
	s.NewProviderSessionRepositoryImpl.Expectations()
	s.NewRestrictedTokenRepositoryImpl.Expectations()
}
