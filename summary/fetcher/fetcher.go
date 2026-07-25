package fetcher

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
)

type DeviceDataCursor interface {
	Decode(val any) error
	RemainingBatchLength() int
	Next(ctx context.Context) bool
	GetNextBatch(ctx context.Context) ([]data.Datum, error)
	Err() error
}

type DeviceDataFetcher interface {
	GetDataSet(ctx context.Context, dataSetID string) (*data.DataSet, error)
	GetLastUpdatedForUser(ctx context.Context, userID string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error)
	GetDataRange(ctx context.Context, userID string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error)
}
