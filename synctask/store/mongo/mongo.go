package mongo

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/synctask/store"
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

func (s *Store) NewSyncTaskSession() store.SyncTaskSession {
	return &SyncTaskSession{
		Session: s.Store.NewSession("syncTasks"),
	}
}

type SyncTaskSession struct {
	*storeStructuredMongo.Session
}

func (s *SyncTaskSession) DestroySyncTasksForUserByID(ctx context.Context, userID string) error {
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
		"_userId": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroySyncTasksForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy sync tasks for user by id")
	}

	return nil
}
