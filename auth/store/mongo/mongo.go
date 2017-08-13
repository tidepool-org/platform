package mongo

import (
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
)

type Store struct {
	*mongo.Store
}

func New(lgr log.Logger, cfg *mongo.Config) (*Store, error) {
	str, err := mongo.New(lgr, cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) NewAuthsSession(lgr log.Logger) store.AuthsSession {
	return &AuthsSession{
		Session: s.Store.NewSession(lgr, "auths"),
	}
}

type AuthsSession struct {
	*mongo.Session
}
