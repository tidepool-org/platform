package test

import (
	"github.com/tidepool-org/platform/auth/store"
)

type Store struct {
	NewProviderSessionSessionInvocations     int
	NewProviderSessionSessionImpl            *ProviderSessionSession
	NewRestrictedTokenSessionInvocations     int
	NewRestrictedTokenSessionImpl            *RestrictedTokenSession
	NewDeviceAuthorizationSessionInvocations int
	NewDeviceAuthorizationSessionImpl        *DeviceAuthorizationSession
}

func NewStore() *Store {
	return &Store{
		NewProviderSessionSessionImpl:     NewProviderSessionSession(),
		NewRestrictedTokenSessionImpl:     NewRestrictedTokenSession(),
		NewDeviceAuthorizationSessionImpl: NewDeviceAuthorizationSession(),
	}
}

func (s *Store) NewProviderSessionSession() store.ProviderSessionSession {
	s.NewProviderSessionSessionInvocations++
	return s.NewProviderSessionSessionImpl
}

func (s *Store) NewRestrictedTokenSession() store.RestrictedTokenSession {
	s.NewRestrictedTokenSessionInvocations++
	return s.NewRestrictedTokenSessionImpl
}

func (s *Store) NewDeviceAuthorizationSession() store.DeviceAuthorizationSession {
	s.NewDeviceAuthorizationSessionInvocations++
	return s.NewDeviceAuthorizationSessionImpl
}

func (s *Store) Expectations() {
	s.NewProviderSessionSessionImpl.Expectations()
	s.NewRestrictedTokenSessionImpl.Expectations()
	s.NewDeviceAuthorizationSessionImpl.Expectations()
}
