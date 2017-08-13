package mongo

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notification/store"
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

func (s *Store) NewNotificationsSession(lgr log.Logger) store.NotificationsSession {
	return &NotificationsSession{
		Session: s.Store.NewSession(lgr, "notifications"),
	}
}

type NotificationsSession struct {
	*mongo.Session
}
