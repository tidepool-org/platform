package summary

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/tidepool-org/platform/log"

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
	TimeInRange float64

	TimeBelowRange     float64
	TimeVeryBelowRange float64

	TimeAboveRange     float64
	TimeVeryAboveRange float64

	TimeCGMUse           float64
	AverageGlucose       float64
	GlucoseMgmtIndicator float64

	DeviceID string

	InRangeMinutes int64

	BelowRangeMinutes     int64
	VeryBelowRangeMinutes int64

	AboveRangeMinutes     int64
	VeryAboveRangeMinutes int64

	TotalGlucose    float64
	TotalCGMMinutes int64
	TotalRecords    int64
}

func NewStats(deviceId string) *Stats {
	return &Stats{
		DeviceID: deviceId,

		InRangeMinutes: 0,

		BelowRangeMinutes:     0,
		VeryBelowRangeMinutes: 0,

		AboveRangeMinutes:     0,
		VeryAboveRangeMinutes: 0,

		TotalGlucose:    0,
		TotalCGMMinutes: 0,
		TotalRecords:    0,
	}
}

type UserLastUpdated struct {
	LastData   time.Time
	LastUpload time.Time
}

type WeightingInput struct {
	StartTime        time.Time
	EndTime          time.Time
	LastData         time.Time
	OldPercentCGMUse float64
	NewPercentCGMUse float64
}

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`

	// date tracking
	LastUpdated   *time.Time `json:"lastUpdated,omitempty" bson:"lastUpdated,omitempty"`
	FirstData     *time.Time `json:"firstData,omitempty" bson:"firstData,omitempty"`
	LastData      *time.Time `json:"lastData,omitempty" bson:"lastData,omitempty"`
	LastUpload    *time.Time `json:"lastUpload,omitempty" bson:"lastUpload,omitempty"`
	OutdatedSince *time.Time `json:"outdatedSince" bson:"outdatedSince"`

	// actual values
	AverageGlucose       *Glucose `json:"avgGlucose,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseMgmtIndicator *float64 `json:"glucoseMgmtIndicator,omitempty" bson:"glucoseMgmtIndicator,omitempty"`
	TimeInRange          *float64 `json:"timeInRange,omitempty" bson:"timeInRange,omitempty"`

	TimeBelowRange     *float64 `json:"timeBelowRange,omitempty" bson:"timeBelowRange,omitempty"`
	TimeVeryBelowRange *float64 `json:"timeVeryBelowRange,omitempty" bson:"timeVeryBelowRange,omitempty"`

	TimeAboveRange     *float64 `json:"timeAboveRange,omitempty" bson:"timeAboveRange,omitempty"`
	TimeVeryAboveRange *float64 `json:"timeVeryAboveRange,omitempty" bson:"timeVeryAboveRange,omitempty"`

	TimeCGMUse *float64 `json:"timeCGMUse,omitempty" bson:"timeCGMUse,omitempty"`

	// these are mostly just constants right now.
	HighGlucoseThreshold     *float64 `json:"highGlucoseThreshold,omitempty" bson:"highGlucoseThreshold,omitempty"`
	VeryHighGlucoseThreshold *float64 `json:"veryHighGlucoseThreshold,omitempty" bson:"veryHighGlucoseThreshold,omitempty"`
	LowGlucoseThreshold      *float64 `json:"lowGlucoseThreshold,omitempty" bson:"lowGlucoseThreshold,omitempty"`
	VeryLowGlucoseThreshold  *float64 `json:"VeryLowGlucoseThreshold,omitempty" bson:"VeryLowGlucoseThreshold,omitempty"`
}

