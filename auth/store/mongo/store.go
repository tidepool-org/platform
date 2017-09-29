package mongo

import (
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
)

type Store struct {
	*mongo.Store
}

func NewStore(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	str, err := mongo.New(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	providerSessionSession := s.NewProviderSessionSession()
	defer providerSessionSession.Close()
	if err := providerSessionSession.EnsureIndexes(); err != nil {
		return err
	}

	restrictedTokenSession := s.NewRestrictedTokenSession()
	defer restrictedTokenSession.Close()
	return restrictedTokenSession.EnsureIndexes()
}

func (s *Store) NewProviderSessionSession() store.ProviderSessionSession {
	return &ProviderSessionSession{
		Session: s.Store.NewSession("provider_sessions"),
	}
}

func (s *Store) NewRestrictedTokenSession() store.RestrictedTokenSession {
	return &RestrictedTokenSession{
		Session: s.Store.NewSession("restricted_tokens"),
	}
}
