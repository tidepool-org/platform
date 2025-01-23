package fetcher

import (
	"context"
	"github.com/tidepool-org/platform/data/summary/types"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
)

type DeviceDataCursor interface {
	Decode(val interface{}) error
	RemainingBatchLength() int
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	GetNextBatch(ctx context.Context) ([]data.Datum, error)
	Err() error
}

type BucketCursor[P types.BucketDataPt[B], B types.BucketData] interface {
	Decode(val *types.Bucket[P, B]) error
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	Err() error
}

// TODO: What's the purpose of this?
// TODO: Remove, not used anymore
type AnyCursor interface {
	Decode(val interface{}) error
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	Err() error
}

type DeviceDataFetcher interface {
	GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error)
	GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error)
	GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error)
}