func New(id string) *Summary {
	return &Summary{
		UserID:        id,
		OutdatedSince: &time.Time{},
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

func CalculateWeight(input *WeightingInput) (*float64, error) {
	if input.EndTime.Before(input.LastData) {
		return nil, errors.New("Invalid time period for calculation, endTime before lastData.")
	}

	var weight float64 = 1.0

	if input.StartTime.Before(input.LastData) {
		// get ratio between start time and actual start time for weights
		wholeTime := input.EndTime.Sub(input.StartTime)
		newTime := input.EndTime.Sub(input.LastData)
		weight = newTime.Seconds() / wholeTime.Seconds()
	}

	// adjust weight for %cgm use
	oldWeight := (1 - weight) * input.OldPercentCGMUse
	newWeight := weight * input.NewPercentCGMUse
	weightMultiplier := 1 / (oldWeight + newWeight)
	weight = newWeight * weightMultiplier

	return &weight, nil
}

func CalculateGMI(averageGlucose float64) float64 {
	gmi := 12.71 + 4.70587*averageGlucose
	gmi = (0.09148 * gmi) + 2.152
	gmi = math.Round(gmi*10) / 10
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
	}

	return stats, nil
}

func Update(ctx context.Context, userSummary *Summary, status *UserLastUpdated, userData []*continuous.Continuous) (*Summary, error) {
	var err error
	logger := log.LoggerFromContext(ctx)

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if userSummary == nil {
		return nil, errors.New("userSummary is missing")
	}

	if len(userData) == 0 {
		return nil, errors.New("userData is empty")
	}

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	userSummary.LastUpdated = &timestamp
	userSummary.OutdatedSince = nil

	// remove 2 weeks for start time
	startTime := status.LastData.AddDate(0, 0, -14)
	endTime := status.LastData

	// hold onto 2 week past date for summary weighting time range
	firstData := startTime

	// if summary already exists with a last data checkpoint, use it for the start of this calculation
	if userSummary.LastData != nil {
		if startTime.Before(*userSummary.LastData) {
			startTime = *userSummary.LastData
			logger.Debugf(
				"Found existing summary for userid %v, adjusting startTime for rolling calculation.",
				userSummary.UserID)
		}
	}

	oldestRecord, err := time.Parse(time.RFC3339Nano, *userData[0].Time)
	if err != nil {
		return nil, err
	}

	newestRecord, err := time.Parse(time.RFC3339Nano, *userData[len(userData)-1].Time)
	if err != nil {
		return nil, err
	}

	// check that the oldest record we are given, fits within the range we expect
	if oldestRecord.Before(startTime) || newestRecord.After(endTime) {
		return nil, errors.New("Received data for summary calculation does not match given start and end points")
	}

	totalMinutes := status.LastData.Sub(startTime).Minutes()
	logger.Debugf("Total minutes for userid %v summary calculation: %v", userSummary.UserID, totalMinutes)

	// don't recalculate if there is no new data/this was double called
	if totalMinutes < 1 {
		logger.Debugf("Total minutes near-zero for userid %v summary calculation, aborting.", userSummary.UserID)
		return userSummary, nil
	}

	stats := CalculateStats(userData, totalMinutes)
	logger.Debugf("Stats for new data for userid %v summary: %+v", userSummary.UserID, stats)

	var newWeight = pointer.FromFloat64(1.0)
	if userSummary.LastData != nil && userSummary.TimeCGMUse != nil {
		logger.Debugf("Calculating rolling weight for userid %v.", userSummary.UserID)
		weightingInput := WeightingInput{
			StartTime:        firstData,
			EndTime:          endTime,
			LastData:         *userSummary.LastData,
			OldPercentCGMUse: *userSummary.TimeCGMUse,
			NewPercentCGMUse: stats.TimeCGMUse,
		}

		newWeight, err = CalculateWeight(&weightingInput)
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf("Weight for userid %v new summary: %v", userSummary.UserID, newWeight)
	//logger.Debugf("Existing summary for userid %v: %v", userSummary.UserID, userSummary)
	//logger.Debugf("New stats for userid %v: %v", userSummary.UserID, stats)

	stats, err = ReweightStats(stats, userSummary, *newWeight)
	if err != nil {
		return nil, err
	}
	//logger.Debugf("New stats for userid %v after reweight: %v", userSummary.UserID, stats)

	userSummary.LastUpload = &status.LastUpload
	userSummary.LastData = &status.LastData
	userSummary.FirstData = &firstData
	userSummary.TimeInRange = pointer.FromFloat64(stats.TimeInRange)
	userSummary.TimeBelowRange = pointer.FromFloat64(stats.TimeBelowRange)
	userSummary.TimeVeryBelowRange = pointer.FromFloat64(stats.TimeVeryBelowRange)
	userSummary.TimeAboveRange = pointer.FromFloat64(stats.TimeAboveRange)
	userSummary.TimeVeryAboveRange = pointer.FromFloat64(stats.TimeVeryAboveRange)
	userSummary.TimeCGMUse = pointer.FromFloat64(stats.TimeCGMUse)
	userSummary.GlucoseMgmtIndicator = pointer.FromFloat64(stats.GlucoseMgmtIndicator)
	userSummary.AverageGlucose = &Glucose{
		Value: pointer.FromFloat64(stats.AverageGlucose),
		Units: pointer.FromString(units),
	}
	userSummary.LowGlucoseThreshold = pointer.FromFloat64(lowBloodGlucose)
	userSummary.VeryLowGlucoseThreshold = pointer.FromFloat64(veryLowBloodGlucose)
	userSummary.HighGlucoseThreshold = pointer.FromFloat64(highBloodGlucose)
	userSummary.VeryHighGlucoseThreshold = pointer.FromFloat64(veryHighBloodGlucose)

	// we only add GMI if cgm use >70%, otherwise clear it
	if *userSummary.TimeCGMUse > 0.7 {
		userSummary.GlucoseMgmtIndicator = pointer.FromFloat64(CalculateGMI(stats.AverageGlucose))
	} else {
		userSummary.GlucoseMgmtIndicator = nil
	}

	//logger.Debugf("Final summary for userid %v: %v", userSummary.UserID, userSummary)
	return userSummary, nil
}
