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

type BucketData interface {
	BGMBucket | CGMBucket
}

type BucketDataPt[T any] interface {
	*T
	BucketData
}

func CreateHourlyBucket[T BucketData, S BucketDataPt[T]](t time.Time) S {
	stat := new(T)
	return S(stat)
	//return stat
}

type Bucket[T BucketData, S BucketDataPt[T]] struct {
	Date           time.Time `json:"date" bson:"date"`
	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`

	Data S
}

type HourlyBuckets[T BucketData, S BucketDataPt[T]] struct {
	Buckets []Bucket[T, S]
}

type Stats interface {
	BGMStats | CGMStats
	GetType() string
	Init()
	CalculateSummary()
	GetHourlyStats() interface{}
}

type Summary[T Stats] struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string
	UserID string

	Config Config `json:"config" bson:"config"`

	Dates Dates `json:"dates" bson:"dates"`
	Stats T
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

func (s Summary[T]) SetOutdated() {
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

func Update[T Stats, R RecordTypes](s T, userData []*R) error {
	err := AddData(s.GetHourlyStats(), userData)
	if err != nil {
		return err
	}
	s.CalculateSummary()

	return nil
}

func Create[T Stats](userId string) Summary[T] {
	stats := new(T)
	(*stats).Init()
	s := Summary[T]{
		Type:   (*stats).GetType(),
		UserID: userId,
		Stats:  *stats,
		Config: NewConfig(),
		Dates:  NewDates(),
	}

	return s
}

func GetTypeString[T Stats]() string {
	t := new(T)
	return (*t).GetType()
}

type Period interface {
	BGMPeriod | CGMPeriod
}

func AddStats[T HourlyBucket](stats []T, newStat T) error {
	var hourCount int
	var oldestHour time.Time
	var oldestHourToKeep time.Time
	var existingDay = false
	var statsGap int
	var newStatsTime time.Time

	// update existing hour if one does exist
	if len(stats) > 0 {
		for i := len(stats) - 1; i >= 0; i-- {

			if (stats[i]).GetDate().Equal(newStat.GetDate()) {
				stats[i] = newStat
				existingDay = true
				break
			}

			// we already passed our date, give up
			if stats[i].GetDate().After(newStat.GetDate()) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(newStat.GetDate().Sub(stats[len(stats)-1].GetDate()).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = newStat.GetDate().Add(time.Duration(-i) * time.Hour)

			stats = append(stats, *CreateHourlyStat[T](newStatsTime))
		}
	}

	if existingDay == false {
		stats = append(stats, newStat)
	}

	// remove extra days to cap at X days of newStat
	hourCount = len(stats)
	if hourCount > hoursAgoToKeep {
		stats = stats[hourCount-hoursAgoToKeep:]
	}

	// remove any newStat that are older than X days from the last stat
	oldestHour = stats[0].GetDate()
	oldestHourToKeep = newStat.GetDate().Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(stats) - 2; i >= 0; i-- {
			if stats[i].GetDate().Before(oldestHourToKeep) {
				stats = stats[i+1:]
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

func AddData[T any, S HourlyStatPt[T]](s []S, userData []RecordTypes) error {
	var recordTime *time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var newStat S

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
			err = AddStats(s, newStat)
			if err != nil {
				return err
			}
			newStat = nil
		}

		if newStat == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(s) > 0 {
				for i := len(s) - 1; i >= 0; i-- {
					if s[i].GetDate().Equal(currentHour) {
						newStat = s[i]
						break
					}

					// we already passed our date, give up
					if s[i].GetDate().After(currentHour) {
						break
					}
				}
			}

			if newStat == nil {
				newStat = CreateHourlyStat[T](currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if newStat.GetLastRecordTime().IsZero() && len(s) > 0 {
			newStat.SetLastRecordTime(s[len(s)-1].GetLastRecordTime())
		}

		newStat.CalculateStats(r)
	}

	// store
	err = AddStats(s, newStat)
	if err != nil {
		return err
	}

	return nil
}
