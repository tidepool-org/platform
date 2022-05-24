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
	summaryGlucoseUnits  = "mmol/L"
	daysAgoToKeep        = 14
)

// Glucose reimplementation with only the fields we need, to avoid inheriting Base, which does
// not belong in this collection
type Glucose struct {
	Units *string  `json:"units," bson:"units"`
	Value *float64 `json:"value" bson:"value"`
}

type Stats struct {
	DeviceID string    `json:"deviceId" bson:"deviceId"`
	Date     time.Time `json:"date" bson:"date"`

	InRangeMinutes int64 `json:"inRangeMinutes" bson:"inRangeMinutes"`
	InRangeRecords int64 `json:"inRangeRecords" bson:"inRangeRecords"`

	BelowRangeMinutes int64 `json:"belowRangeMinutes" bson:"belowRangeMinutes"`
	BelowRangeRecords int64 `json:"belowRangeRecords" bson:"belowRangeRecords"`

	VeryBelowRangeMinutes int64 `json:"veryBelowRangeMinutes" bson:"veryBelowRangeMinutes"`
	VeryBelowRangeRecords int64 `json:"veryBelowRangeRecords" bson:"veryBelowRangeRecords"`

	AboveRangeMinutes int64 `json:"aboveRangeMinutes" bson:"aboveRangeMinutes"`
	AboveRangeRecords int64 `json:"aboveRangeRecords" bson:"aboveRangeRecords"`

	VeryAboveRangeMinutes int64 `json:"veryAboveRangeMinutes" bson:"veryAboveRangeMinutes"`
	VeryAboveRangeRecords int64 `json:"veryAboveRangeRecords" bson:"veryAboveRangeRecords"`

	TotalGlucose    float64   `json:"totalGlucose" bson:"totalGlucose"`
	TotalCGMMinutes int64     `json:"totalCGMMinutes" bson:"totalCGMMinutes"`
	TotalRecords    int64     `json:"totalRecords" bson:"totalRecords"`
	LastRecordTime  time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
}

func NewStats(deviceId string, date time.Time) *Stats {
	return &Stats{
		DeviceID: deviceId,
		Date:     date,

		InRangeMinutes: 0,
		InRangeRecords: 0,

		BelowRangeMinutes: 0,
		BelowRangeRecords: 0,

		VeryBelowRangeMinutes: 0,
		VeryBelowRangeRecords: 0,

		AboveRangeMinutes: 0,
		AboveRangeRecords: 0,

		VeryAboveRangeMinutes: 0,
		VeryAboveRangeRecords: 0,

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

	DailyStats []*Stats `json:"dailyStats" bson:"dailyStats"`

	// date tracking
	LastUpdated   *time.Time `json:"lastUpdated" bson:"lastUpdated"`
	FirstData     *time.Time `json:"firstData" bson:"firstData"`
	LastData      *time.Time `json:"lastData" bson:"lastData"`
	LastUpload    *time.Time `json:"lastUpload" bson:"lastUpload"`
	OutdatedSince *time.Time `json:"outdatedSince" bson:"outdatedSince"`

	TotalDays *int64 `json:"totalDays" bson:"totalDays"`

	TimeCGMUse        *float64 `json:"timeCGMUse" bson:"timeCGMUse"`
	TimeCGMUseMinutes *int64   `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`
	TimeCGMUseRecords *int64   `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`

	// actual values
	AverageGlucose       *Glucose `json:"avgGlucose" bson:"avgGlucose"`
	GlucoseMgmtIndicator *float64 `json:"glucoseMgmtIndicator" bson:"glucoseMgmtIndicator"`

	TimeInRange        *float64 `json:"timeInRange" bson:"timeInRange"`
	TimeInRangeMinutes *int64   `json:"timeInRangeMinutes" bson:"timeInRangeMinutes"`
	TimeInRangeRecords *int64   `json:"timeInRangeRecords" bson:"timeInRangeRecords"`

	TimeBelowRange        *float64 `json:"timeBelowRange" bson:"timeBelowRange"`
	TimeBelowRangeMinutes *int64   `json:"timeBelowRangeMinutes" bson:"timeBelowRangeMinutes"`
	TimeBelowRangeRecords *int64   `json:"timeBelowRangeRecords" bson:"timeBelowRangeRecords"`

	TimeVeryBelowRange        *float64 `json:"timeVeryBelowRange" bson:"timeVeryBelowRange"`
	TimeVeryBelowRangeMinutes *int64   `json:"timeVeryBelowRangeMinutes" bson:"timeVeryBelowRangeMinutes"`
	TimeVeryBelowRangeRecords *int64   `json:"timeVeryBelowRangeRecords" bson:"timeVeryBelowRangeRecords"`

	TimeAboveRange        *float64 `json:"timeAboveRange" bson:"timeAboveRange"`
	TimeAboveRangeMinutes *int64   `json:"timeAboveRangeMinutes" bson:"timeAboveRangeMinutes"`
	TimeAboveRangeRecords *int64   `json:"timeAboveRangeRecords" bson:"timeAboveRangeRecords"`

	TimeVeryAboveRange        *float64 `json:"timeVeryAboveRange" bson:"timeVeryAboveRange"`
	TimeVeryAboveRangeMinutes *int64   `json:"timeVeryAboveRangeMinutes" bson:"timeVeryAboveRangeMinutes"`
	TimeVeryAboveRangeRecords *int64   `json:"timeVeryAboveRangeRecords" bson:"timeVeryAboveRangeRecords"`

	// these are just constants right now.
	HighGlucoseThreshold     *float64 `json:"highGlucoseThreshold" bson:"highGlucoseThreshold"`
	VeryHighGlucoseThreshold *float64 `json:"veryHighGlucoseThreshold" bson:"veryHighGlucoseThreshold"`
	LowGlucoseThreshold      *float64 `json:"lowGlucoseThreshold" bson:"lowGlucoseThreshold"`
	VeryLowGlucoseThreshold  *float64 `json:"VeryLowGlucoseThreshold" bson:"VeryLowGlucoseThreshold"`
}

func New(id string) *Summary {
	return &Summary{
		UserID:        id,
		OutdatedSince: &time.Time{},
	}
}

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *continuous.Continuous) int64 {
	if dataSet.DeviceID != nil {
		if strings.Contains(*dataSet.DeviceID, "AbbottFreeStyleLibre") {
			return 15
		}
	}
	return 5
}

