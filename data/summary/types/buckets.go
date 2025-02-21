package types

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data"
)

const minutesPerDay = 60 * 24

type BaseBucket struct {
	ID        primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserId    string             `json:"userId" bson:"userId"`
	Type      string             `json:"type" bson:"type"`
	Time      time.Time          `json:"time" bson:"time"`
	FirstData time.Time          `json:"firstTime" bson:"firstTime"`
	LastData  time.Time          `json:"lastTime" bson:"lastTime"`

	modified bool
}

func (bb *BaseBucket) Update(datumTime *time.Time) error {
	// check that datumTime is within the bucket bounds
	bucketStart := bb.Time
	bucketEnd := bb.Time.Add(time.Hour)

	if datumTime.Before(bucketStart) || datumTime.After(bucketEnd) {
		return fmt.Errorf("datum with time %s is outside the bounds of bucket with bounds %s - %s", datumTime, bucketStart, bucketEnd)
	}

	if bb.FirstData.IsZero() || datumTime.Before(bb.FirstData) {
		bb.FirstData = *datumTime
		bb.SetModified()
	}

	if bb.LastData.IsZero() || datumTime.After(bb.LastData) {
		bb.LastData = *datumTime
		bb.SetModified()
	}

	return nil
}

func (bb *BaseBucket) SetModified() {
	bb.modified = true
}

func (bb *BaseBucket) IsModified() bool {
	return bb.modified
}

type BucketData interface {
	GlucoseBucket | ContinuousBucket
}

type BucketDataPt[B BucketData] interface {
	*B
	Update(record data.Datum, lastData *time.Time) (bool, error)
}

type Bucket[PB BucketDataPt[B], B BucketData] struct {
	BaseBucket `json:",inline" bson:",inline"`
	Data       PB `json:"data" bson:"data"`
}

func NewBucket[PB BucketDataPt[B], B BucketData](userId string, date time.Time, summaryType string) *Bucket[PB, B] {
	return &Bucket[PB, B]{
		BaseBucket: BaseBucket{
			UserId: userId,
			Type:   summaryType,
			Time:   date,
		},
		Data: new(B),
	}
}

func (b *Bucket[PB, B]) Update(record data.Datum) error {
	updated, err := b.Data.Update(record, &b.BaseBucket.LastData)
	if err != nil {
		return err
	}

	if updated {
		err = b.BaseBucket.Update(record.GetTime())
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateBucketForUser[PB BucketDataPt[B], B BucketData](userId string, typ string) BucketFactoryFn[PB, B] {
	return func(recordHour time.Time) *Bucket[PB, B] {
		return NewBucket[PB](userId, recordHour, typ)
	}
}

type BucketFactoryFn[PB BucketDataPt[B], B BucketData] func(recordHour time.Time) *Bucket[PB, B]

type BucketsByTime[PB BucketDataPt[B], B BucketData] map[time.Time]*Bucket[PB, B]

func (BT BucketsByTime[PB, B]) Update(createBucket BucketFactoryFn[PB, B], userData []data.Datum) error {
	for _, r := range userData {
		// truncate time is not timezone/DST safe here, even if we do expect UTC, never truncate non-utc
		recordHour := r.GetTime().UTC().Truncate(time.Hour)

		if _, ok := BT[recordHour]; !ok {
			// we don't already have a bucket for this data
			BT[recordHour] = createBucket(recordHour)

			// fresh bucket, pull LastData from previous hour if possible for dedup
			if _, ok = BT[recordHour.Add(-time.Hour)]; ok {
				BT[recordHour].BaseBucket.LastData = BT[recordHour.Add(-time.Hour)].LastData
			}
		}

		if err := BT[recordHour].Update(r); err != nil {
			return err
		}
	}

	return nil
}
