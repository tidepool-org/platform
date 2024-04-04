package fetcher

import (
	"context"
	"fmt"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary"
	"go.mongodb.org/mongo-driver/mongo"
)

type DatumCreator func() data.Datum
type DataCursorFactory func(cursor *mongo.Cursor) summary.DeviceDataCursor

func NewDefaultCursor(c *mongo.Cursor, create DatumCreator) summary.DeviceDataCursor {
	return &DefaultCursor{
		c:      c,
		create: create,
	}
}

var _ summary.DeviceDataCursor = &DefaultCursor{}

type DefaultCursor struct {
	c           *mongo.Cursor
	create      DatumCreator
	isExhausted bool
}

func (d *DefaultCursor) Decode(val interface{}) error {
	return d.c.Decode(val)
}

func (d *DefaultCursor) RemainingBatchLength() int {
	return d.c.RemainingBatchLength()
}

func (d *DefaultCursor) Next(ctx context.Context) bool {
	d.isExhausted = d.c.Next(ctx)
	return d.isExhausted
}

func (d *DefaultCursor) Close(ctx context.Context) error {
	return d.c.Close(ctx)
}

func (c *DefaultCursor) GetNextBatch(ctx context.Context) ([]data.Datum, error) {
	if c.isExhausted == true {
		return nil, ErrCursorExhausted
	}

	userData := make([]data.Datum, 0, c.RemainingBatchLength())
	for c.Next(ctx) && c.RemainingBatchLength() != 0 {
		datum := c.create()
		if err := c.Decode(datum); err != nil {
			return nil, fmt.Errorf("unable to decode userData: %w", err)
		}
	}

	return userData, c.Err()
}

func (d *DefaultCursor) Err() error {
	return d.c.Err()
}
