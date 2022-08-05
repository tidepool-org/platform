package summary

import (
	"context"
	"math"
	"strconv"
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
	hoursAgoToKeep       = 30 * 24
)

// Glucose reimplementation with only the fields we need, to avoid inheriting Base, which does
// not belong in this collection
type Glucose struct {
	Units string  `json:"units" bson:"units"`
	Value float64 `json:"value" bson:"value"`
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
	HasAverageGlucose             bool `json:"hasAverageGlucose" bson:"hasAverageGlucose"`
	HasGlucoseManagementIndicator bool `json:"hasGlucoseManagementIndicator" bson:"hasGlucoseManagementIndicator"`
	HasTimeCGMUsePercent          bool `json:"hasTimeCGMUsePercent" bson:"hasTimeCGMUsePercent"`
	HasTimeInTargetPercent        bool `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	HasTimeInHighPercent          bool `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	HasTimeInVeryHighPercent      bool `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	HasTimeInLowPercent           bool `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	HasTimeInVeryLowPercent       bool `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`

	TimeCGMUsePercent *float64 `json:"timeCGMUsePercent" bson:"timeCGMUsePercent"`
	TimeCGMUseMinutes int      `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`
	TimeCGMUseRecords int      `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`

	// actual values
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

type Summary struct {
	ID     primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserID string             `json:"userId" bson:"userId"`

	HasLastUploadDate bool `json:"hasLastUploadDate" bson:"hasLastUploadDate"`

	HourlyStats []*Stats           `json:"hourlyStats" bson:"hourlyStats"`
	Periods     map[string]*Period `json:"periods" bson:"periods"`

	// date tracking
	LastUpdatedDate time.Time  `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	FirstData       time.Time  `json:"firstData" bson:"firstData"`
	LastData        *time.Time `json:"lastData" bson:"lastData"`
	LastUploadDate  time.Time  `json:"lastUploadDate" bson:"lastUploadDate"`
	OutdatedSince   *time.Time `json:"outdatedSince" bson:"outdatedSince"`

	TotalHours int `json:"totalHours" bson:"totalHours"`

	// these are just constants right now.
	HighGlucoseThreshold     float64 `json:"highGlucoseThreshold" bson:"highGlucoseThreshold"`
	VeryHighGlucoseThreshold float64 `json:"veryHighGlucoseThreshold" bson:"veryHighGlucoseThreshold"`
	LowGlucoseThreshold      float64 `json:"lowGlucoseThreshold" bson:"lowGlucoseThreshold"`
	VeryLowGlucoseThreshold  float64 `json:"VeryLowGlucoseThreshold" bson:"VeryLowGlucoseThreshold"`
}

func New(id string) *Summary {
	return &Summary{
		UserID:        id,
		OutdatedSince: &time.Time{},
		Periods:       make(map[string]*Period),
		HourlyStats:   make([]*Stats, 0),
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
	var hourCount int
	var oldestHour time.Time
	var oldestHourToKeep time.Time
	var existingDay = false
	var statsGap int
	var newStatsTime time.Time

	if stats == nil {
		return errors.New("stats empty")
	}

	// update existing hour if one does exist
	if len(userSummary.HourlyStats) > 0 {
		for i := len(userSummary.HourlyStats) - 1; i >= 0; i-- {
			if userSummary.HourlyStats[i].Date.Equal(stats.Date) {
				userSummary.HourlyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if userSummary.HourlyStats[i].Date.After(stats.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(stats.Date.Sub(userSummary.HourlyStats[len(userSummary.HourlyStats)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = stats.Date.Add(time.Duration(-i) * time.Hour)
			userSummary.HourlyStats = append(userSummary.HourlyStats, NewStats(newStatsTime))
		}
	}

	if existingDay == false {
		userSummary.HourlyStats = append(userSummary.HourlyStats, stats)
	}

	// remove extra days to cap at X days of stats
	hourCount = len(userSummary.HourlyStats)
	if hourCount > hoursAgoToKeep {
		userSummary.HourlyStats = userSummary.HourlyStats[hourCount-hoursAgoToKeep:]
	}

	// remove any stats that are older than X days from the last stat
	oldestHour = userSummary.HourlyStats[0].Date
	oldestHourToKeep = stats.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(userSummary.HourlyStats) - 2; i >= 0; i-- {
			if userSummary.HourlyStats[i].Date.Before(oldestHourToKeep) {
				userSummary.HourlyStats = userSummary.HourlyStats[i+1:]
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
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var stats *Stats

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	// skip past data
	if len(userSummary.HourlyStats) > 0 {
		userData, err = SkipUntil(userSummary.HourlyStats[len(userSummary.HourlyStats)-1].Date, userData)
	}

	for _, r := range userData {
		recordTime, err = time.Parse(time.RFC3339Nano, *r.Time)
		if err != nil {
			return errors.Wrap(err, "cannot parse time in record")
		}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentHour = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store stats for the day, if we are now on the next day
		if !lastHour.IsZero() && !currentHour.Equal(lastHour) {
			err = userSummary.AddStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(userSummary.HourlyStats) > 0 {
				for i := len(userSummary.HourlyStats) - 1; i >= 0; i-- {
					if userSummary.HourlyStats[i].Date.Equal(currentHour) {
						stats = userSummary.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if userSummary.HourlyStats[i].Date.After(currentHour) {
						break
					}
				}
			}

			if stats == nil {
				stats = NewStats(currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(userSummary.HourlyStats) > 0 {
			stats.LastRecordTime = userSummary.HourlyStats[len(userSummary.HourlyStats)-1].LastRecordTime
		}

		// duration has never been calculated, use current record's duration for this cycle
		if duration == 0 {
			duration = GetDuration(r)
		}

		// calculate skipWindow based on duration of previous value
		skipWindow := time.Duration(duration)*time.Minute - 3*time.Second

		// if we are too close to the previous value, skip
		if recordTime.Sub(stats.LastRecordTime) > skipWindow {
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

func (userSummary *Summary) CalculateSummary() {
	totalStats := NewStats(time.Time{})

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(userSummary.HourlyStats); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			userSummary.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(userSummary.HourlyStats) - 1 - i
		totalStats.TargetMinutes += userSummary.HourlyStats[currentIndex].TargetMinutes
		totalStats.TargetRecords += userSummary.HourlyStats[currentIndex].TargetRecords

		totalStats.LowMinutes += userSummary.HourlyStats[currentIndex].LowMinutes
		totalStats.LowRecords += userSummary.HourlyStats[currentIndex].LowRecords

		totalStats.VeryLowMinutes += userSummary.HourlyStats[currentIndex].VeryLowMinutes
		totalStats.VeryLowRecords += userSummary.HourlyStats[currentIndex].VeryLowRecords

		totalStats.HighMinutes += userSummary.HourlyStats[currentIndex].HighMinutes
		totalStats.HighRecords += userSummary.HourlyStats[currentIndex].HighRecords

		totalStats.VeryHighMinutes += userSummary.HourlyStats[currentIndex].VeryHighMinutes
		totalStats.VeryHighRecords += userSummary.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += userSummary.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalCGMMinutes += userSummary.HourlyStats[currentIndex].TotalCGMMinutes
		totalStats.TotalCGMRecords += userSummary.HourlyStats[currentIndex].TotalCGMRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		userSummary.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (userSummary *Summary) CalculatePeriod(i int, totalStats *Stats) {
	var timeCGMUsePercent *float64
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var glucoseManagementIndicator *float64
	var averageGlucose *Glucose

	// remove partial hour (data end) from total time for more accurate TimeCGMUse
	totalMinutes := float64(i * 24 * 60)
	lastRecordTime := userSummary.HourlyStats[len(userSummary.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	userSummary.LastData = &lastRecordTime
	userSummary.FirstData = userSummary.HourlyStats[0].Date

	userSummary.TotalHours = len(userSummary.HourlyStats)

	// calculate derived summary stats
	if totalMinutes != 0 {
		timeCGMUsePercent = pointer.FromFloat64(float64(totalStats.TotalCGMMinutes) / totalMinutes)
	}

	if totalStats.TotalCGMMinutes != 0 {
		// if we are storing under 1d, apply 70% rule to TimeIn*
		// if we are storing over 1d, check for 24h cgm use
		if (i <= 1 && *timeCGMUsePercent < 0.7) || (i > 1 && totalStats.TotalCGMMinutes > 1440) {
			timeInTargetPercent = pointer.FromFloat64(float64(totalStats.TargetMinutes) / float64(totalStats.TotalCGMMinutes))
			timeInLowPercent = pointer.FromFloat64(float64(totalStats.LowMinutes) / float64(totalStats.TotalCGMMinutes))
			timeInVeryLowPercent = pointer.FromFloat64(float64(totalStats.VeryLowMinutes) / float64(totalStats.TotalCGMMinutes))
			timeInHighPercent = pointer.FromFloat64(float64(totalStats.HighMinutes) / float64(totalStats.TotalCGMMinutes))
			timeInVeryHighPercent = pointer.FromFloat64(float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalCGMMinutes))
		}

	}

	if totalStats.TotalCGMRecords != 0 {
		averageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalCGMRecords),
			Units: summaryGlucoseUnits,
		}
	}

	// we only add GMI if cgm use >70%, otherwise clear it
	glucoseManagementIndicator = nil
	if *timeCGMUsePercent > 0.7 {
		glucoseManagementIndicator = pointer.FromFloat64(CalculateGMI(averageGlucose.Value))
	}

	// ensure periods exists, just in case
	if userSummary.Periods == nil {
		userSummary.Periods = make(map[string]*Period)
	}

	userSummary.Periods[strconv.Itoa(i)+"d"] = &Period{
		HasAverageGlucose:             averageGlucose != nil,
		HasGlucoseManagementIndicator: glucoseManagementIndicator != nil,
		HasTimeCGMUsePercent:          timeCGMUsePercent != nil,
		HasTimeInTargetPercent:        timeInTargetPercent != nil,
		HasTimeInLowPercent:           timeInLowPercent != nil,
		HasTimeInVeryLowPercent:       timeInVeryLowPercent != nil,
		HasTimeInHighPercent:          timeInHighPercent != nil,
		HasTimeInVeryHighPercent:      timeInVeryHighPercent != nil,

		TimeCGMUsePercent: timeCGMUsePercent,
		TimeCGMUseMinutes: totalStats.TotalCGMMinutes,
		TimeCGMUseRecords: totalStats.TotalCGMRecords,

		AverageGlucose:             averageGlucose,
		GlucoseManagementIndicator: glucoseManagementIndicator,

		TimeInTargetPercent: timeInTargetPercent,
		TimeInTargetMinutes: totalStats.TargetMinutes,
		TimeInTargetRecords: totalStats.TargetRecords,

		TimeInLowPercent: timeInLowPercent,
		TimeInLowMinutes: totalStats.LowMinutes,
		TimeInLowRecords: totalStats.LowRecords,

		TimeInVeryLowPercent: timeInVeryLowPercent,
		TimeInVeryLowMinutes: totalStats.VeryLowMinutes,
		TimeInVeryLowRecords: totalStats.VeryLowRecords,

		TimeInHighPercent: timeInHighPercent,
		TimeInHighMinutes: totalStats.HighMinutes,
		TimeInHighRecords: totalStats.HighRecords,

		TimeInVeryHighPercent: timeInVeryHighPercent,
		TimeInVeryHighMinutes: totalStats.VeryHighMinutes,
		TimeInVeryHighRecords: totalStats.VeryHighRecords,
	}
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
	userSummary.LastUpdatedDate = timestamp
	userSummary.OutdatedSince = nil
	userSummary.LastUploadDate = status.LastUpload

	// technically, this never could be zero, but we check anyway
	userSummary.HasLastUploadDate = !status.LastUpload.IsZero()

	// remove any past values that squeeze through the string date query that feeds this function
	// this mostly occurs when different sources use different time precisions (s vs ms vs ns)
	// resulting in $gt 00:00:01.275Z pulling in 00:00:01Z, which is before.
	if userSummary.LastData != nil {
		userData, err = SkipUntil(*userSummary.LastData, userData)
		if err != nil {
			return err
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

	userSummary.CalculateSummary()

	// add static stuff
	userSummary.LowGlucoseThreshold = lowBloodGlucose
	userSummary.VeryLowGlucoseThreshold = veryLowBloodGlucose
	userSummary.HighGlucoseThreshold = highBloodGlucose
	userSummary.VeryHighGlucoseThreshold = veryHighBloodGlucose

	return nil
}

func SkipUntil(date time.Time, userData []*continuous.Continuous) ([]*continuous.Continuous, error) {
	var skip int
	for i := 0; i < len(userData); i++ {
		recordTime, err := time.Parse(time.RFC3339Nano, *userData[i].Time)
		if err != nil {
			return nil, err
		}

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
