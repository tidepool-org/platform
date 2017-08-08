package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
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

func (s *Store) NewSession(logger log.Logger) store.Session {
	return &Session{
		Session: s.Store.NewSession(logger),
	}
}

type Session struct {
	*mongo.Session
}

func (s *Session) DeleteMessagesFromUser(user *store.User) error {
	if user == nil {
		return errors.New("mongo", "user is missing")
	}
	if user.ID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
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

	loggerFields := log.Fields{"userId": user.ID, "changeInfo": changeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteMessagesFromUser")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to delete messages from user")
	}

	return nil
}

func (s *Session) DestroyMessagesForUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"groupid": userID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyMessagesForUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy messages for user by id")
	}

	return nil
}