func CalculateGMI(averageGlucose float64) float64 {
	gmi := 12.71 + 4.70587*averageGlucose
	gmi = (0.09148 * gmi) + 2.152
	gmi = math.Round(gmi*10) / 10
	return gmi
}

func (userSummary *Summary) StoreWinningStats(stats map[string]*Stats) error {
	var winningStats *Stats
	var dayCount int
	var oldestDay time.Time
	var oldestDayToKeep time.Time
	var existingDay = false

	if len(stats) < 1 {
		return errors.New("candidate stats empty")
	}

	// find stats with most samples
	for deviceId := range stats {
		if winningStats != nil {
			if stats[deviceId].TotalCGMMinutes > winningStats.TotalCGMMinutes {
				winningStats = stats[deviceId]
			}
		} else {
			winningStats = stats[deviceId]
		}
	}

	// update existing day if one does exist
	if len(userSummary.DailyStats) > 1 {
		for i := len(userSummary.DailyStats) - 1; i >= 0; i-- {
			if userSummary.DailyStats[i].Date.Equal(winningStats.Date) {
				userSummary.DailyStats[i] = winningStats
				existingDay = true
				break
			}
		}
	}
	if existingDay == false {
		userSummary.DailyStats = append(userSummary.DailyStats, winningStats)
	}

	// remove extra days to cap at 14 days of stats
	dayCount = len(userSummary.DailyStats)
	if dayCount > daysAgoToKeep {
		userSummary.DailyStats = userSummary.DailyStats[dayCount-daysAgoToKeep:]
	}

	// remove any stats that are older than 14 days from the last stat
	oldestDay = (*userSummary.DailyStats[0]).Date
	oldestDayToKeep = userSummary.DailyStats[len(userSummary.DailyStats)-1].Date.AddDate(0, 0, -daysAgoToKeep)
	if oldestDay.Before(oldestDayToKeep) {
		for i := range userSummary.DailyStats {
			if userSummary.DailyStats[len(userSummary.DailyStats)-1-i].Date.After(oldestDayToKeep) {
				userSummary.DailyStats = userSummary.DailyStats[len(userSummary.DailyStats)-2-i:]
				break
			}
		}
	}

	return nil
}

