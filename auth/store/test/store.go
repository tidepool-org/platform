package test

import (
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth/store"
)

type Store struct {
	NewProviderSessionRepositoryInvocations int
	NewProviderSessionRepositoryImpl        *ProviderSessionRepository
	NewRestrictedTokenRepositoryInvocations int
	NewRestrictedTokenRepositoryImpl        *RestrictedTokenRepository
	NewDeviceTokenRepositoryInvocations     int
	NewDeviceTokenRepositoryImpl            *DeviceTokenRepository
}

func NewStore() *Store {
	return &Store{
		NewProviderSessionRepositoryImpl: NewProviderSessionRepository(),
		NewRestrictedTokenRepositoryImpl: NewRestrictedTokenRepository(),
		NewDeviceTokenRepositoryImpl:     NewDeviceTokenRepository(),
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

func (s *Store) NewDeviceTokenRepository() store.DeviceTokenRepository {
	s.NewRestrictedTokenRepositoryInvocations++
	return s.NewDeviceTokenRepositoryImpl
}

func (s *Store) Expectations() {
	s.NewProviderSessionRepositoryImpl.Expectations()
	s.NewRestrictedTokenRepositoryImpl.Expectations()
	s.NewDeviceTokenRepositoryImpl.Expectations()
}

func (s *Store) NewAppValidateRepository() appvalidate.Repository {
	return nil
}
