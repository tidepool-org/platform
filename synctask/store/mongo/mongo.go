package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/synctask/store"
)

func New(logger log.Logger, config *mongo.Config) (*Store, error) {
	baseStore, err := mongo.New(logger, config)
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

func (s *Store) NewSyncTasksSession(logger log.Logger) store.SyncTasksSession {
	return &SyncTasksSession{
		Session: s.Store.NewSession(logger, "syncTasks"),
	}
}

type SyncTasksSession struct {
	*mongo.Session
}

func (s *SyncTasksSession) DestroySyncTasksForUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"_userId": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroySyncTasksForUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy sync tasks for user by id")
	}

	return nil
}
