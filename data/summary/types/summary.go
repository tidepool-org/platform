package types

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"

	"go.mongodb.org/mongo-driver/bson/primitive"

	mapset "github.com/deckarep/golang-set/v2"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	insulinDatum "github.com/tidepool-org/platform/data/types/insulin"
)

const (
	SummaryTypeCGM = "cgm"
	SummaryTypeBGM = "bgm"
	SchemaVersion  = 3

	lowBloodGlucose      = 3.9
	veryLowBloodGlucose  = 3.0
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	HoursAgoToKeep       = 60 * 24

	OutdatedReasonUploadCompleted = "UPLOAD_COMPLETED"
	OutdatedReasonDataAdded       = "DATA_ADDED"
	OutdatedReasonSchemaMigration = "SCHEMA_MIGRATION"
	OutdatedReasonBackfill        = "BACKFILL"
)

var stopPoints = [...]int{1, 7, 14, 30}

var DeviceDataTypes = []string{continuous.Type, selfmonitored.Type}
var DeviceDataTypesSet = mapset.NewSet[string](DeviceDataTypes...)

var DeviceDataToSummaryTypes = map[string]string{
	continuous.Type:    SummaryTypeCGM,
	selfmonitored.Type: SummaryTypeBGM,
}

type OutdatedSummariesResponse struct {
	UserIds []string  `json:"userIds"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

type BucketData interface {
	CGMBucketData | BGMBucketData
}

type RecordTypes interface {
	glucoseDatum.Glucose | insulinDatum.Insulin
}

type RecordTypesPt[T RecordTypes] interface {
	*T
	GetTime() *time.Time
}

type DeviceDataCursor interface {
	Decode(val interface{}) error
	RemainingBatchLength() int
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
}

type UserLastUpdated struct {
	FirstData time.Time
	LastData  time.Time

	EarliestModified time.Time

	LastUpload time.Time

	LastUpdated     time.Time
	NextLastUpdated time.Time
}

type Config struct {
	SchemaVersion int `json:"schemaVersion" bson:"schemaVersion"`

	// these are just constants right now.
	HighGlucoseThreshold     float64 `json:"highGlucoseThreshold" bson:"highGlucoseThreshold"`
	VeryHighGlucoseThreshold float64 `json:"veryHighGlucoseThreshold" bson:"veryHighGlucoseThreshold"`
	LowGlucoseThreshold      float64 `json:"lowGlucoseThreshold" bson:"lowGlucoseThreshold"`
	VeryLowGlucoseThreshold  float64 `json:"VeryLowGlucoseThreshold" bson:"VeryLowGlucoseThreshold"`
}

type Dates struct {
	LastUpdatedDate   time.Time `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	LastUpdatedReason []string  `json:"lastUpdatedReason" bson:"lastUpdatedReason"`

	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    *time.Time `json:"lastUploadDate" bson:"lastUploadDate"`

	HasFirstData bool       `json:"hasFirstData" bson:"hasFirstData"`
	FirstData    *time.Time `json:"firstData" bson:"firstData"`

	HasLastData bool       `json:"hasLastData" bson:"hasLastData"`
	LastData    *time.Time `json:"lastData" bson:"lastData"`

	HasOutdatedSince bool       `json:"hasOutdatedSince" bson:"hasOutdatedSince"`
	OutdatedSince    *time.Time `json:"outdatedSince" bson:"outdatedSince"`
	OutdatedReason   []string   `json:"outdatedReason" bson:"outdatedReason"`
}

func (d *Dates) Update(status *UserLastUpdated, firstBucketDate time.Time) {
	d.LastUpdatedDate = status.NextLastUpdated
	d.LastUpdatedReason = d.OutdatedReason

	d.HasLastUploadDate = true
	d.LastUploadDate = &status.LastUpload

	d.HasFirstData = true
	d.FirstData = &firstBucketDate

	d.HasLastData = true
	d.LastData = &status.LastData

	d.HasOutdatedSince = false
	d.OutdatedSince = nil
	d.OutdatedReason = nil
}

