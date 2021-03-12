package mongo

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(cfg *Config, lgr log.Logger) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	baseStore, err := storeStructuredMongo.NewStore(cfg.Config, lgr)
	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Store{
		Store:  baseStore,
		config: cfg,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
	config *Config
}

func (s *Store) NewPermissionsSession() store.PermissionsSession {
	return &PermissionsSession{
		Session: s.Store.NewSession("perms"),
		config:  s.config,
	}
}

type PermissionsSession struct {
	*storeStructuredMongo.Session
	config *Config
}
