package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/errors"
	structuredmongo "github.com/tidepool-org/platform/store/structured/mongo"
)

// alertsRepo implements alerts.Repository, writing data to a MongoDB collection.
type alertsRepo structuredmongo.Repository

// Upsert will create or update the given Config.
func (r *alertsRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	opts := options.Update().SetUpsert(true)
	_, err := r.UpdateOne(ctx, r.filter(conf), bson.M{"$set": conf}, opts)
	if err != nil {
		return errors.Wrap(err, "unable to upsert config")
	}
	return nil
}

// Delete will delete the given Config.
func (r *alertsRepo) Delete(ctx context.Context, cfg *alerts.Config) error {
	_, err := r.DeleteMany(ctx, r.filter(cfg), nil)
	if err != nil {
		return errors.Wrap(err, "unable to delete config")
	}
	return nil
}

// Get will retrieve the given Config.
func (r *alertsRepo) Get(ctx context.Context, cfg *alerts.Config) (*alerts.Config, error) {
	out := &alerts.Config{}

	err := r.FindOne(ctx, r.filter(cfg)).Decode(&out)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get config")
	}

	return out, nil
}

// EnsureIndexes to maintain index constraints.
func (r *alertsRepo) EnsureIndexes() error {
	repo := structuredmongo.Repository(*r)
	return (&repo).CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "followedUserId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIdFollowedUserIdTypeUnique"),
		},
	})
}

func (r *alertsRepo) filter(cfg *alerts.Config) interface{} {
	return &alerts.Config{
		UserID:         cfg.UserID,
		FollowedUserID: cfg.FollowedUserID,
	}
}
