package summary

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// reimpliment glucose with only the fields we need, to avoid inheriting Base, which does
// not belong in this collection
type Glucose struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type Stats struct {
	TimeInRange    float64
	TimeBelowRange float64
	TimeAboveRange float64
	TimeCGMUse     float64
	AverageGlucose float64
}

type UserLastUpdated struct {
	LastData   time.Time
	LastUpload time.Time
}

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`

	// date tracking
	LastUpdated *time.Time `json:"lastUpdated" bson:"lastUpdated"`
	FirstData   *time.Time `json:"firstData,omitempty" bson:"firstData,omitempty"`
	LastData    *time.Time `json:"lastData,omitempty" bson:"lastData,omitempty"`
	LastUpload  *time.Time `json:"lastUpload,omitempty" bson:"lastUpload,omitempty"`

	// actual values
	AverageGlucose *Glucose `json:"avgGlucose,omitempty" bson:"avgGlucose,omitempty"`
	TimeInRange    *float64 `json:"timeInRange,omitempty" bson:"timeInRange,omitempty"`
	TimeBelowRange *float64 `json:"timeBelowRange,omitempty" bson:"timeBelowRange,omitempty"`
	TimeAboveRange *float64 `json:"timeAboveRange,omitempty" bson:"timeAboveRange,omitempty"`
	TimeCGMUse     *float64 `json:"timeCGMUse,omitempty" bson:"timeCGMUse,omitempty"`

	// these are mostly just constants right now.
	HighGlucoseThreshold *float64 `json:"highGlucoseThreshold,omitempty" bson:"highGlucoseThreshold,omitempty"`
	LowGlucoseThreshold  *float64 `json:"lowGlucoseThreshold,omitempty" bson:"lowGlucoseThreshold,omitempty"`
}

func New(id string) *Summary {
	return &Summary{
		UserID:      id,
		LastUpdated: &time.Time{},
	}
}
