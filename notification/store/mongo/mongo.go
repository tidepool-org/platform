package mongo

import (
	"github.com/tidepool-org/platform/notification/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(params storeStructuredMongo.Params) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) NewNotificationsRepository() store.NotificationsRepository {
	return &NotificationsRepository{
		s.Store.GetRepository("notifications"),
	}
}

type NotificationsRepository struct {
	*storeStructuredMongo.Repository
}
