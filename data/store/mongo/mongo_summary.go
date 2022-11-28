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
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserID"),
		},
		{
			Keys: bson.D{
				{Key: "lastUpdatedDate", Value: 1},
			},
			Options: options.Index().
				SetName("LastUpdatedDate"),
		},
		{
			Keys: bson.D{
				{Key: "outdatedSince", Value: 1},
			},
			Options: options.Index().
				SetName("OutdatedSince"),
		},
	})
}
