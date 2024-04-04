package fetcher

import (
	"context"
	"fmt"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary"
	"go.mongodb.org/mongo-driver/mongo"
)

type DatumCreator func() data.Datum

func NewDefaultCursor(c *mongo.Cursor, create DatumCreator) summary.DeviceDataCursor {
	return &DefaultCursor{
		c:      c,
		create: create,
	}
}

var _ summary.DeviceDataCursor = &DefaultCursor{}

type DefaultCursor struct {
	c      *mongo.Cursor
	create DatumCreator
}

func (d *DefaultCursor) Decode(val interface{}) error {
	return d.c.Decode(val)
}

func (d *DefaultCursor) RemainingBatchLength() int {
	return d.c.RemainingBatchLength()
}

func (d *DefaultCursor) Next(ctx context.Context) bool {
	return d.c.Next(ctx)
}

func (d *DefaultCursor) Close(ctx context.Context) error {
	return d.c.Close(ctx)
}

func (d *DefaultCursor) GetNextBatch(ctx context.Context) ([]data.Datum, error) {
	userData := make([]data.Datum, 0, d.c.RemainingBatchLength())

	for d.c.RemainingBatchLength() != 0 {
		datum := d.create()
		if err := d.c.Decode(datum); err != nil {
			return nil, fmt.Errorf("unable to decode userData: %w", err)
		}
		userData = append(userData, datum)

		d.c.Next(ctx)
	}

	return userData, d.Err()
}

func (d *DefaultCursor) Err() error {
	return d.c.Err()
}
