package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
)

//go:generate mockgen -source=deduplicator.go -destination=test/deduplicator_mocks.go -package=test -typed

type Factory interface {
	New(ctx context.Context, dataSet *data.DataSet) (Deduplicator, error)
	Get(ctx context.Context, dataSet *data.DataSet) (Deduplicator, error)
}

type Deduplicator interface {
	Open(ctx context.Context, dataSet *data.DataSet) (*data.DataSet, error)
	AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error
	DeleteData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	Close(ctx context.Context, dataSet *data.DataSet) error
	Delete(ctx context.Context, dataSet *data.DataSet) error
}
