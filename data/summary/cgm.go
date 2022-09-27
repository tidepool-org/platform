package summary

import (
	"context"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
)

type UserCGMLastUpdated struct {
	LastData   time.Time
	LastUpload time.Time
}

type CGMStats struct {
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

	TotalGlucose float64 `json:"totalGlucose" bson:"totalGlucose"`
	TotalMinutes int     `json:"totalMinutes" bson:"totalMinutes"`
	TotalRecords int     `json:"totalRecords" bson:"totalRecords"`

	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
}

type CGMPeriod struct {
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

type CGMSummary struct {
	Periods     map[string]*CGMPeriod `json:"periods" bson:"periods"`
	HourlyStats []*CGMStats           `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int                   `json:"totalHours" bson:"totalHours"`

	// date tracking
	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    time.Time  `json:"lastUploadDate" bson:"lastUploadDate"`
	LastUpdatedDate   time.Time  `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	FirstData         time.Time  `json:"firstData" bson:"firstData"`
	LastData          *time.Time `json:"lastData" bson:"lastData"`
	OutdatedSince     *time.Time `json:"outdatedSince" bson:"outdatedSince"`
}

func NewCGMStats(date time.Time) *CGMStats {
	return &CGMStats{
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

		TotalGlucose: 0,
		TotalMinutes: 0,
		TotalRecords: 0,
	}
}

func (userSummary *Summary) AddCGMStats(stats *CGMStats) error {
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
	if len(userSummary.CGM.HourlyStats) > 0 {
		for i := len(userSummary.CGM.HourlyStats) - 1; i >= 0; i-- {
			if userSummary.CGM.HourlyStats[i].Date.Equal(stats.Date) {
				userSummary.CGM.HourlyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if userSummary.CGM.HourlyStats[i].Date.After(stats.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(stats.Date.Sub(userSummary.CGM.HourlyStats[len(userSummary.CGM.HourlyStats)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = stats.Date.Add(time.Duration(-i) * time.Hour)
			userSummary.CGM.HourlyStats = append(userSummary.CGM.HourlyStats, NewCGMStats(newStatsTime))
		}
	}

	if existingDay == false {
		userSummary.CGM.HourlyStats = append(userSummary.CGM.HourlyStats, stats)
	}

	// remove extra days to cap at X days of stats
	hourCount = len(userSummary.CGM.HourlyStats)
	if hourCount > hoursAgoToKeep {
		userSummary.CGM.HourlyStats = userSummary.CGM.HourlyStats[hourCount-hoursAgoToKeep:]
	}

	// remove any stats that are older than X days from the last stat
	oldestHour = userSummary.CGM.HourlyStats[0].Date
	oldestHourToKeep = stats.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(userSummary.CGM.HourlyStats) - 2; i >= 0; i-- {
			if userSummary.CGM.HourlyStats[i].Date.Before(oldestHourToKeep) {
				userSummary.CGM.HourlyStats = userSummary.CGM.HourlyStats[i+1:]
				break
			}
		}
	}

	return nil
}

func (userSummary *Summary) CalculateCGMStats(userData []*glucoseDatum.Glucose) error {
	var normalizedValue float64
	var duration int
	var recordTime time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var stats *CGMStats

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	// skip past data
	if len(userSummary.CGM.HourlyStats) > 0 {
		userData, err = SkipUntil(userSummary.CGM.HourlyStats[len(userSummary.CGM.HourlyStats)-1].Date, userData)
	}

	for _, r := range userData {
		recordTime = *r.Time
		if err != nil {
			return errors.Wrap(err, "cannot parse time in record")
		}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentHour = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store stats for the day, if we are now on the next day
		if !lastHour.IsZero() && !currentHour.Equal(lastHour) {
			err = userSummary.AddCGMStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(userSummary.CGM.HourlyStats) > 0 {
				for i := len(userSummary.CGM.HourlyStats) - 1; i >= 0; i-- {
					if userSummary.CGM.HourlyStats[i].Date.Equal(currentHour) {
						stats = userSummary.CGM.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if userSummary.CGM.HourlyStats[i].Date.After(currentHour) {
						break
					}
				}
			}

			if stats == nil {
				stats = NewCGMStats(currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(userSummary.CGM.HourlyStats) > 0 {
			stats.LastRecordTime = userSummary.CGM.HourlyStats[len(userSummary.CGM.HourlyStats)-1].LastRecordTime
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

			stats.TotalMinutes += duration
			stats.TotalRecords++
			stats.TotalGlucose += normalizedValue
			stats.LastRecordTime = recordTime
		}
	}
	// store
	err = userSummary.AddCGMStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (userSummary *Summary) CalculateCGMSummary() {
	totalStats := NewCGMStats(time.Time{})

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(userSummary.CGM.HourlyStats); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			userSummary.CalculateCGMPeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(userSummary.CGM.HourlyStats) - 1 - i
		totalStats.TargetMinutes += userSummary.CGM.HourlyStats[currentIndex].TargetMinutes
		totalStats.TargetRecords += userSummary.CGM.HourlyStats[currentIndex].TargetRecords

		totalStats.LowMinutes += userSummary.CGM.HourlyStats[currentIndex].LowMinutes
		totalStats.LowRecords += userSummary.CGM.HourlyStats[currentIndex].LowRecords

		totalStats.VeryLowMinutes += userSummary.CGM.HourlyStats[currentIndex].VeryLowMinutes
		totalStats.VeryLowRecords += userSummary.CGM.HourlyStats[currentIndex].VeryLowRecords

		totalStats.HighMinutes += userSummary.CGM.HourlyStats[currentIndex].HighMinutes
		totalStats.HighRecords += userSummary.CGM.HourlyStats[currentIndex].HighRecords

		totalStats.VeryHighMinutes += userSummary.CGM.HourlyStats[currentIndex].VeryHighMinutes
		totalStats.VeryHighRecords += userSummary.CGM.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += userSummary.CGM.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalMinutes += userSummary.CGM.HourlyStats[currentIndex].TotalMinutes
		totalStats.TotalRecords += userSummary.CGM.HourlyStats[currentIndex].TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		userSummary.CalculateCGMPeriod(stopPoints[i], totalStats)
	}
}

func (userSummary *Summary) CalculateCGMPeriod(i int, totalStats *CGMStats) {
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
	lastRecordTime := userSummary.CGM.HourlyStats[len(userSummary.CGM.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	userSummary.CGM.LastData = &lastRecordTime
	userSummary.CGM.FirstData = userSummary.CGM.HourlyStats[0].Date

	userSummary.CGM.TotalHours = len(userSummary.CGM.HourlyStats)

	// calculate derived summary stats
	if totalMinutes != 0 {
		timeCGMUsePercent = pointer.FromFloat64(float64(totalStats.TotalMinutes) / totalMinutes)
	}

	if totalStats.TotalMinutes != 0 {
		// if we are storing under 1d, apply 70% rule to TimeIn*
		// if we are storing over 1d, check for 24h cgm use
		if (i <= 1 && *timeCGMUsePercent > 0.7) || (i > 1 && totalStats.TotalMinutes > 1440) {
			timeInTargetPercent = pointer.FromFloat64(float64(totalStats.TargetMinutes) / float64(totalStats.TotalMinutes))
			timeInLowPercent = pointer.FromFloat64(float64(totalStats.LowMinutes) / float64(totalStats.TotalMinutes))
			timeInVeryLowPercent = pointer.FromFloat64(float64(totalStats.VeryLowMinutes) / float64(totalStats.TotalMinutes))
			timeInHighPercent = pointer.FromFloat64(float64(totalStats.HighMinutes) / float64(totalStats.TotalMinutes))
			timeInVeryHighPercent = pointer.FromFloat64(float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalMinutes))
		}

	}

	if totalStats.TotalRecords != 0 {
		averageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalRecords),
			Units: summaryGlucoseUnits,
		}
	}

	// we only add GMI if cgm use >70%, otherwise clear it
	glucoseManagementIndicator = nil
	if *timeCGMUsePercent > 0.7 {
		glucoseManagementIndicator = pointer.FromFloat64(CalculateGMI(averageGlucose.Value))
	}

	// ensure periods exists, just in case
	if userSummary.CGM.Periods == nil {
		userSummary.CGM.Periods = make(map[string]*CGMPeriod)
	}

	userSummary.CGM.Periods[strconv.Itoa(i)+"d"] = &CGMPeriod{
		HasAverageGlucose:             averageGlucose != nil,
		HasGlucoseManagementIndicator: glucoseManagementIndicator != nil,
		HasTimeCGMUsePercent:          timeCGMUsePercent != nil,
		HasTimeInTargetPercent:        timeInTargetPercent != nil,
		HasTimeInLowPercent:           timeInLowPercent != nil,
		HasTimeInVeryLowPercent:       timeInVeryLowPercent != nil,
		HasTimeInHighPercent:          timeInHighPercent != nil,
		HasTimeInVeryHighPercent:      timeInVeryHighPercent != nil,

		TimeCGMUsePercent: timeCGMUsePercent,
		TimeCGMUseMinutes: totalStats.TotalMinutes,
		TimeCGMUseRecords: totalStats.TotalRecords,

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

func (userSummary *Summary) UpdateCGM(ctx context.Context, status *UserCGMLastUpdated, userCGMData []*glucoseDatum.Glucose) error {
	var err error
	logger := log.LoggerFromContext(ctx)

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	userSummary.CGM.LastUpdatedDate = timestamp
	userSummary.CGM.OutdatedSince = nil
	userSummary.CGM.LastUploadDate = status.LastUpload

	// technically, this never could be zero, but we check anyway
	userSummary.CGM.HasLastUploadDate = !status.LastUpload.IsZero()

	// remove any past values that squeeze through the string date query that feeds this function
	// this mostly occurs when different sources use different time precisions (s vs ms vs ns)
	// resulting in $gt 00:00:01.275Z pulling in 00:00:01Z, which is before.
	if userSummary.CGM.LastData != nil {
		userCGMData, err = SkipUntil(*userSummary.CGM.LastData, userCGMData)
		if err != nil {
			return err
		}
	}

	// don't recalculate if there is no new data/this was double called
	if len(userCGMData) < 1 {
		logger.Debugf("No new records for userid %v summary calculation, aborting.", userSummary.UserID)
		return nil
	}

	err = userSummary.CalculateCGMStats(userCGMData)
	if err != nil {
		return err
	}

	userSummary.CalculateCGMSummary()

	return nil
}
