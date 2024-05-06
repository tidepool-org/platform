package mongo

import (
	"context"
	"fmt"

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
		return fmt.Errorf("upserting alerts.Config: %w", err)
	}
	return nil
}

// Delete will delete the given Config.
func (r *alertsRepo) Delete(ctx context.Context, cfg *alerts.Config) error {
	_, err := r.DeleteMany(ctx, r.filter(cfg), nil)
	if err != nil {
		return fmt.Errorf("upserting alerts.Config: %w", err)
	}
	return nil
}

// List will retrieve any Configs that are defined by followers of the given user.
func (r *alertsRepo) List(ctx context.Context, followedUserID string) ([]*alerts.Config, error) {
	filter := bson.D{
		{Key: "followedUserId", Value: followedUserID},
	}
	cursor, err := r.Find(ctx, filter, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list alerts.Config(s) for followed user %s", followedUserID)
	}
	defer cursor.Close(ctx)
	out := []*alerts.Config{}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, errors.Wrapf(err, "Unable to decode alerts.Config(s) for followed user %s", followedUserID)
	}
	if err := cursor.Err(); err != nil {
		return nil, errors.Wrapf(err, "Unexpected error for followed user %s", followedUserID)
	}
	return out, nil
}

// Get will retrieve the given Config.
func (r *alertsRepo) Get(ctx context.Context, cfg *alerts.Config) (*alerts.Config, error) {
	res := r.FindOne(ctx, r.filter(cfg), nil)
	if res.Err() != nil {
		return nil, fmt.Errorf("getting alerts.Config: %w", res.Err())
	}
	out := &alerts.Config{}
	if err := res.Decode(out); err != nil {
		return nil, err
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
