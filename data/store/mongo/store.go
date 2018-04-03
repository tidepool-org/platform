package mongo

import (
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
)

type Store struct {
	*mongo.Store
}

func NewStore(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	str, err := mongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	dataSourceSession := s.NewDataSourceSession()
	defer dataSourceSession.Close()
	return dataSourceSession.EnsureIndexes()
}

func (s *Store) NewDataSourceSession() store.DataSourceSession {
	return &DataSourceSession{
		Session: s.Store.NewSession("data_sources"),
	}
}
