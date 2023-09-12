package alerts

import (
	"context"
	"fmt"

	"github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository abstracts persistent storage for AlertsConfig data.
type Repository interface {
	Upsert(ctx context.Context, conf *Config) error
	Delete(ctx context.Context, conf *Config) error
}

// mongoRepo implements Repo, writing data to a MongoDB collection.
type mongoRepo mongo.Repository

// NewMongoRepo builds a Repo that writes AlertsConfig data to MongoDB.
func NewMongoRepo(repo *mongo.Repository) *mongoRepo {
	r := mongoRepo(*repo)
	return &r
}

// Upsert will create or update the given AlertsConfig.
func (r *mongoRepo) Upsert(ctx context.Context, conf *Config) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"ownerID": conf.OwnerID, "invitorID": conf.InvitorID}
	_, err := r.UpdateOne(ctx, filter, bson.M{"$set": conf}, opts)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}

// Delete will delete the given AlertsConfig.
func (r *mongoRepo) Delete(ctx context.Context, conf *Config) error {
	filter := bson.M{"ownerID": conf.OwnerID, "invitorID": conf.InvitorID}
	_, err := r.DeleteMany(ctx, filter, nil)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}
