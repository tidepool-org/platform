package types

import (
	"time"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"

	"go.mongodb.org/mongo-driver/bson/primitive"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	insulinDatum "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
)

const (
	SummaryTypeCGM = "cgm"
	SummaryTypeBGM = "bgm"

	lowBloodGlucose      = 3.9
	veryLowBloodGlucose  = 3.0
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	HoursAgoToKeep       = 60 * 24
	dailyStatsBeakpoint  = 14 * 24

	setOutdatedBuffer = 2 * time.Minute
)

var stopPoints = [...]int{1, 7, 14, 30}

var DeviceDataTypes = [...]string{continuous.Type, selfmonitored.Type}
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

	UpdateWithoutChangeCount int `json:"updateWithoutChangeCount" bson:"updateWithoutChangeCount"`
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
	s.Dates.OutdatedSince = pointer.FromAny(time.Now().Add(setOutdatedBuffer).UTC().Truncate(time.Millisecond))
	s.Dates.HasOutdatedSince = true
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
	// NOTE This is only partially able to handle editing the past, and will break if given a bucket which
	//      must be prepended
	existingHour := false

	// we assume the list is fully populated with empty hours for any gaps, so the length should be predictable
	if len(*buckets) > 0 {
		lastBucketPeriod := (*buckets)[len(*buckets)-1].Date
		currentPeriod := newStat.Date

		hoursConversion := 1
		if lastBucketPeriod.Sub(currentPeriod).Hours() > dailyStatsBeakpoint {
			hoursConversion = 24
		}

		// if we need to look for an existing bucket
		if currentPeriod.Equal(lastBucketPeriod) || currentPeriod.Before(lastBucketPeriod) {

			gapPeriods := int(lastBucketPeriod.Sub(currentPeriod).Hours()) / hoursConversion
			if gapPeriods < len(*buckets) {
				if !(*buckets)[len(*buckets)-gapPeriods-1].Date.Equal(currentPeriod) {
					return errors.New("Potentially damaged buckets, offset jump did not find intended record.")
				}
				(*buckets)[len(*buckets)-gapPeriods-1] = newStat
				existingHour = true
			}
		}

		// add hours for any gaps that this new bucket skipped
		statsGap := int(newStat.Date.Sub((*buckets)[len(*buckets)-1].Date).Hours()) / hoursConversion
		// only add gap buckets if the gap is shorter than max tracking amount
		if statsGap > 0 && statsGap < HoursAgoToKeep {
			gapBuckets := make(Buckets[T, A], 0, statsGap)
			for i := statsGap; i > 1; i-- {
				newStatsTime := newStat.Date.Add(time.Duration(-i+1) * time.Hour * time.Duration(hoursConversion))
				gapBuckets = append(gapBuckets, *CreateBucket[T, A](newStatsTime))
			}

			*buckets = append(*buckets, gapBuckets...)
		} else if statsGap > HoursAgoToKeep {
			// otherwise, the gap is larger than our tracking, delete all the old buckets for a clean state
			*buckets = make(S, 1)
		}
	}

	if existingHour == false {
		*buckets = append(*buckets, newStat)
	}

	// TODO handle dailybuckets
	// remove extra hours to cap at X hours of buckets
	if len(*buckets) > HoursAgoToKeep {
		// zero out any to-be-trimmed buckets to lower their impact until reallocation
		for i := 0; i < len(*buckets)-HoursAgoToKeep; i++ {
			(*buckets)[i] = Bucket[T, A]{}
		}
		*buckets = (*buckets)[len(*buckets)-HoursAgoToKeep:]
	}

	return nil
}

func AddData[T BucketData, A BucketDataPt[T], S Buckets[T, A], R RecordTypes, D RecordTypesPt[R]](buckets *S, userData []D) error {
	lastPeriod := time.Time{}
	var newBucket *Bucket[T, A]
	targetBuckets := buckets

	for _, r := range userData {
		recordTime := r.GetTime()

		recordHour := recordTime.Hour()
		//// reduce accuracy of period to daily if over daily breakpoint
		//// TODO this will jump a bit, we need to hold +1 day to handle it
		//if len(*hourlyBuckets) > 0 && (*hourlyBuckets)[len(*hourlyBuckets)-1].Date.Sub(*recordTime).Hours() > dailyStatsBeakpoint {
		//	recordHour = 0
		//}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentPeriod := time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordHour, 0, 0, 0, recordTime.Location())

		// store stats for the period, if we are now on the next period
		if !lastPeriod.IsZero() && currentPeriod.After(lastPeriod) {
			//if len(*hourlyBuckets) > 0 && (*hourlyBuckets)[len(*hourlyBuckets)-1].Date.Sub(lastPeriod).Hours() > dailyStatsBeakpoint {
			//	targetBuckets = dailyBuckets
			//} else {
			//	targetBuckets = hourlyBuckets
			//}

			err := AddBin(targetBuckets, *newBucket)
			if err != nil {
				return err
			}
			newBucket = nil
		}

		//// repeated from above as we need to switch again after adding
		//if len(*hourlyBuckets) > 0 && (*hourlyBuckets)[len(*hourlyBuckets)-1].Date.Sub(currentPeriod).Hours() > dailyStatsBeakpoint {
		//	targetBuckets = dailyBuckets
		//} else {
		//	targetBuckets = hourlyBuckets
		//}

		if newBucket == nil {
			// pull stats if they already exist
			// we assume the list is fully populated with empty hours for any gaps, so the length should be predictable
			if len(*targetBuckets) > 0 {
				lastBucketHour := (*targetBuckets)[len(*targetBuckets)-1].Date

				// if we need to look for an existing bucket
				if currentPeriod.Equal(lastBucketHour) || currentPeriod.Before(lastBucketHour) {
					hoursConversion := 1
					//if targetBuckets == dailyBuckets {
					//	hoursConversion = 24
					//}

					gap := int(lastBucketHour.Sub(currentPeriod).Hours()) / hoursConversion

					if gap < len(*targetBuckets) {
						newBucket = &(*targetBuckets)[len(*targetBuckets)-gap-1]
						if !newBucket.Date.Equal(currentPeriod) {
							return errors.New("Potentially damaged buckets, offset jump did not find intended record.")
						}
					}
				}
			}

			// we still don't have a bucket, make a new one.
			if newBucket == nil {
				newBucket = CreateBucket[T, A](currentPeriod)
			}
		}

		lastPeriod = currentPeriod

		// if on fresh day, pull LastRecordTime from last day if possible
		if newBucket.LastRecordTime.IsZero() && len(*targetBuckets) > 0 {
			newBucket.LastRecordTime = (*targetBuckets)[len(*targetBuckets)-1].LastRecordTime
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
		err := AddBin(targetBuckets, *newBucket)
		if err != nil {
			return err
		}
	}

	return nil
}
