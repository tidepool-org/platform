package mongo

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notification/store"
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

func (s *Store) NewNotificationsSession() store.NotificationsSession {
	return &NotificationsSession{
		Session: s.Store.NewSession("notifications"),
	}
}

type NotificationsSession struct {
	*mongo.Session
}
