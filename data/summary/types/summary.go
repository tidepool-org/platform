package types

import (
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
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

func NewConfig() Config {
	return Config{
		SchemaVersion:            1,
		HighGlucoseThreshold:     highBloodGlucose,
		VeryHighGlucoseThreshold: veryHighBloodGlucose,
		LowGlucoseThreshold:      lowBloodGlucose,
		VeryLowGlucoseThreshold:  veryLowBloodGlucose,
	}
}

type Summary[T Stats] struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Type   string
	UserID string

	Config Config `json:"config" bson:"config"`

	Dates Dates `json:"dates" bson:"dates"`
	Stats T
}

func (s Summary[T]) SetOutdated() {
	s.Dates.OutdatedSince = pointer.FromTime(time.Now().UTC())
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

type Stats interface {
	BGMStats | CGMStats
	GetType() string
	PopulateStats()
}

func Create[T Stats](userId string) Summary[T] {
	stats := new(T)
	s := Summary[T]{
		Type:   (*stats).GetType(),
		UserID: userId,
		Stats:  *stats,
		Config: NewConfig(),
		Dates:  NewDates(),
	}

	s.Stats.PopulateStats()
	return s
}

func GetTypeString[T Stats]() string {
	t := new(T)
	return (*t).GetType()
}

func SkipUntil(date time.Time, userData []*glucoseDatum.Glucose) ([]*glucoseDatum.Glucose, error) {
	var skip int
	for i := 0; i < len(userData); i++ {
		recordTime := *userData[i].Time

		if recordTime.Before(date) {
			skip = i + 1
		} else {
			break
		}
	}

	if skip > 0 {
		userData = userData[skip:]
	}

	return userData, nil
}
