package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type SummaryRepository struct {
	*storeStructuredMongo.Repository
}

func (d *SummaryRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIDTypeUnique"),
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "dates.outdatedSince", Value: 1},
				{Key: "config.schemaVersion", Value: 1},
				{Key: "dates.lastUpdatedDate", Value: 1},
			},
			Options: options.Index().
				SetName("OutdatedSinceSchemaLastUpdated"),
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "dates.outdatedSince", Value: 1},
				{Key: "dates.lastUpdatedDate", Value: 1},
				{Key: "config.schemaVersion", Value: 1},
			},
			Options: options.Index().
				SetName("OutdatedAndSchemaMigration"),
		},
	})
}

func (d *SummaryRepository) GetStore() *storeStructuredMongo.Repository {
	return d.Repository
}

type BucketsRepository struct {
	*storeStructuredMongo.Repository
}

func (d *BucketsRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "time", Value: -1},
			},
			Options: options.Index().
				// This could be made unique without issue, but the performance cost is probably not worth it.
				SetName("UserIdTypeTime"),
		},
	})
}

func (d *BucketsRepository) GetStore() *storeStructuredMongo.Repository {
	return d.Repository
}
