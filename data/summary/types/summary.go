package types

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const (
	SummaryTypeCGM = "cgm"
	SummaryTypeBGM = "bgm"

	lowBloodGlucose      = 3.9
	veryLowBloodGlucose  = 3.0
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	summaryGlucoseUnits  = "mmol/L"
	hoursAgoToKeep       = 30 * 24
)

type RecordTypes interface {
	//glucoseDatum.Glucose | insulinDatum.Insulin
}

type RecordTypesPt[T any] interface {
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
	// date tracking
	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    time.Time  `json:"lastUploadDate" bson:"lastUploadDate"`
	LastUpdatedDate   time.Time  `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	FirstData         time.Time  `json:"firstData" bson:"firstData"`
	LastData          *time.Time `json:"lastData" bson:"lastData"`
	OutdatedSince     *time.Time `json:"outdatedSince" bson:"outdatedSince"`
}

type Bucket[T any, S BucketDataPt[T]] struct {
	Date           time.Time `json:"date" bson:"date"`
	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`

	Data S
}

type BucketDataPt[T any] interface {
	*T
	CalculateStats(interface{}, *time.Time) error
}

func CreateBucket[T any, A BucketDataPt[T]](t time.Time) *Bucket[T, A] {
	bucket := new(Bucket[T, A])
	bucket.Date = t
	return bucket
}

type Buckets[T any, S BucketDataPt[T]] []Bucket[T, S]

type Stats interface{}

type StatsPt[T any] interface {
	*T
	GetType() string
	Init()
}

type Summary[A Stats] struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string
	UserID string

	Config Config `json:"config" bson:"config"`

	Dates Dates `json:"dates" bson:"dates"`
	Stats A
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

func (s Summary[A]) SetOutdated() {
	s.Dates.OutdatedSince = pointer.FromTime(time.Now().UTC())
}

func NewDates() Dates {
	return Dates{
		HasLastUploadDate: false,
		LastUploadDate:    time.Time{},
		LastUpdatedDate:   time.Time{},
		FirstData:         time.Time{},
		LastData:          nil,
		OutdatedSince:     nil,
	}
}

func Create[T any, A StatsPt[T]](userId string) *Summary[A] {
	s := new(Summary[A])
	s.UserID = userId
	s.Stats.Init()
	s.Type = s.Stats.GetType()
	s.Config = NewConfig()
	s.Dates = NewDates()

	return s
}

func GetTypeString[T any, A StatsPt[T]]() string {
	s := new(Summary[A])
	return s.Stats.GetType()
}

type Period interface {
	BGMPeriod | CGMPeriod
}

func AddBin[T any, A BucketDataPt[T], S Buckets[T, A]](buckets S, newStat Bucket[T, A]) error {
	var hourCount int
	var oldestHour time.Time
	var oldestHourToKeep time.Time
	var existingDay = false
	var statsGap int
	var newStatsTime time.Time

	// update existing hour if one does exist
	if len(buckets) > 0 {
		for i := len(buckets) - 1; i >= 0; i-- {

			if (buckets[i]).Date.Equal(newStat.Date) {
				buckets[i] = newStat
				existingDay = true
				break
			}

			// we already passed our date, give up
			if buckets[i].Date.After(newStat.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(newStat.Date.Sub(buckets[len(buckets)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = newStat.Date.Add(time.Duration(-i) * time.Hour)

			buckets = append(buckets, *CreateBucket[T, A](newStatsTime))
		}
	}

	if existingDay == false {
		buckets = append(buckets, newStat)
	}

	// remove extra days to cap at X days of newStat
	hourCount = len(buckets)
	if hourCount > hoursAgoToKeep {
		buckets = buckets[hourCount-hoursAgoToKeep:]
	}

	// remove any newStat that are older than X days from the last stat
	oldestHour = buckets[0].Date
	oldestHourToKeep = newStat.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(buckets) - 2; i >= 0; i-- {
			if buckets[i].Date.Before(oldestHourToKeep) {
				buckets = buckets[i+1:]
				break
			}
		}
	}

	return nil
}

// CalculateRealMinutes remove partial hour (data end) from total time for more accurate TimeCGMUse
func CalculateRealMinutes(i int, lastRecordTime time.Time) float64 {
	realMinutes := float64(i * 24 * 60)
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	realMinutes = realMinutes - nextHour.Sub(lastRecordTime).Minutes()

	return realMinutes
}

func AddData[T any, A BucketDataPt[T], S Buckets[T, A], R RecordTypes, D RecordTypesPt[R]](s S, userData []D) error {
	var recordTime *time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var newBucket *Bucket[T, A]

	for _, r := range userData {
		recordTime = r.GetTime()
		if err != nil {
			return errors.Wrap(err, "cannot parse time in record")
		}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentHour = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store stats for the day, if we are now on the next hour
		if !lastHour.IsZero() && !currentHour.Equal(lastHour) {
			err = AddBin(s, *newBucket)
			if err != nil {
				return err
			}
			newBucket = nil
		}

		if newBucket == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(s) > 0 {
				for i := len(s) - 1; i >= 0; i-- {
					if s[i].Date.Equal(currentHour) {
						newBucket = &s[i]
						break
					}

					// we already passed our date, give up
					if s[i].Date.After(currentHour) {
						break
					}
				}
			}

			if newBucket == nil {
				newBucket = CreateBucket[T, A](currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if newBucket.LastRecordTime.IsZero() && len(s) > 0 {
			newBucket.LastRecordTime = s[len(s)-1].LastRecordTime
		}

		newBucket.Data.CalculateStats(r, recordTime)

	}

	// store
	err = AddBin(s, *newBucket)
	if err != nil {
		return err
	}

	return nil
}
