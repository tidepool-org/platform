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

type UserBGMLastUpdated struct {
	LastData   time.Time
	LastUpload time.Time
}

type BGMStats struct {
	Date time.Time `json:"date" bson:"date"`

	TargetRecords   int `json:"targetRecords" bson:"targetRecords"`
	LowRecords      int `json:"lowRecords" bson:"lowRecords"`
	VeryLowRecords  int `json:"veryLowRecords" bson:"veryLowRecords"`
	HighRecords     int `json:"highRecords" bson:"highRecords"`
	VeryHighRecords int `json:"veryHighRecords" bson:"veryHighRecords"`

	TotalGlucose float64 `json:"totalGlucose" bson:"totalGlucose"`
	TotalRecords int     `json:"totalRecords" bson:"totalRecords"`

	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
}

type BGMPeriod struct {
	HasAverageGlucose        bool `json:"hasAverageGlucose" bson:"hasAverageGlucose"`
	HasTimeInTargetPercent   bool `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	HasTimeInHighPercent     bool `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	HasTimeInVeryHighPercent bool `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	HasTimeInLowPercent      bool `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	HasTimeInVeryLowPercent  bool `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`

	// actual values
	AverageGlucose *Glucose `json:"averageGlucose" bson:"avgGlucose"`

	TimeInTargetPercent *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`
	TimeInTargetRecords int      `json:"timeInTargetRecords" bson:"timeInTargetRecords"`

	TimeInLowPercent *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`
	TimeInLowRecords int      `json:"timeInLowRecords" bson:"timeInLowRecords"`

	TimeInVeryLowPercent *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`
	TimeInVeryLowRecords int      `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`

	TimeInHighPercent *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`
	TimeInHighRecords int      `json:"timeInHighRecords" bson:"timeInHighRecords"`

	TimeInVeryHighPercent *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`
	TimeInVeryHighRecords int      `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
}

