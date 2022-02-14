package summary

import (
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

const (
	lowBloodGlucose  = 3.9
	highBloodGlucose = 10
	units            = "mmol/l"
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
	DeviceID       string
}

type UserLastUpdated struct {
	LastData   time.Time
	LastUpload time.Time
}

type WeightingResult struct {
	Weight    float64
	StartTime time.Time
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

// assumes all except freestyle is 5 minutes
func GetDuration(dataSet *continuous.Continuous) int64 {
	if dataSet.DeviceID != nil {
		if strings.Contains(*dataSet.DeviceID, "AbbottFreeStyleLibre") {
			return 15
		}
	}
	return 5
}

func CalculateWeight(startTime time.Time, endTime time.Time, lastData time.Time) (*WeightingResult, error) {
	if endTime.Before(lastData) {
		return nil, errors.New("Invalid time period for calculation, endTime before lastData.")
	}

	result := WeightingResult{Weight: 1.0, StartTime: startTime}

	if startTime.Before(lastData) {
		// get ratio between start time and actual start time for weights
		wholeTime := endTime.Sub(startTime)
		newTime := endTime.Sub(lastData)
		result.Weight = newTime.Seconds() / wholeTime.Seconds()

		result.StartTime = lastData
	}

	return &result, nil
}

func CalculateStats(userData []*continuous.Continuous, totalWallMinutes float64) *Stats {
	var inRangeMinutes int64 = 0
	var belowRangeMinutes int64 = 0
	var aboveRangeMinutes int64 = 0
	var totalGlucose float64 = 0
	var totalCGMMinutes int64 = 0

	var normalizedValue float64
	var duration int64

	for _, r := range userData {
		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(units))
		duration = GetDuration(r)

		if normalizedValue <= lowBloodGlucose {
			belowRangeMinutes += duration
		} else if normalizedValue >= highBloodGlucose {
			aboveRangeMinutes += duration
		} else {
			inRangeMinutes += duration
		}

		totalCGMMinutes += duration
		totalGlucose += normalizedValue
	}

	averageGlucose := totalGlucose / float64(len(userData))
	timeInRange := float64(inRangeMinutes) / float64(totalCGMMinutes)
	timeBelowRange := float64(belowRangeMinutes) / float64(totalCGMMinutes)
	timeAboveRange := float64(aboveRangeMinutes) / float64(totalCGMMinutes)
	timeCGMUse := float64(totalCGMMinutes) / totalWallMinutes

	return &Stats{
		TimeInRange:    math.Round(timeInRange*100) / 100,
		TimeBelowRange: math.Round(timeBelowRange*100) / 100,
		TimeAboveRange: math.Round(timeAboveRange*100) / 100,
		TimeCGMUse:     math.Round(timeCGMUse*1000) / 1000,
		AverageGlucose: math.Round(averageGlucose*100) / 100,
	}
}

func ReweightStats(stats *Stats, userSummary *Summary, weight float64) (*Stats, error) {
	if weight < 0 || weight > 1 {
		return stats, errors.New("Invalid weight (<0||>1) for stats")
	}
	// if we are rolling in previous averages
	if weight != 1 && weight >= 0 {
		// check for nil to cover for any new stats that get added after creation
		if userSummary.AverageGlucose.Value != nil {
			stats.AverageGlucose = stats.AverageGlucose*weight + *userSummary.AverageGlucose.Value*(1-weight)
		}

		if userSummary.TimeInRange != nil {
			stats.TimeInRange = stats.TimeInRange*weight + *userSummary.TimeInRange*(1-weight)
		}

		if userSummary.TimeBelowRange != nil {
			stats.TimeBelowRange = stats.TimeBelowRange*weight + *userSummary.TimeBelowRange*(1-weight)
		}

		if userSummary.TimeAboveRange != nil {
			stats.TimeAboveRange = stats.TimeAboveRange*weight + *userSummary.TimeAboveRange*(1-weight)
		}

		if userSummary.TimeCGMUse != nil {
			stats.TimeCGMUse = stats.TimeCGMUse*weight + *userSummary.TimeCGMUse*(1-weight)
		}
	}

	return stats, nil
}
