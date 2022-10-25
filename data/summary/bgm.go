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
	Summary `json:",inline" bson:",inline"`

	Periods     map[string]*BGMPeriod `json:"periods" bson:"periods"`
	HourlyStats []*BGMStats           `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int                   `json:"totalHours" bson:"totalHours"`
}

func NewBGMSummary(id string) *BGMSummary {
	return &BGMSummary{
		Summary: Summary{
			UserID: id,
			Type:   "bgm",

			HasLastUploadDate: false,
			LastUploadDate:    time.Time{},
			LastUpdatedDate:   time.Time{},
			FirstData:         time.Time{},
			LastData:          nil,
			OutdatedSince:     nil,

			Config: Config{
				SchemaVersion:            1,
				HighGlucoseThreshold:     highBloodGlucose,
				VeryHighGlucoseThreshold: veryHighBloodGlucose,
				LowGlucoseThreshold:      lowBloodGlucose,
				VeryLowGlucoseThreshold:  veryLowBloodGlucose,
			},
		},
		Periods:     make(map[string]*BGMPeriod),
		HourlyStats: make([]*BGMStats, 0),
		TotalHours:  0,
	}
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

func (summaryData *BGMSummary) AddStats(stats *BGMStats) error {
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
	if len(summaryData.HourlyStats) > 0 {
		for i := len(summaryData.HourlyStats) - 1; i >= 0; i-- {
			if summaryData.HourlyStats[i].Date.Equal(stats.Date) {
				summaryData.HourlyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if summaryData.HourlyStats[i].Date.After(stats.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(stats.Date.Sub(summaryData.HourlyStats[len(summaryData.HourlyStats)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = stats.Date.Add(time.Duration(-i) * time.Hour)
			summaryData.HourlyStats = append(summaryData.HourlyStats, NewBGMStats(newStatsTime))
		}
	}

	if existingDay == false {
		summaryData.HourlyStats = append(summaryData.HourlyStats, stats)
	}

	// remove extra days to cap at X days of stats
	hourCount = len(summaryData.HourlyStats)
	if hourCount > hoursAgoToKeep {
		summaryData.HourlyStats = summaryData.HourlyStats[hourCount-hoursAgoToKeep:]
	}

	// remove any stats that are older than X days from the last stat
	oldestHour = summaryData.HourlyStats[0].Date
	oldestHourToKeep = stats.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(summaryData.HourlyStats) - 2; i >= 0; i-- {
			if summaryData.HourlyStats[i].Date.Before(oldestHourToKeep) {
				summaryData.HourlyStats = summaryData.HourlyStats[i+1:]
				break
			}
		}
	}

	return nil
}

func (summaryData *BGMSummary) CalculateStats(userData []*glucoseDatum.Glucose) error {
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
	if len(summaryData.HourlyStats) > 0 {
		userData, err = SkipUntil(summaryData.HourlyStats[len(summaryData.HourlyStats)-1].Date, userData)
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
			err = summaryData.AddStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(summaryData.HourlyStats) > 0 {
				for i := len(summaryData.HourlyStats) - 1; i >= 0; i-- {
					if summaryData.HourlyStats[i].Date.Equal(currentHour) {
						stats = summaryData.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if summaryData.HourlyStats[i].Date.After(currentHour) {
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
		if stats.LastRecordTime.IsZero() && len(summaryData.HourlyStats) > 0 {
			stats.LastRecordTime = summaryData.HourlyStats[len(summaryData.HourlyStats)-1].LastRecordTime
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
	err = summaryData.AddStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (summaryData *BGMSummary) CalculateSummary() {
	totalStats := NewBGMStats(time.Time{})

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(summaryData.HourlyStats); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			summaryData.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(summaryData.HourlyStats) - 1 - i
		totalStats.TargetRecords += summaryData.HourlyStats[currentIndex].TargetRecords
		totalStats.LowRecords += summaryData.HourlyStats[currentIndex].LowRecords
		totalStats.VeryLowRecords += summaryData.HourlyStats[currentIndex].VeryLowRecords
		totalStats.HighRecords += summaryData.HourlyStats[currentIndex].HighRecords
		totalStats.VeryHighRecords += summaryData.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += summaryData.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalRecords += summaryData.HourlyStats[currentIndex].TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		summaryData.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (summaryData *BGMSummary) CalculatePeriod(i int, totalStats *BGMStats) {
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var averageGlucose *Glucose

	// remove partial hour (data end) from total time for more accurate TimeBGMUse
	totalMinutes := float64(i * 24 * 60)
	lastRecordTime := summaryData.HourlyStats[len(summaryData.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	summaryData.LastData = &lastRecordTime
	summaryData.FirstData = summaryData.HourlyStats[0].Date

	summaryData.TotalHours = len(summaryData.HourlyStats)

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
	if summaryData.Periods == nil {
		summaryData.Periods = make(map[string]*BGMPeriod)
	}

	summaryData.Periods[strconv.Itoa(i)+"d"] = &BGMPeriod{
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

func (summaryData *BGMSummary) Update(ctx context.Context, status *UserLastUpdated, userBGMData []*glucoseDatum.Glucose) error {
	var err error
	logger := log.LoggerFromContext(ctx)

	// prepare state of existing summary
	timestamp := time.Now().UTC()
	summaryData.LastUpdatedDate = timestamp
	summaryData.OutdatedSince = nil
	summaryData.LastUploadDate = status.LastUpload

	// technically, this never could be zero, but we check anyway
	summaryData.HasLastUploadDate = !status.LastUpload.IsZero()

	// remove any past values that squeeze through the string date query that feeds this function
	// this mostly occurs when different sources use different time precisions (s vs ms vs ns)
	// resulting in $gt 00:00:01.275Z pulling in 00:00:01Z, which is before.
	if summaryData.LastData != nil {
		userBGMData, err = SkipUntil(*summaryData.LastData, userBGMData)
		if err != nil {
			return err
		}
	}

	// don't recalculate if there is no new data/this was double called
	if len(userBGMData) < 1 {
		logger.Debugf("No new records for userid %v summary calculation, aborting.", summaryData.UserID)
		return nil
	}

	err = summaryData.CalculateStats(userBGMData)
	if err != nil {
		return err
	}

	summaryData.CalculateSummary()

	return nil
}
