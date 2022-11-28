package types

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"strconv"
	"time"
)

type BGMHourlyStat struct {
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

func NewBGMHourlyStat(date time.Time) *BGMHourlyStat {
	return &BGMHourlyStat{
		Date: date,

		TargetRecords:   0,
		LowRecords:      0,
		VeryLowRecords:  0,
		HighRecords:     0,
		VeryHighRecords: 0,

		TotalGlucose: 0,
	}
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

type BGMStats struct {
	Periods     map[string]*BGMPeriod `json:"periods" bson:"periods"`
	HourlyStats []*BGMHourlyStat      `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int                   `json:"totalHours" bson:"totalHours"`
}

func (BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (s BGMStats) PopulateStats() {
	s.Periods = make(map[string]*BGMPeriod)
	s.HourlyStats = make([]*BGMHourlyStat, 0)
	s.TotalHours = 0
}

func (s BGMStats) AddStats(stats *BGMHourlyStat) error {
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
	if len(s.HourlyStats) > 0 {
		for i := len(s.HourlyStats) - 1; i >= 0; i-- {
			if s.HourlyStats[i].Date.Equal(stats.Date) {
				s.HourlyStats[i] = stats
				existingDay = true
				break
			}

			// we already passed our date, give up
			if s.HourlyStats[i].Date.After(stats.Date) {
				break
			}
		}

		// add hours for any gaps that this new stat skipped
		statsGap = int(stats.Date.Sub(s.HourlyStats[len(s.HourlyStats)-1].Date).Hours())
		for i := statsGap; i > 1; i-- {
			newStatsTime = stats.Date.Add(time.Duration(-i) * time.Hour)
			s.HourlyStats = append(s.HourlyStats, NewBGMHourlyStat(newStatsTime))
		}
	}

	if existingDay == false {
		s.HourlyStats = append(s.HourlyStats, stats)
	}

	// remove extra days to cap at X days of stats
	hourCount = len(s.HourlyStats)
	if hourCount > hoursAgoToKeep {
		s.HourlyStats = s.HourlyStats[hourCount-hoursAgoToKeep:]
	}

	// remove any stats that are older than X days from the last stat
	oldestHour = s.HourlyStats[0].Date
	oldestHourToKeep = stats.Date.Add(-hoursAgoToKeep * time.Hour)
	if oldestHour.Before(oldestHourToKeep) {
		// we don't check the last entry because we just added/updated it
		for i := len(s.HourlyStats) - 2; i >= 0; i-- {
			if s.HourlyStats[i].Date.Before(oldestHourToKeep) {
				s.HourlyStats = s.HourlyStats[i+1:]
				break
			}
		}
	}

	return nil
}

func (s BGMStats) CalculateStats(userData []*glucoseDatum.Glucose) error {
	var normalizedValue float64
	var recordTime time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var stats *BGMHourlyStat

	if len(userData) < 1 {
		return errors.New("userData is empty, nothing to calculate stats for")
	}

	// skip past data
	if len(s.HourlyStats) > 0 {
		userData, err = SkipUntil(s.HourlyStats[len(s.HourlyStats)-1].Date, userData)
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
			err = s.AddStats(stats)
			if err != nil {
				return err
			}
			stats = nil
		}

		if stats == nil {
			// pull stats if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(s.HourlyStats) > 0 {
				for i := len(s.HourlyStats) - 1; i >= 0; i-- {
					if s.HourlyStats[i].Date.Equal(currentHour) {
						stats = s.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if s.HourlyStats[i].Date.After(currentHour) {
						break
					}
				}
			}

			if stats == nil {
				stats = NewBGMHourlyStat(currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(s.HourlyStats) > 0 {
			stats.LastRecordTime = s.HourlyStats[len(s.HourlyStats)-1].LastRecordTime
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
	err = s.AddStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (s BGMStats) CalculateSummary() {
	totalStats := NewBGMHourlyStat(time.Time{})

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(s.HourlyStats); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			s.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(s.HourlyStats) - 1 - i
		totalStats.TargetRecords += s.HourlyStats[currentIndex].TargetRecords
		totalStats.LowRecords += s.HourlyStats[currentIndex].LowRecords
		totalStats.VeryLowRecords += s.HourlyStats[currentIndex].VeryLowRecords
		totalStats.HighRecords += s.HourlyStats[currentIndex].HighRecords
		totalStats.VeryHighRecords += s.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += s.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalRecords += s.HourlyStats[currentIndex].TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (s BGMStats) CalculatePeriod(i int, totalStats *BGMHourlyStat) {
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var averageGlucose *Glucose

	// remove partial hour (data end) from total time for more accurate TimeBGMUse
	totalMinutes := float64(i * 24 * 60)
	lastRecordTime := s.HourlyStats[len(s.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	s.TotalHours = len(s.HourlyStats)

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
	if s.Periods == nil {
		s.Periods = make(map[string]*BGMPeriod)
	}

	s.Periods[strconv.Itoa(i)+"d"] = &BGMPeriod{
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
