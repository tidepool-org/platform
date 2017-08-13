package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

func (s *Store) NewConfirmationsSession(logger log.Logger) store.ConfirmationsSession {
	return &ConfirmationsSession{
		Session: s.Store.NewSession(logger, "confirmations"),
	}
}

type ConfirmationsSession struct {
	*mongo.Session
}

func (c *ConfirmationsSession) DestroyConfirmationsForUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if c.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"$or": []bson.M{
			{"userId": userID},
			{"creatorId": userID},
		},
	}
	removeInfo, err := c.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	c.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyConfirmationsForUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy confirmations for user by id")
	}
	return nil
}
