package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type EventsRepository struct {
	*storeStructuredMongo.Repository
}

func (e *EventsRepository) EnsureIndexes() error {
	return e.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIdTypeUnique"),
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "time", Value: 1},
			},
			Options: options.Index().
				SetName("TypeTime"),
		},
	})
}

func (e *EventsRepository) GetStore() *storeStructuredMongo.Repository {
	return e.Repository
}
