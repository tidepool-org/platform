package summary

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/errors"
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
	CGM *UserCGMLastUpdated
	BGM *UserBGMLastUpdated
}

type OutdatedUserIDs struct {
	CGM []string
	BGM []string
}

type UserData struct {
	CGM []*glucoseDatum.Glucose
	BGM []*glucoseDatum.Glucose
}

type TypeOutdatedTimes struct {
	CGM *time.Time
	BGM *time.Time
}

type Period struct {
	HasAverageGlucose             bool `json:"hasAverageGlucose" bson:"hasAverageGlucose"`
	HasGlucoseManagementIndicator bool `json:"hasGlucoseManagementIndicator" bson:"hasGlucoseManagementIndicator"`
	HasTimeCGMUsePercent          bool `json:"hasTimeCGMUsePercent" bson:"hasTimeCGMUsePercent"`
	HasTimeInTargetPercent        bool `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	HasTimeInHighPercent          bool `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	HasTimeInVeryHighPercent      bool `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	HasTimeInLowPercent           bool `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	HasTimeInVeryLowPercent       bool `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`

	// actual values
	TimeCGMUsePercent *float64 `json:"timeCGMUsePercent" bson:"timeCGMUsePercent"`
	TimeCGMUseMinutes int      `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`
	TimeCGMUseRecords int      `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`

	AverageGlucose             *Glucose `json:"averageGlucose" bson:"avgGlucose"`
	GlucoseManagementIndicator *float64 `json:"glucoseManagementIndicator" bson:"glucoseManagementIndicator"`

	TimeInTargetPercent *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`
	TimeInTargetMinutes int      `json:"timeInTargetMinutes" bson:"timeInTargetMinutes"`
	TimeInTargetRecords int      `json:"timeInTargetRecords" bson:"timeInTargetRecords"`

	TimeInLowPercent *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`
	TimeInLowMinutes int      `json:"timeInLowMinutes" bson:"timeInLowMinutes"`
	TimeInLowRecords int      `json:"timeInLowRecords" bson:"timeInLowRecords"`

	TimeInVeryLowPercent *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`
	TimeInVeryLowMinutes int      `json:"timeInVeryLowMinutes" bson:"timeInVeryLowMinutes"`
	TimeInVeryLowRecords int      `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`

	TimeInHighPercent *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`
	TimeInHighMinutes int      `json:"timeInHighMinutes" bson:"timeInHighMinutes"`
	TimeInHighRecords int      `json:"timeInHighRecords" bson:"timeInHighRecords"`

	TimeInVeryHighPercent *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`
	TimeInVeryHighMinutes int      `json:"timeInVeryHighMinutes" bson:"timeInVeryHighMinutes"`
	TimeInVeryHighRecords int      `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
}

type Config struct {
	SchemaVersion string `json:"schemaVersion" bson:"schemaVersion"`

	// these are just constants right now.
	HighGlucoseThreshold     float64 `json:"highGlucoseThreshold" bson:"highGlucoseThreshold"`
	VeryHighGlucoseThreshold float64 `json:"veryHighGlucoseThreshold" bson:"veryHighGlucoseThreshold"`
	LowGlucoseThreshold      float64 `json:"lowGlucoseThreshold" bson:"lowGlucoseThreshold"`
	VeryLowGlucoseThreshold  float64 `json:"VeryLowGlucoseThreshold" bson:"VeryLowGlucoseThreshold"`
}

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`

	CGM CGMSummary `json:"cgmSummary" bson:"cgmSummary"`
	BGM BGMSummary `json:"bgmSummary" bson:"bgmSummary"`

	Periods map[string]*Period `json:"periods" bson:"periods"`

	Config Config `json:"config" bson:"config"`
}

func New(id string) *Summary {
	return &Summary{
		UserID: id,

		CGM: CGMSummary{
			Periods:           make(map[string]*CGMPeriod),
			HourlyStats:       make([]*CGMStats, 0),
			TotalHours:        0,
			HasLastUploadDate: false,
			LastUploadDate:    time.Time{},
			LastUpdatedDate:   time.Time{},
			FirstData:         time.Time{},
			LastData:          nil,
			OutdatedSince:     nil,
		},

		BGM: BGMSummary{
			Periods:           make(map[string]*BGMPeriod),
			HourlyStats:       make([]*BGMStats, 0),
			TotalHours:        0,
			HasLastUploadDate: false,
			LastUploadDate:    time.Time{},
			LastUpdatedDate:   time.Time{},
			FirstData:         time.Time{},
			LastData:          nil,
			OutdatedSince:     nil,
		},

		Config: Config{
			SchemaVersion:            "1",
			HighGlucoseThreshold:     highBloodGlucose,
			VeryHighGlucoseThreshold: veryHighBloodGlucose,
			LowGlucoseThreshold:      lowBloodGlucose,
			VeryLowGlucoseThreshold:  veryLowBloodGlucose,
		},
	}
}

func (userSummary *Summary) Update(ctx context.Context, status *UserLastUpdated, userData *UserData) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if userData == nil {
		return errors.New("userData is missing")
	}

	userSummary.UpdateCGM(ctx, status.CGM, userData.CGM)
	userSummary.UpdateBGM(ctx, status.BGM, userData.BGM)

	return nil
}
