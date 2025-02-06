package mongo

import (
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(c *storeStructuredMongo.Config) (*Store, error) {
	if c == nil {
		return nil, errors.New("config is missing")
	}

	str, err := storeStructuredMongo.NewStore(c)
	return &Store{
		str,
	}, err
}

func (s *Store) EnsureIndexes() error {
	providerSessionRepository := s.providerSessionRepository()
	if err := providerSessionRepository.EnsureIndexes(); err != nil {
		return err
	}

	deviceTokensRepository := s.deviceTokenRepository()
	if err := deviceTokensRepository.EnsureIndexes(); err != nil {
		return err
	}

	appValidateRepository := s.restrictedAppValidateRepository()
	if err := appValidateRepository.EnsureIndexes(); err != nil {
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

func (s *Store) NewDeviceTokenRepository() store.DeviceTokenRepository {
	return s.deviceTokenRepository()
}

func (s *Store) NewAppValidateRepository() appvalidate.Repository {
	return s.restrictedAppValidateRepository()
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

func (s *Store) deviceTokenRepository() devicetokens.Repository {
	r := deviceTokenRepo(*s.Store.GetRepository("deviceTokens"))
	return &r
}

func (s *Store) restrictedAppValidateRepository() *AppValidateRepository {
	return &AppValidateRepository{
		s.Store.GetRepository("app_validations"),
	}
}
