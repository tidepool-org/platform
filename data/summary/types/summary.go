package types

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

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

	setOutdatedBuffer = 2 * time.Minute
	setOutdatedLimit  = 30 * time.Minute

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

// Glucose reimplementation with only the fields we need, to avoid inheriting Base, which does
// not belong in this collection
type Glucose struct {
	Units string  `json:"units" bson:"units"`
	Value float64 `json:"value" bson:"value"`
}

type UserLastUpdated struct {
	FirstData       time.Time
	LastData        time.Time
	LastUpload      time.Time
	NextLastUpdated time.Time
	LastUpdated     time.Time
}

type ModifiedPeriod struct {
	Bucket       time.Time `bson:"_id"`
	EarliestTime time.Time `bson:"minTime"`
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

	HasOutdatedSince   bool       `json:"hasOutdatedSince" bson:"hasOutdatedSince"`
	OutdatedSince      *time.Time `json:"outdatedSince" bson:"outdatedSince"`
	OutdatedSinceLimit *time.Time `json:"outdatedSinceLimit" bson:"outdatedSinceLimit"`
	OutdatedReason     []string   `json:"outdatedReason" bson:"outdatedReason"`
}

func (d *Dates) Update(status *UserLastUpdated, firstData time.Time) {
	d.LastUpdatedDate = status.NextLastUpdated
	d.LastUpdatedReason = d.OutdatedReason

	d.HasLastUploadDate = true
	d.LastUploadDate = &status.LastUpload

	d.HasFirstData = true
	d.FirstData = &firstData

	d.HasLastData = true
	d.LastData = &status.LastData

	d.HasOutdatedSince = false
	d.OutdatedSince = nil
	d.OutdatedSinceLimit = nil
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

type Buckets[T BucketData, S BucketDataPt[T]] []*Bucket[S, T]

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
	Update(context.Context, *mongo.Cursor) error
}

type Summary[T Stats, A StatsPt[T]] struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string             `json:"type" bson:"type"`
	UserID string             `json:"userId" bson:"userId"`

	Config Config `json:"config" bson:"config"`

	Dates Dates `json:"dates" bson:"dates"`
	Stats A     `json:"stats" bson:"stats"`

	UpdateWithoutChangeCount int `json:"updateWithoutChangeCount" bson:"updateWithoutChangeCount"`
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
		*s = *Create[T, A](s.UserID)
	}

	s.Dates.OutdatedReason = set.ToSlice()

	timestamp := time.Now().Truncate(time.Millisecond).UTC()
	if s.Dates.OutdatedSinceLimit == nil {
		newOutdatedSinceLimit := timestamp.Add(setOutdatedLimit)
		s.Dates.OutdatedSinceLimit = &newOutdatedSinceLimit
	}

	if s.Dates.OutdatedSince == nil || s.Dates.OutdatedSince.Before(*s.Dates.OutdatedSinceLimit) {
		newOutdatedSince := timestamp.Add(setOutdatedBuffer)
		s.Dates.OutdatedSince = &newOutdatedSince
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

		HasOutdatedSince:   false,
		OutdatedSince:      nil,
		OutdatedSinceLimit: nil,
		OutdatedReason:     nil,
	}
}

