package types

import (
	"context"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
)

const (
	SummaryTypeCGM        = "cgm"
	SummaryTypeBGM        = "bgm"
	SummaryTypeContinuous = "con"
	SchemaVersion         = 6

	lowBloodGlucose         = 3.9
	veryLowBloodGlucose     = 3.0
	highBloodGlucose        = 10.0
	veryHighBloodGlucose    = 13.9
	extremeHighBloodGlucose = 19.4
	HoursAgoToKeep          = 60 * 24
)

var DeviceDataTypesSet = mapset.NewSet[string](continuous.Type, selfmonitored.Type)

var DeviceDataToSummaryTypes = map[string][]string{
	continuous.Type:    {SummaryTypeCGM, SummaryTypeContinuous},
	selfmonitored.Type: {SummaryTypeBGM, SummaryTypeContinuous},
}

var AllSummaryTypes = []string{SummaryTypeCGM, SummaryTypeBGM, SummaryTypeContinuous}

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
	LastUploadDate  time.Time `json:"lastUploadDate,omitempty" bson:"lastUploadDate,omitempty"`

	FirstData time.Time `json:"firstData,omitempty" bson:"firstData,omitempty"`
	LastData  time.Time `json:"lastData,omitempty" bson:"lastData,omitempty"`
}

type CalcState struct {
	Final bool

	FirstCountedDay time.Time
	LastCountedDay  time.Time

	FirstCountedHour time.Time
	LastCountedHour  time.Time

	LastData  time.Time
	FirstData time.Time

	LastRecordDuration int
}

func (d *Dates) Update(status *data.UserDataStatus, firstBucketDate time.Time) {
	d.LastUpdatedDate = status.NextLastUpdated
	d.LastUploadDate = status.LastUpload

	d.FirstData = firstBucketDate
	d.LastData = status.LastData
}

type Periods interface {
	CGMPeriods | BGMPeriods | ContinuousPeriods
}

type PeriodsPt[P Periods, PB BucketDataPt[B], B BucketData] interface {
	*P
	GetType() string
	GetDeviceDataTypes() []string
	Init()
	Update(context.Context, *mongo.Cursor) error
}

type BaseSummary struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type   string             `json:"type" bson:"type"`
	UserID string             `json:"userId" bson:"userId"`
	Config Config             `json:"config" bson:"config"`
	Dates  Dates              `json:"dates" bson:"dates"`
}

type Summary[PP PeriodsPt[P, PB, B], PB BucketDataPt[B], P Periods, B BucketData] struct {
	BaseSummary `bson:",inline"`

	Periods PP `json:"periods" bson:"periods"`
}

type CGMSummary = Summary[*CGMPeriods, *GlucoseBucket, CGMPeriods, GlucoseBucket]
type BGMSummary = Summary[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]
type ContinuousSummary = Summary[*ContinuousPeriods, *ContinuousBucket, ContinuousPeriods, ContinuousBucket]

func NewConfig() Config {
	return Config{
		SchemaVersion:            SchemaVersion,
		HighGlucoseThreshold:     highBloodGlucose,
		VeryHighGlucoseThreshold: veryHighBloodGlucose,
		LowGlucoseThreshold:      lowBloodGlucose,
		VeryLowGlucoseThreshold:  veryLowBloodGlucose,
	}
}

func NewDates() Dates {
	return Dates{
		LastUpdatedDate: time.Time{},
	}
}

func Create[PP PeriodsPt[P, PB, B], PB BucketDataPt[B], P Periods, B BucketData](userId string) *Summary[PP, PB, P, B] {
	s := new(Summary[PP, PB, P, B])
	s.UserID = userId
	s.Periods = new(P)
	s.Periods.Init()
	s.Type = s.Periods.GetType()
	s.Config = NewConfig()
	s.Dates = NewDates()

	return s
}

func GetType[PP PeriodsPt[P, PB, B], PB BucketDataPt[B], P Periods, B BucketData]() string {
	s := new(Summary[PP, PB, P, B])
	return s.Periods.GetType()
}

func GetDeviceDataType[PS PeriodsPt[P, PB, B], PB BucketDataPt[B], P Periods, B BucketData]() []string {
	s := new(Summary[PS, PB, P, B])
	return s.Periods.GetDeviceDataTypes()
}

func (d *Dates) Reset() {
	*d = Dates{}
}
