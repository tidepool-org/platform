package fetcher

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
)

type DatumCreator func(typ string) (data.Datum, error)
type DataCursorFactory func(cursor *mongo.Cursor) DeviceDataCursor

type datumType struct {
	Type string `bson:"type"`
}

func NewDefaultCursor(c *mongo.Cursor, create DatumCreator) DeviceDataCursor {
	return &DefaultCursor{
		c:      c,
		create: create,
	}
}

var _ DeviceDataCursor = &DefaultCursor{}

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
	d.isExhausted = !d.c.Next(ctx)
	return !d.isExhausted
}

func (d *DefaultCursor) Close(ctx context.Context) error {
	return d.c.Close(ctx)
}

func (d *DefaultCursor) GetNextBatch(ctx context.Context) ([]data.Datum, error) {
	if d.isExhausted == true {
		return nil, ErrCursorExhausted
	}

	var userData []data.Datum
	for d.Next(ctx) {
		if userData == nil {
			userData = make([]data.Datum, 0, d.RemainingBatchLength())
		}

		typ := &datumType{}
		if err := d.Decode(typ); err != nil {
			return nil, fmt.Errorf("unable to decode datum type: %w", err)
		}

		datum, err := d.create(typ.Type)
		if err != nil {
			return nil, fmt.Errorf("unable to create datum with type: %w", err)
		}

		if err := d.Decode(datum); err != nil {
			return nil, fmt.Errorf("unable to decode userData: %w", err)
		}

		userData = append(userData, datum)

		if d.RemainingBatchLength() == 0 {
			break
		}
	}

	return userData, d.Err()
}

func (d *DefaultCursor) Err() error {
	return d.c.Err()
}