func (userSummary *Summary) CalculateStats(userData []*continuous.Continuous) error {
	stats := make(map[string]*Stats)

	var normalizedValue float64
	var duration int64
	var deviceId string
	var recordTime time.Time
	var lastDay time.Time
	var currentDay time.Time
	var err error
	var deviceIdExists bool

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	for _, r := range userData {
		if r.DeviceID != nil {
			deviceId = *r.DeviceID
		} else {
			deviceId = ""
		}

		recordTime, err = time.Parse(time.RFC3339Nano, *r.Time)
		if err != nil {
			return errors.Wrap(err, "cannot parse time in record")
		}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentDay = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			0, 0, 0, 0, recordTime.Location())

		// check if data is in the past somehow, it would currently corrupt stats, this shouldn't be possible
		// but the check is cheap insurance
		if len(userSummary.DailyStats) > 0 {
			if recordTime.Before((*userSummary.DailyStats[len(userSummary.DailyStats)-1]).Date) {
				return errors.Newf("CalculateStats given data before oldest stats for user %s", userSummary.UserID)
			}
		}

		// pick winner for the day, if we are now on the next day
		if !lastDay.IsZero() && !currentDay.Equal(lastDay) {
			err = userSummary.StoreWinningStats(stats)
			if err != nil {
				return err
			}
			stats = make(map[string]*Stats)
		}
		lastDay = currentDay

		_, deviceIdExists = stats[deviceId]
		if !deviceIdExists {
			// create new deviceId in map
			stats[deviceId] = NewStats(deviceId, recordTime.Truncate(24*time.Hour))

			// overwrite with stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			// NOTE2 this may have a rare race condition with multiple devices, as we never store the losing
			// device, resulting in larger batches always winning, even if less complete over time.
			if len(userSummary.DailyStats) > 1 {
				for i := len(userSummary.DailyStats) - 1; i >= 0; i-- {
					if userSummary.DailyStats[i].Date.Equal(currentDay) && userSummary.DailyStats[i].DeviceID == deviceId {
						stats[deviceId] = userSummary.DailyStats[i]
						break
					}
				}
			}
		}

		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(summaryGlucoseUnits))
		duration = GetDuration(r)

		if normalizedValue <= veryLowBloodGlucose {
			stats[deviceId].VeryBelowRangeMinutes += duration
			stats[deviceId].VeryBelowRangeRecords += 1
		} else if normalizedValue >= veryHighBloodGlucose {
			stats[deviceId].VeryAboveRangeMinutes += duration
			stats[deviceId].VeryAboveRangeRecords += 1
		} else if normalizedValue <= lowBloodGlucose {
			stats[deviceId].BelowRangeMinutes += duration
			stats[deviceId].BelowRangeRecords += 1
		} else if normalizedValue >= highBloodGlucose {
			stats[deviceId].AboveRangeMinutes += duration
			stats[deviceId].AboveRangeRecords += 1
		} else {
			stats[deviceId].InRangeMinutes += duration
			stats[deviceId].InRangeRecords += 1
		}

		stats[deviceId].TotalCGMMinutes += duration
		stats[deviceId].TotalGlucose += normalizedValue
		stats[deviceId].TotalRecords += 1
		stats[deviceId].LastRecordTime = recordTime
	}
	// store
	err = userSummary.StoreWinningStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (userSummary *Summary) CalculateSummary() error {
	totalStats := NewStats("summary", time.Time{})
	for _, stats := range userSummary.DailyStats {
		totalStats.InRangeMinutes += stats.InRangeMinutes
		totalStats.InRangeRecords += stats.InRangeRecords

		totalStats.BelowRangeMinutes += stats.BelowRangeMinutes
		totalStats.BelowRangeRecords += stats.BelowRangeRecords

		totalStats.VeryBelowRangeMinutes += stats.VeryBelowRangeMinutes
		totalStats.VeryBelowRangeRecords += stats.VeryBelowRangeRecords

		totalStats.AboveRangeMinutes += stats.AboveRangeMinutes
		totalStats.AboveRangeRecords += stats.AboveRangeRecords

		totalStats.VeryAboveRangeMinutes += stats.VeryAboveRangeMinutes
		totalStats.VeryAboveRangeRecords += stats.VeryAboveRangeRecords

		totalStats.TotalGlucose += stats.TotalGlucose
		totalStats.TotalCGMMinutes += stats.TotalCGMMinutes
		totalStats.TotalRecords += stats.TotalRecords
	}

	// remove partial day (data end) from total time for more accurate TimeCGMUse
	totalMinutes := float64(daysAgoToKeep * 1440)
	lastRecordTime := userSummary.DailyStats[len(userSummary.DailyStats)-1].LastRecordTime
	tomorrow := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day()+1,
		0, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - tomorrow.Sub(lastRecordTime).Minutes()

	userSummary.LastData = &lastRecordTime
	userSummary.FirstData = &userSummary.DailyStats[0].Date

	userSummary.TotalDays = pointer.FromInt64(int64(len(userSummary.DailyStats)))

	userSummary.AverageGlucose = &Glucose{
		Value: pointer.FromFloat64(totalStats.TotalGlucose / float64(totalStats.TotalRecords)),
		Units: pointer.FromString(summaryGlucoseUnits),
	}
	userSummary.TimeInRange = pointer.FromFloat64(
		float64(totalStats.InRangeMinutes) / float64(totalStats.TotalCGMMinutes))
	userSummary.TimeInRangeMinutes = &totalStats.InRangeMinutes
	userSummary.TimeInRangeRecords = &totalStats.InRangeRecords

	userSummary.TimeBelowRange = pointer.FromFloat64(
		float64(totalStats.BelowRangeMinutes) / float64(totalStats.TotalCGMMinutes))
	userSummary.TimeBelowRangeMinutes = &totalStats.BelowRangeMinutes
	userSummary.TimeBelowRangeRecords = &totalStats.BelowRangeRecords

	userSummary.TimeVeryBelowRange = pointer.FromFloat64(
		float64(totalStats.VeryBelowRangeMinutes) / float64(totalStats.TotalCGMMinutes))
	userSummary.TimeVeryBelowRangeMinutes = &totalStats.VeryBelowRangeMinutes
	userSummary.TimeVeryBelowRangeRecords = &totalStats.VeryBelowRangeRecords

	userSummary.TimeAboveRange = pointer.FromFloat64(
		float64(totalStats.AboveRangeMinutes) / float64(totalStats.TotalCGMMinutes))
	userSummary.TimeAboveRangeMinutes = &totalStats.AboveRangeMinutes
	userSummary.TimeAboveRangeRecords = &totalStats.AboveRangeRecords

	userSummary.TimeVeryAboveRange = pointer.FromFloat64(
		float64(totalStats.VeryAboveRangeMinutes) / float64(totalStats.TotalCGMMinutes))
	userSummary.TimeVeryAboveRangeMinutes = &totalStats.VeryAboveRangeMinutes
	userSummary.TimeVeryAboveRangeRecords = &totalStats.VeryAboveRangeRecords

	userSummary.TimeCGMUse = pointer.FromFloat64(
		float64(totalStats.TotalCGMMinutes) / totalMinutes)
	userSummary.TimeCGMUseRecords = &totalStats.TotalRecords
	userSummary.TimeCGMUseMinutes = &totalStats.TotalCGMMinutes

	// we only add GMI if cgm use >70%, otherwise clear it
	if *userSummary.TimeCGMUse > 0.7 {
		userSummary.GlucoseMgmtIndicator = pointer.FromFloat64(CalculateGMI(*userSummary.AverageGlucose.Value))
	} else {
		userSummary.GlucoseMgmtIndicator = nil
	}

	return nil
}