type Bucket[S BucketDataPt[T], T BucketData] struct {
	Date           time.Time `json:"date" bson:"date"`
	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`

	Data S `json:"data" bson:"data"`
}

type BucketDataPt[T BucketData] interface {
	*T
	CalculateStats(interface{}, *time.Time) (bool, error)
}

func CreateBucket[A BucketDataPt[T], T BucketData](t time.Time) *Bucket[A, T] {
	bucket := new(Bucket[A, T])
	bucket.Date = t
	bucket.Data = new(T)
	return bucket
}

type Stats interface {
	CGMStats | BGMStats
}

type StatsPt[T Stats] interface {
	*T
	GetType() string
	GetDeviceDataType() string
	Init()
	GetBucketsLen() int
	GetBucketDate(int) time.Time
	Update(context.Context, DeviceDataCursor) error
	ClearInvalidatedBuckets(earliestModified time.Time) time.Time
}

type Summary[T Stats, A StatsPt[T]] struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string             `json:"type" bson:"type"`
	UserID string             `json:"userId" bson:"userId"`

	Config Config `json:"config" bson:"config"`

	Dates Dates `json:"dates" bson:"dates"`
	Stats A     `json:"stats" bson:"stats"`
}

func NewConfig() Config {
	return Config{
		SchemaVersion:            SchemaVersion,
		HighGlucoseThreshold:     highBloodGlucose,
		VeryHighGlucoseThreshold: veryHighBloodGlucose,
		LowGlucoseThreshold:      lowBloodGlucose,
		VeryLowGlucoseThreshold:  veryLowBloodGlucose,
	}
}

func (s *Summary[T, A]) SetOutdated(reason string) {
	set := mapset.NewSet[string](reason)
	if len(s.Dates.OutdatedReason) > 0 {
		set.Append(s.Dates.OutdatedReason...)
	}

	if reason == OutdatedReasonSchemaMigration {
		*s = *Create[A](s.UserID)
	}

	s.Dates.OutdatedReason = set.ToSlice()

	if s.Dates.OutdatedSince == nil {
		s.Dates.OutdatedSince = pointer.FromAny(time.Now().Truncate(time.Millisecond).UTC())
		s.Dates.HasOutdatedSince = true
	}
}

func NewDates() Dates {
	return Dates{
		LastUpdatedDate:   time.Time{},
		LastUpdatedReason: nil,

		HasLastUploadDate: false,
		LastUploadDate:    nil,

		HasFirstData: false,
		FirstData:    nil,

		HasLastData: false,
		LastData:    nil,

		HasOutdatedSince: false,
		OutdatedSince:    nil,
		OutdatedReason:   nil,
	}
}

func Create[A StatsPt[T], T Stats](userId string) *Summary[T, A] {
	s := new(Summary[T, A])
	s.UserID = userId
	s.Stats = new(T)
	s.Stats.Init()
	s.Type = s.Stats.GetType()
	s.Config = NewConfig()
	s.Dates = NewDates()

	return s
}

func GetTypeString[T Stats, A StatsPt[T]]() string {
	s := new(Summary[T, A])
	return s.Stats.GetType()
}

func GetDeviceDataTypeString[T Stats, A StatsPt[T]]() string {
	s := new(Summary[T, A])
	return s.Stats.GetDeviceDataType()
}

type Period interface {
	BGMPeriod | CGMPeriod
}

func AddBin[T BucketData, A BucketDataPt[T]](buckets *[]*Bucket[A, T], newBucket *Bucket[A, T]) error {
	if len(*buckets) == 0 {
		*buckets = append(*buckets, newBucket)
		return nil
	}

	if lastBucket := (*buckets)[len(*buckets)-1]; newBucket.Date.After(lastBucket.Date) {
		return addBinAfter(buckets, newBucket)
	} else if firstBucket := (*buckets)[0]; newBucket.Date.Before(firstBucket.Date) {
		return addBinBefore(buckets, newBucket)
	}
	return replaceBin(buckets, newBucket)
}

// MaxBucketGap denotes the duration after which a bucket isn't useful.
const MaxBucketGap = -time.Hour * HoursAgoToKeep

// addBinAfter readjusts buckets so that newBucket is at the end.
//
// addBinAfter assumes that newBucket comes after the last element of
// buckets. Any gaps between buckets and newBucket are padded appropriately.
func addBinAfter[A BucketDataPt[T], T BucketData](buckets *[]*Bucket[A, T], newBucket *Bucket[A, T]) error {
	var newDate = newBucket.Date
	var lastBucket = (*buckets)[len(*buckets)-1]

	if newDate.Add(MaxBucketGap).After(lastBucket.Date) {
		*buckets = []*Bucket[A, T]{newBucket}
		return nil
	}

	var gapStart, gapEnd = lastBucket.Date.Add(time.Hour), newDate
	var gapBucketsLen = int(newDate.Sub(lastBucket.Date).Hours())
	var gapBuckets = make([]*Bucket[A, T], 0, gapBucketsLen)
	for i := gapStart; i.Before(gapEnd); i = i.Add(time.Hour) {
		gapBuckets = append(gapBuckets, CreateBucket[A](i))
	}
	*buckets = append(*buckets, gapBuckets...)
	*buckets = append(*buckets, newBucket)

	removeExcessBuckets(buckets)

	return nil
}

// addBinBefore readjusts buckets to that newBucket is at the start.
//
// addBinBefore assumes that newBucket comes before the first element of
// buckets. Any gaps between buckets and newBucket are padded appropriately.
func addBinBefore[T BucketData, A BucketDataPt[T]](buckets *[]*Bucket[A, T], newBucket *Bucket[A, T]) error {
	var newDate = newBucket.Date
	var lastBucket = (*buckets)[len(*buckets)-1]

	if newDate.Before(lastBucket.Date.Add(MaxBucketGap)) {
		return errors.New("bucket is too old")
	}

	var firstBucket = (*buckets)[0]
	var gapStart, gapEnd = newDate.Add(time.Hour), firstBucket.Date
	var gapBucketsLen = Abs(int(firstBucket.Date.Sub(newDate).Hours()))
	var gapBuckets = make([]*Bucket[A, T], 0, gapBucketsLen)
	for i := gapStart; i.Before(gapEnd); i = i.Add(time.Hour) {
		gapBuckets = append(gapBuckets, CreateBucket[A](i))
	}

	*buckets = append(gapBuckets, *buckets...)
	*buckets = append([]*Bucket[A, T]{newBucket}, *buckets...)

	removeExcessBuckets(buckets)

	return nil
}

func replaceBin[A BucketDataPt[T], T BucketData](buckets *[]*Bucket[A, T], newBucket *Bucket[A, T]) error {
	var newDate = newBucket.Date
	var offset = int(newDate.Sub((*buckets)[0].Date).Hours())
	var toReplace = (*buckets)[offset]
	if !toReplace.Date.Equal(newDate) {
		return fmt.Errorf("potentially damaged buckets, offset jump did not find intended record. Found %s, wanted %s",
			toReplace.Date, newDate)
	}
	(*buckets)[offset] = newBucket
	return nil
}

func removeExcessBuckets[A BucketDataPt[T], T BucketData](buckets *[]*Bucket[A, T]) {
	var excess = len(*buckets) - HoursAgoToKeep
	if excess < 1 {
		return
	}
	// zero out excess buckets to lower their impact until reallocation
	for i := 0; i < excess; i++ {
		(*buckets)[i] = nil
	}
	*buckets = (*buckets)[excess:]
}

func AddData[A BucketDataPt[T], T BucketData, R RecordTypes, D RecordTypesPt[R]](buckets *[]*Bucket[A, T], userData []D) error {
	previousPeriod := time.Time{}
	var newBucket *Bucket[A, T]

	for _, r := range userData {
		recordTime := r.GetTime().UTC()

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentPeriod := recordTime.Truncate(time.Hour)

		// store stats for the period, if we are now on the next period
		if !previousPeriod.IsZero() && currentPeriod.After(previousPeriod) {
			err := AddBin(buckets, newBucket)
			if err != nil {
				return err
			}
			newBucket = nil
		}

		if newBucket == nil {
			offset := -1
			var firstBucketHour time.Time
			var lastBucketHour time.Time

			// pull stats if they already exist
			// we assume the list is fully populated with empty hours for any gaps, so the length should be predictable
			if len(*buckets) > 0 {
				firstBucketHour = (*buckets)[0].Date
				lastBucketHour = (*buckets)[len(*buckets)-1].Date

				// if we need to look for an existing bucket
				if !currentPeriod.After(lastBucketHour) && !currentPeriod.Before(firstBucketHour) {
					offset = int(currentPeriod.Sub(firstBucketHour).Hours())

					if offset < len(*buckets) {
						newBucket = (*buckets)[offset]
						if !newBucket.Date.Equal(currentPeriod) {
							return fmt.Errorf("potentially damaged buckets, offset jump did not find intended record. Found %s, wanted %s", newBucket.Date, currentPeriod)
						}
					}

				}
			}

			// we still don't have a bucket, make a new one.
			if newBucket == nil {
				newBucket = CreateBucket[A](currentPeriod)
			}

			// if on fresh bucket, pull LastRecordTime from previous bucket if possible
			if newBucket.LastRecordTime.IsZero() && len(*buckets) > 0 {
				if offset != -1 && offset+1 < len(*buckets) {
					newBucket.LastRecordTime = (*buckets)[offset-1].LastRecordTime
				} else if !newBucket.Date.Before(firstBucketHour) {
					newBucket.LastRecordTime = (*buckets)[len(*buckets)-1].LastRecordTime
				}
			}
		}

		previousPeriod = currentPeriod

		skipped, err := newBucket.Data.CalculateStats(r, &newBucket.LastRecordTime)
		if err != nil {
			return err
		}
		if !skipped {
			newBucket.LastRecordTime = recordTime
		}
	}

	// store any partial bucket
	if newBucket != nil {
		err := AddBin(buckets, newBucket)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Dates) Reset() {
	*d = Dates{
		OutdatedReason: d.OutdatedReason,
	}
}
