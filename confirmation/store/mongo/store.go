package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(params storeStructuredMongo.Params) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	repository := s.ConfirmationRepository()
	return repository.EnsureIndexes()
}

func (s *Store) NewConfirmationRepository() store.ConfirmationRepository {
	return s.ConfirmationRepository()
}

func (s *Store) ConfirmationRepository() *ConfirmationRepository {
	return &ConfirmationRepository{
		s.Store.GetRepository("confirmations"),
	}
}

type ConfirmationRepository struct {
	*storeStructuredMongo.Repository
}

func (c *ConfirmationRepository) EnsureIndexes() error {
	return c.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		// Additional indexes are also created in `hydrophone`.
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
	})
}

func (c *ConfirmationRepository) DeleteUserConfirmations(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("userId", userID)

	selector := bson.M{
		"$or": []bson.M{
			{"userId": userID},
			{"creatorId": userID},
		},
	}
	changeInfo, err := c.DeleteMany(ctx, selector)
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteUserConfirmations")
	if err != nil {
		return errors.Wrap(err, "unable to delete user confirmations")
	}

	return nil
}
