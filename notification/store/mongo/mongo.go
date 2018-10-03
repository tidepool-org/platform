package mongo

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notification/store"
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

func (s *Store) NewNotificationsSession() store.NotificationsSession {
	return &NotificationsSession{
		Session: s.Store.NewSession("notifications"),
	}
}

type NotificationsSession struct {
	*storeStructuredMongo.Session
}
