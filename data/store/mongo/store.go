package mongo

import (
	"github.com/tidepool-org/platform/data/store"
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
	dataSourceSession := s.dataSourceSession()
	defer dataSourceSession.Close()
	return dataSourceSession.EnsureIndexes()
}

func (s *Store) NewDataSourceSession() store.DataSourceSession {
	return s.dataSourceSession()
}

func (s *Store) dataSourceSession() *DataSourceSession {
	return &DataSourceSession{
		Session: s.Store.NewSession("data_sources"),
	}
}
