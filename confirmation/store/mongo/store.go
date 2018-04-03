package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
)

type Store struct {
	*mongo.Store
}

func NewStore(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	str, err := mongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	ssn := s.NewConfirmationSession()
	defer ssn.Close()
	return ssn.EnsureIndexes()
}

func (s *Store) NewConfirmationSession() store.ConfirmationSession {
	return &ConfirmationSession{
		Session: s.Store.NewSession("confirmations"),
	}
}

type ConfirmationSession struct {
	*mongo.Session
}

func (c *ConfirmationSession) EnsureIndexes() error {
	return c.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"email"}, Background: true},
		{Key: []string{"status"}, Background: true},
		{Key: []string{"type"}, Background: true},
		{Key: []string{"userId"}, Background: true},
	})
}

func (c *ConfirmationSession) DeleteUserConfirmations(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if c.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	selector := bson.M{
		"$or": []bson.M{
			{"userId": userID},
			{"creatorId": userID},
		},
	}
	changeInfo, err := c.C().RemoveAll(selector)
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteUserConfirmations")
	if err != nil {
		return errors.Wrap(err, "unable to delete user confirmations")
	}

	return nil
}
