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
	Units *string  `json:"units" bson:"units"`
	Value *float64 `json:"value" bson:"value"`
}

type Stats struct {
	Date time.Time `json:"date" bson:"date"`

	TargetMinutes int `json:"targetMinutes" bson:"targetMinutes"`
	TargetRecords int `json:"targetRecords" bson:"targetRecords"`

	LowMinutes int `json:"lowMinutes" bson:"lowMinutes"`
	LowRecords int `json:"lowRecords" bson:"lowRecords"`

	VeryLowMinutes int `json:"veryLowMinutes" bson:"veryLowMinutes"`
	VeryLowRecords int `json:"veryLowRecords" bson:"veryLowRecords"`

	HighMinutes int `json:"highMinutes" bson:"highMinutes"`
	HighRecords int `json:"highRecords" bson:"highRecords"`

	VeryHighMinutes int `json:"veryHighMinutes" bson:"veryHighMinutes"`
	VeryHighRecords int `json:"veryHighRecords" bson:"veryHighRecords"`

	TotalGlucose    float64 `json:"totalGlucose" bson:"totalGlucose"`
	TotalCGMMinutes int     `json:"totalCGMMinutes" bson:"totalCGMMinutes"`
	TotalCGMRecords int     `json:"totalCGMRecords" bson:"totalCGMRecords"`

	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
}