func (userSummary *Summary) Update(ctx context.Context, status *UserLastUpdated, userData []*continuous.Continuous) error {
	var err error
	logger := log.LoggerFromContext(ctx)

	if ctx == nil {
		return errors.New("context is missing")
	}

	if userSummary == nil {
		return errors.New("userSummary is missing")
	}

	if len(userData) == 0 {
		return errors.New("userData is empty")
	}

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	userSummary.LastUpdated = &timestamp
	userSummary.LastUpload = &status.LastUpload
	userSummary.OutdatedSince = nil

	// ensure new data being calculated is after the previously added data to prevent corruption
	if userSummary.LastData != nil {
		oldestRecordTime, err := time.Parse(time.RFC3339Nano, *userData[0].Time)
		if err != nil {
			return err
		}
		if oldestRecordTime.Before(*userSummary.LastData) {
			logger.Debugf("New CGM data for userid %s is before last calculated data", userSummary.UserID)
			return errors.Newf("New CGM data for userid %s is before last calculated data", userSummary.UserID)
		}
	}

	// don't recalculate if there is no new data/this was double called
	if len(userData) < 1 {
		logger.Debugf("No new records for userid %v summary calculation, aborting.", userSummary.UserID)
		return nil
	}

	err = userSummary.CalculateStats(userData)
	if err != nil {
		return err
	}

	err = userSummary.CalculateSummary()
	if err != nil {
		return err
	}

	// add static stuff
	userSummary.LowGlucoseThreshold = pointer.FromFloat64(lowBloodGlucose)
	userSummary.VeryLowGlucoseThreshold = pointer.FromFloat64(veryLowBloodGlucose)
	userSummary.HighGlucoseThreshold = pointer.FromFloat64(highBloodGlucose)
	userSummary.VeryHighGlucoseThreshold = pointer.FromFloat64(veryHighBloodGlucose)

	return nil
}
