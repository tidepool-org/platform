package mongo

import (
	"context"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/session/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(cfg *storeStructuredMongo.Config, lgr log.Logger) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) NewSessionsSession() store.SessionsSession {
	return &SessionsSession{
		Session: s.Store.NewSession("tokens"),
	}
}

type SessionsSession struct {
	*storeStructuredMongo.Session
}

func (s *SessionsSession) DestroySessionsForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if s.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()

	selector := bson.M{
		"userId": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroySessionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy sessions for user by id")
	}
	return nil
}
