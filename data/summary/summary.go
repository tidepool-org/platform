package summary

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
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

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`
	Type   string             `json:"type" bson:"type"`

	// date tracking
	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    time.Time  `json:"lastUploadDate" bson:"lastUploadDate"`
	LastUpdatedDate   time.Time  `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	FirstData         time.Time  `json:"firstData" bson:"firstData"`
	LastData          *time.Time `json:"lastData" bson:"lastData"`
	OutdatedSince     *time.Time `json:"outdatedSince" bson:"outdatedSince"`

	Config Config `json:"config" bson:"config"`
}

func NewSummary(id string, typ string) *Summary {
	return &Summary{
		UserID: id,
		Type:   typ,

		HasLastUploadDate: false,
		LastUploadDate:    time.Time{},
		LastUpdatedDate:   time.Time{},
		FirstData:         time.Time{},
		LastData:          nil,
		OutdatedSince:     nil,

		Config: Config{
			SchemaVersion:            1,
			HighGlucoseThreshold:     highBloodGlucose,
			VeryHighGlucoseThreshold: veryHighBloodGlucose,
			LowGlucoseThreshold:      lowBloodGlucose,
			VeryLowGlucoseThreshold:  veryLowBloodGlucose,
		},
	}
}
