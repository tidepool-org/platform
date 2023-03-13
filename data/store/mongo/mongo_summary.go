package mongo

import (
	"context"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type SummaryRepository struct {
	*storeStructuredMongo.Repository
}

func (d *SummaryRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{})
}