func Create[T Stats, A StatsPt[T]](userId string) *Summary[T, A] {
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

func AddBin[T BucketData, A BucketDataPt[T], S Buckets[T, A]](buckets *S, newStat *Bucket[A, T]) error {
	// NOTE This is only partially able to handle editing the past, and will break if given a bucket which
	//      must be prepended
	if len(*buckets) == 0 {
		*buckets = append(*buckets, newStat)
	}

	lastBucketPeriod := (*buckets)[len(*buckets)-1].Date
	firstBucketPeriod := (*buckets)[0].Date
	newPeriod := newStat.Date
	statsGap := 0
	var gapStart time.Time
	var gapEnd time.Time

	// get the gap range
	if newPeriod.After(lastBucketPeriod) {
		// check if we are so far ahead that we should just nuke it
		if newPeriod.Add(-HoursAgoToKeep * time.Hour).Before(lastBucketPeriod) {
			*buckets = make(S, 0, 1)
		} else {
			gapStart = newPeriod.Add(time.Duration(-statsGap+1) * time.Hour)
			gapEnd = newPeriod
			statsGap = int(newPeriod.Sub(lastBucketPeriod).Hours())

			// remove extra hours to cap at X hours of buckets
			if len(*buckets) > HoursAgoToKeep {
				// zero out any to-be-trimmed buckets to lower their impact until reallocation
				for i := 0; i < len(*buckets)-HoursAgoToKeep; i++ {
					(*buckets)[i] = nil
				}
				*buckets = (*buckets)[len(*buckets)-HoursAgoToKeep:]
			}
		}
	} else if newPeriod.Before(firstBucketPeriod) {
		// check if we are so far behind that we are irrelevant
		if lastBucketPeriod.Add(-HoursAgoToKeep * time.Hour).Before(newPeriod) {
			return errors.New("given irrelevant bucket, too old")
		} else {
			gapStart = newPeriod.Add(time.Hour)
			gapEnd = newPeriod.Add(time.Duration(statsGap) * time.Hour)
			statsGap = int(firstBucketPeriod.Sub(newPeriod).Hours())
		}
	} else { // we are replacing an existing bucket
		// check if bucket being replaced has the right date
		offset := int(newPeriod.Sub(firstBucketPeriod).Hours())
		if !(*buckets)[offset].Date.Equal(newPeriod) {
			return errors.New("Potentially damaged buckets, offset jump did not find intended record when replacing bucket.")
		}
		(*buckets)[offset] = newStat
		return nil
	}

	// TODO move to CreateEmptyBucketsBetween
	gapBuckets := make(S, 0, Abs(statsGap))
	for i := gapStart; i.Before(gapEnd); i = i.Add(time.Hour) {
		gapBuckets = append(gapBuckets, CreateBucket[A](i))
	}

	if newPeriod.After(lastBucketPeriod) {
		*buckets = append(*buckets, newStat)
	} else if newPeriod.Before(firstBucketPeriod) {
		*buckets = append([]*Bucket[A, T]{newStat}, *buckets...)
	} else {
		return errors.New("eh? bucket not before or after, but not existing?")
	}

	return nil
}

func AddData[D RecordTypesPt[R], A BucketDataPt[T], T BucketData, R RecordTypes](ctx context.Context, buckets *Buckets[T, A], userData *mongo.Cursor) error {
	var r D
	var newBucket *Bucket[A, T]
	lastPeriod := time.Time{}

	for userData.Next(ctx) {
		if err := userData.Decode(r); err != nil {
			return errors.New("Unable to decode userData")
		}
		recordTime := r.GetTime()

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentPeriod := time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store stats for the period, if we are now on a different period
		if !lastPeriod.IsZero() && currentPeriod.After(lastPeriod) {
			err := AddBin(buckets, newBucket)
			if err != nil {
				return err
			}
			newBucket = nil
		}

		if newBucket == nil {
			// pull stats if they already exist
			// we assume the list is fully populated with empty hours for any gaps, so the length should be predictable
			if len(*buckets) > 0 {
				lastBucketHour := (*buckets)[len(*buckets)-1].Date

				// if we need to look for an existing bucket
				if currentPeriod.Equal(lastBucketHour) || currentPeriod.Before(lastBucketHour) {
					fmt.Println("getting existing bucket")
					gap := int(lastBucketHour.Sub(currentPeriod).Hours())

					fmt.Println("going back", gap, "buckets, have", len(*buckets), "buckets")
					if gap < len(*buckets) {
						newBucket = (*buckets)[len(*buckets)-gap-1]
						fmt.Println(newBucket.Date, "!=", currentPeriod)
						if !newBucket.Date.Equal(currentPeriod) {
							return errors.New("Potentially damaged buckets, offset jump did not find intended record when adding data.")
						}
					}
				}
			}

			// we still don't have a bucket, make a new one.
			if newBucket == nil {
				newBucket = CreateBucket[A](currentPeriod)
			}
		}

		lastPeriod = currentPeriod

		// if on fresh hour, pull LastRecordTime from last day if possible
		if newBucket.LastRecordTime.IsZero() && len(*buckets) > 0 {
			newBucket.LastRecordTime = (*buckets)[len(*buckets)-1].LastRecordTime
		}

		skipped, err := newBucket.Data.CalculateStats(r, &newBucket.LastRecordTime)
		if err != nil {
			return err
		}
		if !skipped {
			newBucket.LastRecordTime = *recordTime
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

func SetStartTime[T Stats, A StatsPt[T]](userSummary *Summary[T, A], status *UserLastUpdated) {
	// remove HoursAgoToKeep/24 days for start time
	status.FirstData = status.LastData.AddDate(0, 0, -HoursAgoToKeep/24)
	status.LastUpdated = userSummary.Dates.LastUpdatedDate

	//if userSummary.Dates.LastData != nil {
	//	// if summary already exists with a last data checkpoint, start data pull there
	//	if startTime.Before(*userSummary.Dates.LastData) {
	//		startTime = *userSummary.Dates.LastData
	//	}
	//
	//	// ensure LastData does not move backwards by capping it at summary LastData
	//	if status.LastData.Before(*userSummary.Dates.LastData) {
	//		status.LastData = *userSummary.Dates.LastData
	//	}
	//}
	//
	//return startTime
}

func (d *Dates) Reset() {
	*d = Dates{
		OutdatedReason: d.OutdatedReason,
	}
}
