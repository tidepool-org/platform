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
	lowBloodGlucose      = 3.9
	veryLowBloodGlucose  = 3.0
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	units                = "mmol/l"
)

// reimpliment glucose with only the fields we need, to avoid inheriting Base, which does
// not belong in this collection
type Glucose struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type Stats struct {
	TimeInRange    float64

	TimeBelowRange     float64
	TimeVeryBelowRange float64

	TimeAboveRange     float64
	TimeVeryAboveRange float64

	TimeCGMUse           float64
	AverageGlucose       float64
	GlucoseMgmtIndicator float64

	DeviceID string

	InRangeMinutes        int64

	BelowRangeMinutes     int64
	VeryBelowRangeMinutes int64

	AboveRangeMinutes     int64
	VeryAboveRangeMinutes int64

	TotalGlucose          float64
	TotalCGMMinutes       int64
	TotalRecords          int64
}

func NewStats(deviceId string) *Stats {
	return &Stats{
		DeviceID: deviceId,

		InRangeMinutes:    0,

		BelowRangeMinutes:     0,
		VeryBelowRangeMinutes: 0,

		AboveRangeMinutes:     0,
		VeryAboveRangeMinutes: 0,

		TotalGlucose:      0,
		TotalCGMMinutes:   0,
		TotalRecords:      0,
	}
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
	LastUpdated    *time.Time `json:"lastUpdated" bson:"lastUpdated"`
	FirstData      *time.Time `json:"firstData,omitempty" bson:"firstData,omitempty"`
	LastData       *time.Time `json:"lastData,omitempty" bson:"lastData,omitempty"`
	LastUpload     *time.Time `json:"lastUpload,omitempty" bson:"lastUpload,omitempty"`
	OutdatedSince  *time.Time `json:"outdatedSince,omitempty" bson:"outdatedSince,omitempty"`

	// actual values
	AverageGlucose       *Glucose `json:"avgGlucose,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseMgmtIndicator *float64 `json:"glucoseMgmtIndicator,omitempty" bson:"glucoseMgmtIndicator,omitempty"`
	TimeInRange          *float64 `json:"timeInRange,omitempty" bson:"timeInRange,omitempty"`

	TimeBelowRange     *float64 `json:"timeBelowRange,omitempty" bson:"timeBelowRange,omitempty"`
	TimeVeryBelowRange *float64 `json:"timeVeryBelowRange,omitempty" bson:"timeVeryBelowRange,omitempty"`

	TimeAboveRange     *float64 `json:"timeAboveRange,omitempty" bson:"timeAboveRange,omitempty"`
	TimeVeryAboveRange *float64 `json:"timeVeryAboveRange,omitempty" bson:"timeVeryAboveRange,omitempty"`

	TimeCGMUse         *float64 `json:"timeCGMUse,omitempty" bson:"timeCGMUse,omitempty"`

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

func CalculateGMI(averageGlucose float64) float64 {
	gmi := 12.71 + 4.70587 * averageGlucose
	gmi = (0.09148*gmi) + 2.152
	gmi = math.Round(gmi*100)/100
	return gmi
}

func CalculateStats(userData []*continuous.Continuous, totalWallMinutes float64) *Stats {
	stats := make(map[string]*Stats)
	var winningStats *Stats

	var normalizedValue float64
	var duration int64
	var deviceId string

	for _, r := range userData {
		if r.DeviceID != nil {
			deviceId = *r.DeviceID
		} else {
			deviceId = ""
		}

		if _, ok := stats[deviceId]; !ok {
			stats[deviceId] = NewStats(deviceId)
		}

		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(units))
		duration = GetDuration(r)

		if normalizedValue <= veryLowBloodGlucose {
			stats[deviceId].VeryBelowRangeMinutes += duration
		} else if normalizedValue >= veryHighBloodGlucose {
			stats[deviceId].VeryAboveRangeMinutes += duration
		} else if normalizedValue <= lowBloodGlucose {
			stats[deviceId].BelowRangeMinutes += duration
		} else if normalizedValue >= highBloodGlucose {
			stats[deviceId].AboveRangeMinutes += duration
		} else {
			stats[deviceId].InRangeMinutes += duration
		}

		stats[deviceId].TotalCGMMinutes += duration
		stats[deviceId].TotalGlucose += normalizedValue
		stats[deviceId].TotalRecords += 1
	}

	for deviceId := range stats {
		stats[deviceId].AverageGlucose = stats[deviceId].TotalGlucose / float64(stats[deviceId].TotalRecords)
		stats[deviceId].TimeInRange = float64(stats[deviceId].InRangeMinutes) / float64(stats[deviceId].TotalCGMMinutes)
		stats[deviceId].TimeBelowRange = float64(stats[deviceId].BelowRangeMinutes) / float64(stats[deviceId].TotalCGMMinutes)
		stats[deviceId].TimeVeryBelowRange = float64(stats[deviceId].VeryBelowRangeMinutes) / float64(stats[deviceId].TotalCGMMinutes)
		stats[deviceId].TimeAboveRange = float64(stats[deviceId].AboveRangeMinutes) / float64(stats[deviceId].TotalCGMMinutes)
		stats[deviceId].TimeVeryAboveRange = float64(stats[deviceId].VeryAboveRangeMinutes) / float64(stats[deviceId].TotalCGMMinutes)
		stats[deviceId].TimeCGMUse = float64(stats[deviceId].TotalCGMMinutes) / totalWallMinutes

		stats[deviceId].TimeInRange = math.Round(stats[deviceId].TimeInRange*100) / 100
		stats[deviceId].TimeBelowRange = math.Round(stats[deviceId].TimeBelowRange*100) / 100
		stats[deviceId].TimeVeryBelowRange = math.Round(stats[deviceId].TimeVeryBelowRange*100) / 100
		stats[deviceId].TimeAboveRange = math.Round(stats[deviceId].TimeAboveRange*100) / 100
		stats[deviceId].TimeVeryAboveRange = math.Round(stats[deviceId].TimeVeryAboveRange*100) / 100
		stats[deviceId].TimeCGMUse = math.Round(stats[deviceId].TimeCGMUse*1000) / 1000
		stats[deviceId].GlucoseMgmtIndicator = CalculateGMI(stats[deviceId].AverageGlucose)
		stats[deviceId].AverageGlucose = math.Round(stats[deviceId].AverageGlucose*100) / 100

		if winningStats != nil {
			if stats[deviceId].TimeCGMUse > winningStats.TimeCGMUse {
				winningStats = stats[deviceId]
			}
		} else {
			winningStats = stats[deviceId]
		}
	}

	return winningStats
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

		if userSummary.TimeVeryBelowRange != nil {
			stats.TimeVeryBelowRange = stats.TimeVeryBelowRange*weight + *userSummary.TimeVeryBelowRange*(1-weight)
		}

		if userSummary.TimeAboveRange != nil {
			stats.TimeAboveRange = stats.TimeAboveRange*weight + *userSummary.TimeAboveRange*(1-weight)
		}

		if userSummary.TimeVeryAboveRange != nil {
			stats.TimeVeryAboveRange = stats.TimeVeryAboveRange*weight + *userSummary.TimeVeryAboveRange*(1-weight)
		}

		if userSummary.TimeCGMUse != nil {
			stats.TimeCGMUse = stats.TimeCGMUse*weight + *userSummary.TimeCGMUse*(1-weight)
		}

		if userSummary.GlucoseMgmtIndicator != nil {
			stats.GlucoseMgmtIndicator = stats.GlucoseMgmtIndicator*weight + *userSummary.GlucoseMgmtIndicator*(1-weight)
		}
	}

	return stats, nil
}
