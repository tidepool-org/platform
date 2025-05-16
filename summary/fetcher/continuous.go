package fetcher

import (
	"context"
	"fmt"

	"github.com/tidepool-org/platform/data"
)

var ErrCursorExhausted = fmt.Errorf("cursor is exhausted")

func NewContinuousDeviceDataCursor(cursor DeviceDataCursor, fetcher DeviceDataFetcher, create DatumCreator) *ContinuousDeviceDataCursor {
	return &ContinuousDeviceDataCursor{
		DeviceDataCursor: cursor,

		create:        create,
		fetcher:       fetcher,
		uploadIdCache: map[string]bool{},
	}
}

var _ DeviceDataCursor = &ContinuousDeviceDataCursor{}

type ContinuousDeviceDataCursor struct {
	DeviceDataCursor

	create        DatumCreator
	fetcher       DeviceDataFetcher
	isExhausted   bool
	uploadIdCache map[string]bool
}

func (c *ContinuousDeviceDataCursor) GetNextBatch(ctx context.Context) ([]data.Datum, error) {
	if c.isExhausted == true {
		return nil, ErrCursorExhausted
	}

	var userData []data.Datum
	for c.Next(ctx) {
		if userData == nil {
			userData = make([]data.Datum, 0, c.RemainingBatchLength())
		}

		datum := c.create()
		if err := c.Decode(datum); err != nil {
			return nil, fmt.Errorf("unable to decode userData: %w", err)
		}

		if isContinuous, err := c.isUploadContinuous(ctx, datum.GetUploadID()); err != nil {
			return nil, err
		} else if isContinuous {
			userData = append(userData, datum)
		}

		if c.RemainingBatchLength() == 0 {
			break
		}
	}

	return userData, c.Err()
}

func (c *ContinuousDeviceDataCursor) Next(ctx context.Context) bool {
	c.isExhausted = !c.DeviceDataCursor.Next(ctx)
	return !c.isExhausted
}

func (c *ContinuousDeviceDataCursor) isUploadContinuous(ctx context.Context, uploadId *string) (bool, error) {
	if uploadId == nil {
		return false, nil
	}

	// check if we already cached if the uploadId is continuous or not, query if unknown
	if _, ok := c.uploadIdCache[*uploadId]; !ok {
		uploadRecord, err := c.fetcher.GetDataSet(ctx, *uploadId)
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
