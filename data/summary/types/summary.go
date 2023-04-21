package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	insulinDatum "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
)

const (
	SummaryTypeCGM    = "cgm"
	DeviceDataTypeCGM = "cbg"
	SummaryTypeBGM    = "bgm"
	DeviceDataTypeBGM = "smbg"

	lowBloodGlucose      = 3.9
	veryLowBloodGlucose  = 3.0
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	hoursAgoToKeep       = 30 * 24
)

var stopPoints = [...]int{1, 7, 14, 30}

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
	LastData   time.Time
	LastUpload time.Time
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
	LastUpdatedDate time.Time `json:"lastUpdatedDate" bson:"lastUpdatedDate"`

	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    *time.Time `json:"lastUploadDate" bson:"lastUploadDate"`

	HasFirstData bool       `json:"hasFirstData" bson:"hasFirstData"`
	FirstData    *time.Time `json:"firstData" bson:"firstData"`

	HasLastData bool       `json:"hasLastData" bson:"hasLastData"`
	LastData    *time.Time `json:"lastData" bson:"lastData"`

	HasOutdatedSince bool       `json:"hasOutdatedSince" bson:"hasOutdatedSince"`
	OutdatedSince    *time.Time `json:"outdatedSince" bson:"outdatedSince"`
}

func (d *Dates) ZeroOut() {
	d.LastUpdatedDate = time.Now().UTC()

	d.HasLastUploadDate = false
	d.LastUploadDate = nil

	d.HasFirstData = false
	d.FirstData = nil

	d.HasLastData = false
	d.LastData = nil

	d.HasOutdatedSince = false
	d.OutdatedSince = nil
}

func (d *Dates) Update(status *UserLastUpdated, firstData time.Time) {
	d.LastUpdatedDate = time.Now().UTC()

	d.HasLastUploadDate = true
	d.LastUploadDate = &status.LastUpload

	d.HasFirstData = true
	d.FirstData = &firstData

	d.HasLastData = true
	d.LastData = &status.LastData

	d.HasOutdatedSince = false
	d.OutdatedSince = nil
}

type Bucket[T BucketData, S BucketDataPt[T]] struct {
	Date           time.Time `json:"date" bson:"date"`
	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`

	Data S `json:"data" bson:"data"`
}

type BucketDataPt[T BucketData] interface {
	*T
	CalculateStats(interface{}, *time.Time) (bool, error)
}

func CreateBucket[T BucketData, A BucketDataPt[T]](t time.Time) *Bucket[T, A] {
	bucket := new(Bucket[T, A])
	bucket.Date = t
	bucket.Data = new(T)
	return bucket
}

type Buckets[T BucketData, S BucketDataPt[T]] []Bucket[T, S]

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
	Update(any) error
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
		SchemaVersion:            1,
		HighGlucoseThreshold:     highBloodGlucose,
		VeryHighGlucoseThreshold: veryHighBloodGlucose,
		LowGlucoseThreshold:      lowBloodGlucose,
		VeryLowGlucoseThreshold:  veryLowBloodGlucose,
	}
}

func (s *Summary[T, A]) SetOutdated() {
	s.Dates.OutdatedSince = pointer.FromTime(time.Now().UTC())
}

func NewDates() Dates {
	return Dates{
		LastUpdatedDate: time.Time{},

		HasLastUploadDate: false,
		LastUploadDate:    nil,

		HasFirstData: false,
		FirstData:    nil,

		HasLastData: false,
		LastData:    nil,

		HasOutdatedSince: false,
		OutdatedSince:    nil,
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

func AddBin[T BucketData, A BucketDataPt[T], S Buckets[T, A]](buckets *S, newStat Bucket[T, A]) error {
	var existingHour = false

	// update existing hour if one does exist
	if len(*buckets) > 0 {
		for i := len(*buckets) - 1; i >= 0; i-- {

			if ((*buckets)[i]).Date.Equal(newStat.Date) {
				(*buckets)[i] = newStat
				existingHour = true
				break
			}

			// we already passed our date, give up
			if (*buckets)[i].Date.Before(newStat.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		var statsGap = int(newStat.Date.Sub((*buckets)[len(*buckets)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			var newStatsTime = newStat.Date.Add(time.Duration(-i+1) * time.Hour)

			*buckets = append(*buckets, *CreateBucket[T, A](newStatsTime))
		}
	}

	if existingHour == false {
		*buckets = append(*buckets, newStat)
	}

	// remove extra days to cap at X days of newStat
	var hourCount = len(*buckets)
	if hourCount > hoursAgoToKeep {
		*buckets = (*buckets)[hourCount-hoursAgoToKeep:]
	}

	// remove any newStat that are older than X days from the last stat
	var oldestHour = (*buckets)[0].Date
	var oldestHourToKeep = newStat.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(*buckets) - 2; i >= 0; i-- {
			if (*buckets)[i].Date.Before(oldestHourToKeep) {
				*buckets = (*buckets)[i+1:]
				break
			}
		}
	}

	return nil
}

func AddData[T BucketData, A BucketDataPt[T], S Buckets[T, A], R RecordTypes, D RecordTypesPt[R]](buckets *S, userData []D) error {
	var lastHour time.Time
	var newBucket *Bucket[T, A]

	for _, r := range userData {
		var recordTime = r.GetTime()

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		var currentHour = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store stats for the day, if we are now on the next hour
		if !lastHour.IsZero() && !currentHour.Equal(lastHour) {
			err := AddBin(buckets, *newBucket)
			if err != nil {
				return err
			}
			newBucket = nil
		}

		if newBucket == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			for i := len(*buckets) - 1; i >= 0; i-- {
				if (*buckets)[i].Date.Equal(currentHour) {
					newBucket = &(*buckets)[i]
					break
				}

				// we already passed our date, give up
				if (*buckets)[i].Date.Before(currentHour) {
					break
				}
			}

			// we still don't have a bucket, make a new one.
			if newBucket == nil {
				newBucket = CreateBucket[T, A](currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
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

	// store
	if newBucket != nil {
		err := AddBin(buckets, *newBucket)
		if err != nil {
			return err
		}
	}

	return nil
}
