package fetcher

import (
	"context"
	"fmt"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary"
)

func NewContinuousDeviceDataCursor(cursor summary.DeviceDataCursor, fetcher summary.DeviceDataFetcher, create DatumCreator) *ContinuousDeviceDataCursor {
	return &ContinuousDeviceDataCursor{
		DeviceDataCursor: cursor,

		create:        create,
		fetcher:       fetcher,
		uploadIdCache: map[string]bool{},
	}
}

var _ summary.DeviceDataCursor = &ContinuousDeviceDataCursor{}

type ContinuousDeviceDataCursor struct {
	summary.DeviceDataCursor

	create        DatumCreator
	fetcher       summary.DeviceDataFetcher
	uploadIdCache map[string]bool
}

func (c *ContinuousDeviceDataCursor) GetNextBatch(ctx context.Context) ([]data.Datum, error) {
	userData := make([]data.Datum, 0, c.RemainingBatchLength())

	for c.RemainingBatchLength() != 0 {
		datum := c.create()
		if err := c.Decode(datum); err != nil {
			return nil, fmt.Errorf("unable to decode userData: %w", err)
		}
		if isContinuous, err := c.isUploadContinuous(ctx, datum.GetUploadID()); err != nil {
			return nil, err
		} else if isContinuous {
			userData = append(userData, datum)
		}

		c.Next(ctx)
	}

	return userData, c.Err()
}

func (c *ContinuousDeviceDataCursor) isUploadContinuous(ctx context.Context, uploadId *string) (bool, error) {
	if uploadId == nil {
		return false, nil
	}

	// check if we already cached if the uploadId is continuous or not, query if unknown
	if _, ok := c.uploadIdCache[*uploadId]; !ok {
		uploadRecord, err := c.fetcher.GetDataSetByID(ctx, *uploadId)
		if err != nil {
			return false, err
		}

		if uploadRecord != nil && uploadRecord.HasDataSetTypeContinuous() {
			c.uploadIdCache[*uploadId] = true
		} else {
			c.uploadIdCache[*uploadId] = false
		}
	}

	return c.uploadIdCache[*uploadId], nil
}
