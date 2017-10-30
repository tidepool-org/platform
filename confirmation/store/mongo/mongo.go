package mongo

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

func (s *Store) NewConfirmationsSession() store.ConfirmationsSession {
	return &ConfirmationsSession{
		Session: s.Store.NewSession("confirmations"),
	}
}

type ConfirmationsSession struct {
	*mongo.Session
}

func (c *ConfirmationsSession) DestroyConfirmationsForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if c.IsClosed() {
		return errors.New("session closed")
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
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyConfirmationsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy confirmations for user by id")
	}
	return nil
}
