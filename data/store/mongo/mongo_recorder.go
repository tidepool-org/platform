package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/errors"
	structuredmongo "github.com/tidepool-org/platform/store/structured/mongo"
)

// recorderRepo implements RecorderRepository, writing data to a MongoDB collection.
type recorderRepo structuredmongo.Repository

func (r *recorderRepo) RecordReceivedDeviceData(ctx context.Context,
	lastComm alerts.LastCommunication) error {

	opts := options.Update().SetUpsert(true)
	_, err := r.UpdateOne(ctx, r.filter(lastComm), bson.M{"$set": lastComm}, opts)
	if err != nil {
		return fmt.Errorf("upserting alerts.LastCommunication: %w", err)
	}
	return nil
}

func (r *recorderRepo) EnsureIndexes() error {
	repo := structuredmongo.Repository(*r)
	return (&repo).CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "lastReceivedDeviceData", Value: 1},
			},
			Options: options.Index().
				SetName("LastReceivedDeviceData"),
		},
		{
			Keys: bson.D{
				{Key: "dataSetId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("DataSetIdUnique"),
		},
	})
}

func (r *recorderRepo) filter(lastComm alerts.LastCommunication) map[string]any {
	return map[string]any{
		"userId":    lastComm.UserID,
		"dataSetId": lastComm.DataSetID,
	}
}

func (d *recorderRepo) UsersWithoutCommunication(ctx context.Context) ([]alerts.LastCommunication, error) {
	start := time.Now().Add(-5 * time.Minute)
	selector := bson.M{
		"lastReceivedDeviceData": bson.M{"$lte": start},
	}
	findOptions := options.Find().SetSort(bson.D{{Key: "lastReceivedDeviceData", Value: 1}})
	cursor, err := d.Find(ctx, selector, findOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list users without communication")
	}
	records := []alerts.LastCommunication{}
	if err := cursor.All(ctx, &records); err != nil {
		return nil, errors.Wrapf(err, "Unable to iterate users without communication cursor")
	}
	return records, nil
}
