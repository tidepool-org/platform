package mongo

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task/store"
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

func (s *Store) NewSession(lgr log.Logger) store.StoreSession {
	return &Session{
		Session: s.Store.NewSession(lgr),
	}
}

type Session struct {
	*mongo.Session
}
