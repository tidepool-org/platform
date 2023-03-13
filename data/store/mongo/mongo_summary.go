package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type SummaryRepository struct {
	*storeStructuredMongo.Repository
}

func (d *SummaryRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{})
}
