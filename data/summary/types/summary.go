package types

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/pointer"
)

const (
	SummaryTypeCGM        = "cgm"
	SummaryTypeBGM        = "bgm"
	SummaryTypeContinuous = "con"
	SchemaVersion         = 4

	lowBloodGlucose         = 3.9
	veryLowBloodGlucose     = 3.0
	highBloodGlucose        = 10.0
	veryHighBloodGlucose    = 13.9
	extremeHighBloodGlucose = 19.4
	HoursAgoToKeep          = 60 * 24

	OutdatedReasonUploadCompleted = "UPLOAD_COMPLETED"
	OutdatedReasonDataAdded       = "DATA_ADDED"
	OutdatedReasonSchemaMigration = "SCHEMA_MIGRATION"
)

var DeviceDataTypesSet = mapset.NewSet[string](continuous.Type, selfmonitored.Type)

var DeviceDataToSummaryTypes = map[string][]string{
	continuous.Type:    {SummaryTypeCGM, SummaryTypeContinuous},
	selfmonitored.Type: {SummaryTypeBGM, SummaryTypeContinuous},
}

var AllSummaryTypes = []string{SummaryTypeCGM, SummaryTypeBGM, SummaryTypeContinuous}

type OutdatedSummariesResponse struct {
	UserIds []string  `json:"userIds"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
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
	LastUpdatedReason []string  `json:"lastUpdatedReason,omitempty" bson:"lastUpdatedReason,omitempty"`
	LastUploadDate    time.Time `json:"lastUploadDate,omitempty" bson:"lastUploadDate,omitempty"`

	FirstData time.Time `json:"firstData,omitempty" bson:"firstData,omitempty"`
	LastData  time.Time `json:"lastData,omitempty" bson:"lastData,omitempty"`

	OutdatedSince  *time.Time `json:"outdatedSince,omitempty" bson:"outdatedSince,omitempty"`
	OutdatedReason []string   `json:"outdatedReason,omitempty" bson:"outdatedReason,omitempty"`
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
	d.LastUpdatedReason = d.OutdatedReason
	d.LastUploadDate = status.LastUpload

	d.FirstData = firstBucketDate
	d.LastData = status.LastData

	d.OutdatedSince = nil
	d.OutdatedReason = nil
}

type Stats interface {
	CGMStats | BGMStats | ContinuousStats
}

type StatsPt[S Stats, PB BucketDataPt[B], B BucketData] interface {
	*S
	GetType() string
	GetDeviceDataTypes() []string
	Init()
	Update(context.Context, *mongo.Cursor) error
}

type SummaryShared struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string             `json:"type" bson:"type"`
	UserID string             `json:"userId" bson:"userId"`
	Config Config             `json:"config" bson:"config"`
	Dates  Dates              `json:"dates" bson:"dates"`
}

type Summary[PS StatsPt[S, PB, B], PB BucketDataPt[B], S Stats, B BucketData] struct {
	SummaryShared `json:",inline" bson:",inline"`

	Stats PS `json:"stats" bson:"stats"`
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

func (s *Summary[PS, PB, S, PB]) SetOutdated(reason string) {
	set := mapset.NewSet[string](reason)
	if len(s.Dates.OutdatedReason) > 0 {
		set.Append(s.Dates.OutdatedReason...)
	}

	if reason == OutdatedReasonSchemaMigration {
		*s = *Create[PS, PB](s.UserID)
	}

	s.Dates.OutdatedReason = set.ToSlice()

	if s.Dates.OutdatedSince == nil {
		s.Dates.OutdatedSince = pointer.FromAny(time.Now().Truncate(time.Millisecond).UTC())
	}
}

func (s *Summary[PS, PB, S, B]) SetNotOutdated() {
	s.Dates.OutdatedReason = nil
	s.Dates.OutdatedSince = nil
}

func NewDates() Dates {
	return Dates{
		LastUpdatedDate: time.Time{},
	}
}

func Create[PS StatsPt[S, PB, B], PB BucketDataPt[B], S Stats, B BucketData](userId string) *Summary[PS, PB, S, B] {
	s := new(Summary[PS, PB, S, B])
	s.UserID = userId
	s.Stats = new(S)
	s.Stats.Init()
	s.Type = s.Stats.GetType()
	s.Config = NewConfig()
	s.Dates = NewDates()

	return s
}

func GetType[PS StatsPt[S, PB, B], PB BucketDataPt[B], S Stats, B BucketData]() string {
	s := new(Summary[PS, PB, S, B])
	return s.Stats.GetType()
}

func GetDeviceDataType[PS StatsPt[S, PB, B], PB BucketDataPt[B], S Stats, B BucketData]() []string {
	s := new(Summary[PS, PB, S, B])
	return s.Stats.GetDeviceDataTypes()
}

func (d *Dates) Reset() {
	*d = Dates{
		OutdatedReason: d.OutdatedReason,
	}
}
