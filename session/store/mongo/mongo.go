package mongo

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/store/mongo"
)

func New(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	baseStore, err := mongo.New(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*mongo.Store
}

func (s *Store) NewSessionsSession() store.SessionsSession {
	return &SessionsSession{
		Session: s.Store.NewSession("tokens"),
	}
}

type SessionsSession struct {
	*mongo.Session
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

	startTime := time.Now()

	selector := bson.M{
		"userId": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroySessionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy sessions for user by id")
	}
	return nil
}
