package mongo

import (
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store: store,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) EnsureIndexes() error {
	return nil
}
