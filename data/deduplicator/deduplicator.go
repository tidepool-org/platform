package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
)

type Factory interface {
	New(ctx context.Context, dataSet *data.DataSet) (Deduplicator, error)
	Get(ctx context.Context, dataSet *data.DataSet) (Deduplicator, error)
}

type Deduplicator interface {
	Open(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) (*data.DataSet, error)
	AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, dataSetData data.Data) error
	DeleteData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, selectors *data.Selectors) error
	Close(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) error
	Delete(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) error
}
