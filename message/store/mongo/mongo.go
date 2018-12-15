package mongo

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/message/store"
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

func (s *Store) NewMessagesSession() store.MessagesSession {
	return &MessagesSession{
		Session: s.Store.NewSession("messages"),
	}
}

type MessagesSession struct {
	*storeStructuredMongo.Session
}

func (m *MessagesSession) DeleteMessagesFromUser(ctx context.Context, user *store.User) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if user == nil {
		return errors.New("user is missing")
	}
	if user.ID == "" {
		return errors.New("user id is missing")
	}

	if m.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()

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
	changeInfo, err := m.C().UpdateAll(selector, update)

	loggerFields := log.Fields{"userId": user.ID, "changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteMessagesFromUser")

	if err != nil {
		return errors.Wrap(err, "unable to delete messages from user")
	}

	return nil
}

func (m *MessagesSession) DestroyMessagesForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if m.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()

	selector := bson.M{
		"groupid": userID,
	}
	removeInfo, err := m.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyMessagesForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy messages for user by id")
	}

	return nil
}