func NewStats(date time.Time) *Stats {
	return &Stats{
		Date: date,

		TargetMinutes: 0,
		TargetRecords: 0,

		LowMinutes: 0,
		LowRecords: 0,

		VeryLowMinutes: 0,
		VeryLowRecords: 0,

		HighMinutes: 0,
		HighRecords: 0,

		VeryHighMinutes: 0,
		VeryHighRecords: 0,

		TotalGlucose:    0,
		TotalCGMMinutes: 0,
		TotalCGMRecords: 0,
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

type Period struct {
	TimeCGMUsePercent *float64 `json:"timeCGMUsePercent" bson:"timeCGMUsePercent"`
	TimeCGMUseMinutes *int     `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`
	TimeCGMUseRecords *int     `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`

	// actual values
	AverageGlucose             *Glucose `json:"avgGlucose" bson:"avgGlucose"`
	GlucoseManagementIndicator *float64 `json:"glucoseManagementIndicator" bson:"glucoseManagementIndicator"`

	TimeInTargetPercent *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`
	TimeInTargetMinutes *int     `json:"timeInTargetMinutes" bson:"timeInTargetMinutes"`
	TimeInTargetRecords *int     `json:"timeInTargetRecords" bson:"timeInTargetRecords"`

	TimeInLowPercent *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`
	TimeInLowMinutes *int     `json:"timeInLowMinutes" bson:"timeInLowMinutes"`
	TimeInLowRecords *int     `json:"timeInLowRecords" bson:"timeInLowRecords"`

	TimeInVeryLowPercent *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`
	TimeInVeryLowMinutes *int     `json:"timeInVeryLowMinutes" bson:"timeInVeryLowMinutes"`
	TimeInVeryLowRecords *int     `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`

	TimeInHighPercent *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`
	TimeInHighMinutes *int     `json:"timeInHighMinutes" bson:"timeInHighMinutes"`
	TimeInHighRecords *int     `json:"timeInHighRecords" bson:"timeInHighRecords"`

	TimeInVeryHighPercent *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`
	TimeInVeryHighMinutes *int     `json:"timeInVeryHighMinutes" bson:"timeInVeryHighMinutes"`
	TimeInVeryHighRecords *int     `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
}

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`

	DailyStats []*Stats           `json:"dailyStats" bson:"dailyStats"`
	Periods    map[string]*Period `json:"periods" bson:"periods"`

	// date tracking
	LastUpdatedDate *time.Time `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	FirstData       *time.Time `json:"firstData" bson:"firstData"`
	LastData        *time.Time `json:"lastData" bson:"lastData"`
	LastUploadDate  *time.Time `json:"lastUploadDate" bson:"lastUploadDate"`
	OutdatedSince   *time.Time `json:"outdatedSince" bson:"outdatedSince"`

	TotalDays *int `json:"totalDays" bson:"totalDays"`

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
		Periods:       make(map[string]*Period),
		DailyStats:    make([]*Stats, 0),
	}
}

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *continuous.Continuous) int {
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

func (userSummary *Summary) AddStats(stats *Stats) error {
	var dayCount int
	var oldestDay time.Time
	var oldestDayToKeep time.Time
	var existingDay = false

	if stats == nil {
		return errors.New("stats empty")
	}

	// update existing day if one does exist
	if len(userSummary.DailyStats) > 0 {
		for i := len(userSummary.DailyStats) - 1; i >= 0; i-- {
			if userSummary.DailyStats[i].Date.Equal(stats.Date) {
				userSummary.DailyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if userSummary.DailyStats[i].Date.After(stats.Date) {
				break
			}
		}
	}

	if existingDay == false {
		userSummary.DailyStats = append(userSummary.DailyStats, stats)
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
		// we don't check the last entry because we just added/updated it
		for i := len(userSummary.DailyStats) - 2; i >= 0; i-- {
			if userSummary.DailyStats[i].Date.Before(oldestDayToKeep) {
				userSummary.DailyStats = userSummary.DailyStats[i+1:]
				break
			}
		}
	}

	return nil
}

func (userSummary *Summary) CalculateStats(userData []*continuous.Continuous) error {
	var normalizedValue float64
	var duration int
	var recordTime time.Time
	var lastDay time.Time
	var currentDay time.Time
	var err error
	var stats *Stats

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	for _, r := range userData {
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
			if recordTime.Before(userSummary.DailyStats[len(userSummary.DailyStats)-1].Date) {
				return errors.Newf("CalculateStats given data before oldest stats for user %s", userSummary.UserID)
			}
		}

		// store stats for the day, if we are now on the next day
		if !lastDay.IsZero() && !currentDay.Equal(lastDay) {
			err = userSummary.AddStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(userSummary.DailyStats) > 0 {
				for i := len(userSummary.DailyStats) - 1; i >= 0; i-- {
					if userSummary.DailyStats[i].Date.Equal(currentDay) {
						stats = userSummary.DailyStats[i]
						break
					}

					// we already passed our date, give up
					if userSummary.DailyStats[i].Date.After(currentDay) {
						break
					}
				}
			}

			if stats == nil {
				stats = NewStats(recordTime.Truncate(24 * time.Hour))
			}
		}

		lastDay = currentDay

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(userSummary.DailyStats) > 1 {
			stats.LastRecordTime = userSummary.DailyStats[len(userSummary.DailyStats)-1].LastRecordTime
		}

		// if we are too close to the previous value, skip
		// 45 seconds is arbitrary, but under one minute to allow future devices with 1 minute readings
		if recordTime.Sub(stats.LastRecordTime) > 45*time.Second {
			normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(summaryGlucoseUnits))
			duration = GetDuration(r)

			if normalizedValue <= veryLowBloodGlucose {
				stats.VeryLowMinutes += duration
				stats.VeryLowRecords++
			} else if normalizedValue >= veryHighBloodGlucose {
				stats.VeryHighMinutes += duration
				stats.VeryHighRecords++
			} else if normalizedValue <= lowBloodGlucose {
				stats.LowMinutes += duration
				stats.LowRecords++
			} else if normalizedValue >= highBloodGlucose {
				stats.HighMinutes += duration
				stats.HighRecords++
			} else {
				stats.TargetMinutes += duration
				stats.TargetRecords++
			}

			stats.TotalCGMMinutes += duration
			stats.TotalCGMRecords++
			stats.TotalGlucose += normalizedValue
			stats.LastRecordTime = recordTime
		}
	}
	// store
	err = userSummary.AddStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (userSummary *Summary) CalculateSummary() error {
	var timeCGMUsePercent float64
	var timeInTargetPercent float64
	var timeInLowPercent float64
	var timeInVeryLowPercent float64
	var timeInHighPercent float64
	var timeInVeryHighPercent float64
	var glucoseManagementIndicator *float64
	var averageGlucose float64

	totalStats := NewStats(time.Time{})

	for _, stats := range userSummary.DailyStats {
		totalStats.TargetMinutes += stats.TargetMinutes
		totalStats.TargetRecords += stats.TargetRecords

		totalStats.LowMinutes += stats.LowMinutes
		totalStats.LowRecords += stats.LowRecords

		totalStats.VeryLowMinutes += stats.VeryLowMinutes
		totalStats.VeryLowRecords += stats.VeryLowRecords

		totalStats.HighMinutes += stats.HighMinutes
		totalStats.HighRecords += stats.HighRecords

		totalStats.VeryHighMinutes += stats.VeryHighMinutes
		totalStats.VeryHighRecords += stats.VeryHighRecords

		totalStats.TotalGlucose += stats.TotalGlucose
		totalStats.TotalCGMMinutes += stats.TotalCGMMinutes
		totalStats.TotalCGMRecords += stats.TotalCGMRecords
	}

	// remove partial day (data end) from total time for more accurate TimeCGMUse
	totalMinutes := float64(daysAgoToKeep * 1440)
	lastRecordTime := userSummary.DailyStats[len(userSummary.DailyStats)-1].LastRecordTime
	tomorrow := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day()+1,
		0, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - tomorrow.Sub(lastRecordTime).Minutes()

	userSummary.LastData = &lastRecordTime
	userSummary.FirstData = &userSummary.DailyStats[0].Date

	userSummary.TotalDays = pointer.FromInt(len(userSummary.DailyStats))

	// calculate derived summary stats
	timeCGMUsePercent = float64(totalStats.TotalCGMMinutes) / totalMinutes
	timeInTargetPercent = float64(totalStats.TargetMinutes) / float64(totalStats.TotalCGMMinutes)
	timeInLowPercent = float64(totalStats.LowMinutes) / float64(totalStats.TotalCGMMinutes)
	timeInVeryLowPercent = float64(totalStats.VeryLowMinutes) / float64(totalStats.TotalCGMMinutes)
	timeInHighPercent = float64(totalStats.HighMinutes) / float64(totalStats.TotalCGMMinutes)
	timeInVeryHighPercent = float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalCGMMinutes)
	averageGlucose = totalStats.TotalGlucose / float64(totalStats.TotalCGMRecords)

	// we only add GMI if cgm use >70%, otherwise clear it
	glucoseManagementIndicator = nil
	if timeCGMUsePercent > 0.7 {
		glucoseManagementIndicator = pointer.FromFloat64(CalculateGMI(averageGlucose))
	}

	// ensure periods exists, just in case
	if userSummary.Periods == nil {
		userSummary.Periods = make(map[string]*Period)
	}

	// statically place stats into the 14-day period slot for now.
	userSummary.Periods["14d"] = &Period{
		TimeCGMUsePercent: &timeCGMUsePercent,
		TimeCGMUseMinutes: &totalStats.TotalCGMMinutes,
		TimeCGMUseRecords: &totalStats.TotalCGMRecords,

		AverageGlucose: &Glucose{
			Value: pointer.FromFloat64(averageGlucose),
			Units: pointer.FromString(summaryGlucoseUnits),
		},
		GlucoseManagementIndicator: glucoseManagementIndicator,

		TimeInTargetPercent: &timeInTargetPercent,
		TimeInTargetMinutes: &totalStats.TargetMinutes,
		TimeInTargetRecords: &totalStats.TargetRecords,

		TimeInLowPercent: &timeInLowPercent,
		TimeInLowMinutes: &totalStats.LowMinutes,
		TimeInLowRecords: &totalStats.LowRecords,

		TimeInVeryLowPercent: &timeInVeryLowPercent,
		TimeInVeryLowMinutes: &totalStats.VeryLowMinutes,
		TimeInVeryLowRecords: &totalStats.VeryLowRecords,

		TimeInHighPercent: &timeInHighPercent,
		TimeInHighMinutes: &totalStats.HighMinutes,
		TimeInHighRecords: &totalStats.HighRecords,

		TimeInVeryHighPercent: &timeInVeryHighPercent,
		TimeInVeryHighMinutes: &totalStats.VeryHighMinutes,
		TimeInVeryHighRecords: &totalStats.VeryHighRecords,
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

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	userSummary.LastUpdatedDate = &timestamp
	userSummary.LastUploadDate = &status.LastUpload
	userSummary.OutdatedSince = nil

	// remove any past values that squeeze through the string date query that feeds this function
	// this mostly occurs when different sources use different time precisions (s vs ms vs ns)
	// resulting in $gt 00:00:01.275Z pulling in 00:00:01Z, which is before.
	if userSummary.LastData != nil {
		var skip int
		for i := 0; i < len(userData); i++ {
			recordTime, err := time.Parse(time.RFC3339Nano, *userData[i].Time)
			if err != nil {
				return err
			}

			if recordTime.Before(*userSummary.LastData) {
				skip = i + 1
			} else {
				break
			}
		}

		if skip > 0 {
			logger.Debugf("New CGM data for userid %s is before last calculated data, skipping first %d records", userSummary.UserID, skip)
			userData = userData[skip:]
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
