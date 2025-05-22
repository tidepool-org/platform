package fetcher

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
)

type DeviceDataCursor interface {
	Decode(val interface{}) error
	RemainingBatchLength() int
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	GetNextBatch(ctx context.Context) ([]data.Datum, error)
	Err() error
}

type DeviceDataFetcher interface {
	GetDataSetByID(ctx context.Context, dataSetID string) (*data.DataSet, error)
	GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error)
	GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error)
}
