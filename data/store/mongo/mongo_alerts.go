package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/store/structured/mongo"
)

// alertsRepo implements alerts.Repository, writing data to a MongoDB collection.
type alertsRepo mongo.Repository

// Upsert will create or update the given AlertsConfig.
func (r *alertsRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"ownerID": conf.OwnerID, "invitorID": conf.InvitorID}
	_, err := r.UpdateOne(ctx, filter, bson.M{"$set": conf}, opts)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}

// Delete will delete the given AlertsConfig.
func (r *alertsRepo) Delete(ctx context.Context, conf *alerts.Config) error {
	filter := bson.M{"ownerID": conf.OwnerID, "invitorID": conf.InvitorID}
	_, err := r.DeleteMany(ctx, filter, nil)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}