type BGMSummary struct {
	Periods     map[string]*BGMPeriod `json:"periods" bson:"periods"`
	HourlyStats []*BGMStats           `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int                   `json:"totalHours" bson:"totalHours"`

	// date tracking
	HasLastUploadDate bool       `json:"hasLastUploadDate" bson:"hasLastUploadDate"`
	LastUploadDate    time.Time  `json:"lastUploadDate" bson:"lastUploadDate"`
	LastUpdatedDate   time.Time  `json:"lastUpdatedDate" bson:"lastUpdatedDate"`
	OutdatedSince     *time.Time `json:"outdatedSince" bson:"outdatedSince"`
	FirstData         time.Time  `json:"firstData" bson:"firstData"`
	LastData          *time.Time `json:"lastData" bson:"lastData"`
}

func NewBGMStats(date time.Time) *BGMStats {
	return &BGMStats{
		Date: date,

		TargetRecords:   0,
		LowRecords:      0,
		VeryLowRecords:  0,
		HighRecords:     0,
		VeryHighRecords: 0,

		TotalGlucose: 0,
	}
}

func (userSummary *Summary) AddBGMStats(stats *BGMStats) error {
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
	if len(userSummary.BGM.HourlyStats) > 0 {
		for i := len(userSummary.BGM.HourlyStats) - 1; i >= 0; i-- {
			if userSummary.BGM.HourlyStats[i].Date.Equal(stats.Date) {
				userSummary.BGM.HourlyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if userSummary.BGM.HourlyStats[i].Date.After(stats.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(stats.Date.Sub(userSummary.BGM.HourlyStats[len(userSummary.BGM.HourlyStats)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = stats.Date.Add(time.Duration(-i) * time.Hour)
			userSummary.BGM.HourlyStats = append(userSummary.BGM.HourlyStats, NewBGMStats(newStatsTime))
		}
	}

	if existingDay == false {
		userSummary.BGM.HourlyStats = append(userSummary.BGM.HourlyStats, stats)
	}

	// remove extra days to cap at X days of stats
	hourCount = len(userSummary.BGM.HourlyStats)
	if hourCount > hoursAgoToKeep {
		userSummary.BGM.HourlyStats = userSummary.BGM.HourlyStats[hourCount-hoursAgoToKeep:]
	}

	// remove any stats that are older than X days from the last stat
	oldestHour = userSummary.BGM.HourlyStats[0].Date
	oldestHourToKeep = stats.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(userSummary.BGM.HourlyStats) - 2; i >= 0; i-- {
			if userSummary.BGM.HourlyStats[i].Date.Before(oldestHourToKeep) {
				userSummary.BGM.HourlyStats = userSummary.BGM.HourlyStats[i+1:]
				break
			}
		}
	}

	return nil
}

func (userSummary *Summary) CalculateBGMStats(userData []*glucoseDatum.Glucose) error {
	var normalizedValue float64
	var recordTime time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var stats *BGMStats

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	// skip past data
	if len(userSummary.BGM.HourlyStats) > 0 {
		userData, err = SkipUntil(userSummary.BGM.HourlyStats[len(userSummary.BGM.HourlyStats)-1].Date, userData)
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
			err = userSummary.AddBGMStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(userSummary.BGM.HourlyStats) > 0 {
				for i := len(userSummary.BGM.HourlyStats) - 1; i >= 0; i-- {
					if userSummary.BGM.HourlyStats[i].Date.Equal(currentHour) {
						stats = userSummary.BGM.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if userSummary.BGM.HourlyStats[i].Date.After(currentHour) {
						break
					}
				}
			}

			if stats == nil {
				stats = NewBGMStats(currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(userSummary.BGM.HourlyStats) > 0 {
			stats.LastRecordTime = userSummary.BGM.HourlyStats[len(userSummary.BGM.HourlyStats)-1].LastRecordTime
		}

		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(summaryGlucoseUnits))

		if normalizedValue <= veryLowBloodGlucose {
			stats.VeryLowRecords++
		} else if normalizedValue >= veryHighBloodGlucose {
			stats.VeryHighRecords++
		} else if normalizedValue <= lowBloodGlucose {
			stats.LowRecords++
		} else if normalizedValue >= highBloodGlucose {
			stats.HighRecords++
		} else {
			stats.TargetRecords++
		}

		stats.TotalRecords++
		stats.TotalGlucose += normalizedValue
		stats.LastRecordTime = recordTime
	}
	// store
	err = userSummary.AddBGMStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (userSummary *Summary) CalculateBGMSummary() {
	totalStats := NewBGMStats(time.Time{})

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(userSummary.BGM.HourlyStats); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			userSummary.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(userSummary.BGM.HourlyStats) - 1 - i
		totalStats.TargetRecords += userSummary.BGM.HourlyStats[currentIndex].TargetRecords
		totalStats.LowRecords += userSummary.BGM.HourlyStats[currentIndex].LowRecords
		totalStats.VeryLowRecords += userSummary.BGM.HourlyStats[currentIndex].VeryLowRecords
		totalStats.HighRecords += userSummary.BGM.HourlyStats[currentIndex].HighRecords
		totalStats.VeryHighRecords += userSummary.BGM.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += userSummary.BGM.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalRecords += userSummary.BGM.HourlyStats[currentIndex].TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		userSummary.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (userSummary *Summary) CalculatePeriod(i int, totalStats *BGMStats) {
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var averageGlucose *Glucose

	// remove partial hour (data end) from total time for more accurate TimeBGMUse
	totalMinutes := float64(i * 24 * 60)
	lastRecordTime := userSummary.BGM.HourlyStats[len(userSummary.BGM.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	userSummary.BGM.LastData = &lastRecordTime
	userSummary.BGM.FirstData = userSummary.BGM.HourlyStats[0].Date

	userSummary.BGM.TotalHours = len(userSummary.BGM.HourlyStats)

	if totalStats.TotalRecords != 0 {
		timeInTargetPercent = pointer.FromFloat64(float64(totalStats.TargetRecords) / float64(totalStats.TotalRecords))
		timeInLowPercent = pointer.FromFloat64(float64(totalStats.LowRecords) / float64(totalStats.TotalRecords))
		timeInVeryLowPercent = pointer.FromFloat64(float64(totalStats.VeryLowRecords) / float64(totalStats.TotalRecords))
		timeInHighPercent = pointer.FromFloat64(float64(totalStats.HighRecords) / float64(totalStats.TotalRecords))
		timeInVeryHighPercent = pointer.FromFloat64(float64(totalStats.VeryHighRecords) / float64(totalStats.TotalRecords))

		averageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalRecords),
			Units: summaryGlucoseUnits,
		}
	}

	// ensure periods exists, just in case
	if userSummary.BGM.Periods == nil {
		userSummary.BGM.Periods = make(map[string]*BGMPeriod)
	}

	userSummary.BGM.Periods[strconv.Itoa(i)+"d"] = &BGMPeriod{
		HasAverageGlucose:        averageGlucose != nil,
		HasTimeInTargetPercent:   timeInTargetPercent != nil,
		HasTimeInLowPercent:      timeInLowPercent != nil,
		HasTimeInVeryLowPercent:  timeInVeryLowPercent != nil,
		HasTimeInHighPercent:     timeInHighPercent != nil,
		HasTimeInVeryHighPercent: timeInVeryHighPercent != nil,

		AverageGlucose: averageGlucose,

		TimeInTargetPercent: timeInTargetPercent,
		TimeInTargetRecords: totalStats.TargetRecords,

		TimeInLowPercent: timeInLowPercent,
		TimeInLowRecords: totalStats.LowRecords,

		TimeInVeryLowPercent: timeInVeryLowPercent,
		TimeInVeryLowRecords: totalStats.VeryLowRecords,

		TimeInHighPercent: timeInHighPercent,
		TimeInHighRecords: totalStats.HighRecords,

		TimeInVeryHighPercent: timeInVeryHighPercent,
		TimeInVeryHighRecords: totalStats.VeryHighRecords,
	}
}

func (userSummary *Summary) UpdateBGM(ctx context.Context, status *UserBGMLastUpdated, userBGMData []*glucoseDatum.Glucose) error {
	var err error
	logger := log.LoggerFromContext(ctx)

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	userSummary.BGM.LastUpdatedDate = timestamp
	userSummary.BGM.OutdatedSince = nil
	userSummary.BGM.LastUploadDate = status.LastUpload

	// technically, this never could be zero, but we check anyway
	userSummary.BGM.HasLastUploadDate = !status.LastUpload.IsZero()

	// remove any past values that squeeze through the string date query that feeds this function
	// this mostly occurs when different sources use different time precisions (s vs ms vs ns)
	// resulting in $gt 00:00:01.275Z pulling in 00:00:01Z, which is before.
	if userSummary.BGM.LastData != nil {
		userBGMData, err = SkipUntil(*userSummary.BGM.LastData, userBGMData)
		if err != nil {
			return err
		}
	}

	// don't recalculate if there is no new data/this was double called
	if len(userBGMData) < 1 {
		logger.Debugf("No new records for userid %v summary calculation, aborting.", userSummary.UserID)
		return nil
	}

	err = userSummary.CalculateBGMStats(userBGMData)
	if err != nil {
		return err
	}

	userSummary.CalculateBGMSummary()

	return nil
}
