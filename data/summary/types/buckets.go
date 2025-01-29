package types

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
)

// TODO: Not a fetcher as it updates buckets too
// TODO: Not used anymore, delete
type BucketFetcher[PB BucketDataPt[B], B BucketData] interface {
	GetBucketsByTime(ctx context.Context, userId string, startTime, endTime time.Time) (BucketsByTime[PB, B], error)
	GetAllBuckets(ctx context.Context, userId string) (*mongo.Cursor, error)
	WriteModifiedBuckets(ctx context.Context, buckets BucketsByTime[PB, B]) error
}

// TODO: define consts at the top
const minutesPerDay = 60 * 24

// TODO: Follow naming conventions in data service. Rename to BaseBucket
type BucketShared struct {
	ID        primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserId    string             `json:"userId" bson:"userId"`
	Type      string             `json:"type" bson:"type"`
	Time      time.Time          `json:"time" bson:"time"`
	FirstData time.Time          `json:"firstTime" bson:"firstTime"`
	LastData  time.Time          `json:"lastTime" bson:"lastTime"`

	modified bool
}

// TODO: single letter lower case pointer receiver
func (b *BucketShared) Update(datumTime *time.Time) error {
	// check that datumTime is within the bucket bounds
	bucketStart := b.Time
	bucketEnd := b.Time.Add(time.Hour)
	// TODO: both ends are exclusive
	if datumTime.Before(bucketStart) || datumTime.After(bucketEnd) {
		return fmt.Errorf("datum with time %s is outside the bounds of bucket with bounds %s - %s", datumTime, bucketStart, bucketEnd)
	}

	if b.FirstData.IsZero() || datumTime.Before(b.FirstData) {
		b.FirstData = *datumTime
		b.SetModified()
	}

	if b.LastData.IsZero() || datumTime.After(b.LastData) {
		b.LastData = *datumTime
		b.SetModified()
	}

	return nil
}

func (b *BucketShared) SetModified() {
	b.modified = true
}

func (b *BucketShared) IsModified() bool {
	return b.modified
}

type BucketData interface {
	GlucoseBucket | ContinuousBucket
}

type BucketDataPt[B BucketData] interface {
	*B
	Update(record data.Datum, base *BucketShared) (bool, error)
}

type Bucket[PB BucketDataPt[B], B BucketData] struct {
	BucketShared `json:",inline" bson:",inline"`
	Data         PB `json:"data" bson:"data"`
}

// TODO: Not clear what is type here. BGM/CGM/Glucose/SummaryType?
func NewBucket[PB BucketDataPt[B], B BucketData](userId string, date time.Time, typ string) *Bucket[PB, B] {
	return &Bucket[PB, B]{
		BucketShared: BucketShared{
			UserId: userId,
			Type:   typ,
			Time:   date,
		},
		Data: new(B),
	}
}

// TODO: single letter lowercase pointer receiver
func (BU *Bucket[PB, B]) Update(record data.Datum) error {
	updated, err := BU.Data.Update(record, &BU.BucketShared)
	if err != nil {
		return err
	}

	if updated {
		err = BU.BucketShared.Update(record.GetTime())
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

// TODO: simplify this by removing user id and type and set it by the caller or pass a factory fn
func (BT BucketsByTime[PB, B]) Update(createBucket BucketFactoryFn[PB, B], userData []data.Datum) error {
	for _, r := range userData {
		// truncate time is not timezone/DST safe here, even if we do expect UTC, never truncate non-utc
		recordHour := r.GetTime().UTC().Truncate(time.Hour)

		// OPTIMIZATION this could check if recordHour equal to previous hour, to save a map lookup, probably saves 0ms
		if _, ok := BT[recordHour]; !ok {
			// we don't already have a bucket for this data
			BT[recordHour] = createBucket(recordHour)

			// fresh bucket, pull LastData from previous hour if possible for dedup
			if _, ok = BT[recordHour.Add(-time.Hour)]; ok {
				BT[recordHour].BucketShared.LastData = BT[recordHour.Add(-time.Hour)].LastData
			}
		}

		err := BT[recordHour].Update(r)
		if err != nil {
			return err
		}

	}

	return nil
}
