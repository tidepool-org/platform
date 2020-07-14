package mongo

import (
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(p storeStructuredMongo.Params) (*Store, error) {
	if p.DatabaseConfig == nil {
		return nil, errors.New("config is missing")
	}

	str, err := storeStructuredMongo.NewStore(p)
	return &Store{
		str,
	}, err
}

func (s *Store) EnsureIndexes() error {
	providerSessionRepository := s.providerSessionRepository()
	if err := providerSessionRepository.EnsureIndexes(); err != nil {
		return err
	}

	restrictedTokenRepository := s.restrictedTokenRepository()
	return restrictedTokenRepository.EnsureIndexes()
}

func (s *Store) NewProviderSessionRepository() store.ProviderSessionRepository {
	return s.providerSessionRepository()
}

func (s *Store) NewRestrictedTokenRepository() store.RestrictedTokenRepository {
	return s.restrictedTokenRepository()
}

func (s *Store) providerSessionRepository() *ProviderSessionRepository {
	return &ProviderSessionRepository{
		s.Store.GetRepository("provider_sessions"),
	}
}

func (s *Store) restrictedTokenRepository() *RestrictedTokenRepository {
	return &RestrictedTokenRepository{
		s.Store.GetRepository("restricted_tokens"),
	}
}
