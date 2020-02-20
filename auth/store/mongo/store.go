package mongo

import (
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(cfg *storeStructuredMongo.Config, lgr log.Logger) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	providerSessionSession := s.providerSessionSession()
	defer providerSessionSession.Close()
	if err := providerSessionSession.EnsureIndexes(); err != nil {
		return err
	}

	restrictedTokenSession := s.restrictedTokenSession()
	defer restrictedTokenSession.Close()
	if err := restrictedTokenSession.EnsureIndexes(); err != nil {
		return err
	}

	deviceAuthorizationSession := s.deviceAuthorizationSession()
	defer deviceAuthorizationSession.Close()
	return deviceAuthorizationSession.EnsureIndexes()
}

func (s *Store) NewProviderSessionSession() store.ProviderSessionSession {
	return s.providerSessionSession()
}

func (s *Store) NewRestrictedTokenSession() store.RestrictedTokenSession {
	return s.restrictedTokenSession()
}

func (s *Store) NewDeviceAuthorizationSession() store.DeviceAuthorizationSession {
	return s.deviceAuthorizationSession()
}

func (s *Store) providerSessionSession() *ProviderSessionSession {
	return &ProviderSessionSession{
		Session: s.Store.NewSession("provider_sessions"),
	}
}

func (s *Store) restrictedTokenSession() *RestrictedTokenSession {
	return &RestrictedTokenSession{
		Session: s.Store.NewSession("restricted_tokens"),
	}
}

func (s *Store) deviceAuthorizationSession() *DeviceAuthorizationSession {
	return &DeviceAuthorizationSession{
		Session: s.Store.NewSession("dthorizations"),
	}
}
