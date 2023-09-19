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

// Upsert will create or update the given Config.
func (r *alertsRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	opts := options.Update().SetUpsert(true)
	doc := NewAlertsConfigDocument(conf)
	filter := bson.M{"_id": doc.ID}
	_, err := r.UpdateOne(ctx, filter, bson.M{"$set": doc}, opts)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}

// Delete will delete the given Config.
func (r *alertsRepo) Delete(ctx context.Context, conf *alerts.Config) error {
	filter := bson.M{"_id": AlertsID(conf)}
	_, err := r.DeleteMany(ctx, filter, nil)
	if err != nil {
		return fmt.Errorf("upserting AlertsConfig: %w", err)
	}
	return nil
}

// AlertsConfigDocument wraps alerts.Config to provide an ID for mongodb.
type AlertsConfigDocument struct {
	ID             string `bson:"_id"`
	*alerts.Config `bson:",inline"`
}

func NewAlertsConfigDocument(cfg *alerts.Config) *AlertsConfigDocument {
	return &AlertsConfigDocument{
		ID:     AlertsID(cfg),
		Config: cfg,
	}
}

// AlertsID generates a unique ID for a mongo document.
func AlertsID(cfg *alerts.Config) string {
	return cfg.OwnerID + ":" + cfg.InvitorID
}
