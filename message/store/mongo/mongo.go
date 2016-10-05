package mongo

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/store/mongo"
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

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	baseSession, err := s.Store.NewSession(logger)
	if err != nil {
		return nil, err
	}

	return &Session{
		Session: baseSession,
	}, nil
}

type Session struct {
	*mongo.Session
}

func (s *Session) DeleteMessagesFromUser(user *store.User) error {
	if user == nil {
		return app.Error("mongo", "user is missing")
	}
	if user.ID == "" {
		return app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	// TODO: Add deletedTime/deletedUserId to user object???

	selector := bson.M{
		"userid": user.ID,
	}
	update := bson.M{
		"$unset": bson.M{"userid": ""},
		"$set": bson.M{
			"user": bson.M{
				"fullName": user.FullName,
			},
		},
	}
	changeInfo, err := s.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"userID": user.ID, "change-info": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteMessagesFromUser")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to delete messages from user")
	}

	return nil
}

func (s *Session) DestroyMessagesForUserByID(userID string) error {
	if userID == "" {
		return app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"groupid": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userID": userID, "remove-info": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyMessagesForUserByID")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to destroy messages for user by id")
	}

	return nil
}
