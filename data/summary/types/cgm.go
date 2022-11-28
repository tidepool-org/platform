package types

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"math"
	"strconv"
	"strings"
	"time"
)

type CGMHourlyStat struct {
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

func NewCGMHourlyStat(date time.Time) *CGMHourlyStat {
	return &CGMHourlyStat{
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

type CGMStats struct {
	Periods     map[string]*CGMPeriod `json:"periods" bson:"periods"`
	HourlyStats []*CGMHourlyStat      `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int                   `json:"totalHours" bson:"totalHours"`
}

func (CGMStats) GetType() string {
	return SummaryTypeCGM
}

// GetDuration assumes all except freestyle is 5 minutes
func GetDuration(dataSet *glucoseDatum.Glucose) int {
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

func (s CGMStats) PopulateStats() {
	s.Periods = make(map[string]*CGMPeriod)
	s.HourlyStats = make([]*CGMHourlyStat, 0)
	s.TotalHours = 0
}

func (s CGMStats) AddStats(stats *CGMHourlyStat) error {
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
			s.HourlyStats = append(s.HourlyStats, NewCGMHourlyStat(newStatsTime))
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

func (s CGMStats) CalculateStats(userData []*glucoseDatum.Glucose) error {
	var normalizedValue float64
	var duration int
	var recordTime time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var stats *CGMHourlyStat

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
				stats = NewCGMHourlyStat(currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if stats.LastRecordTime.IsZero() && len(s.HourlyStats) > 0 {
			stats.LastRecordTime = s.HourlyStats[len(s.HourlyStats)-1].LastRecordTime
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
	err = s.AddStats(stats)
	if err != nil {
		return err
	}

	return nil
}

func (s CGMStats) CalculateSummary() {
	totalStats := NewCGMHourlyStat(time.Time{})

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
		totalStats.TargetMinutes += s.HourlyStats[currentIndex].TargetMinutes
		totalStats.TargetRecords += s.HourlyStats[currentIndex].TargetRecords

		totalStats.LowMinutes += s.HourlyStats[currentIndex].LowMinutes
		totalStats.LowRecords += s.HourlyStats[currentIndex].LowRecords

		totalStats.VeryLowMinutes += s.HourlyStats[currentIndex].VeryLowMinutes
		totalStats.VeryLowRecords += s.HourlyStats[currentIndex].VeryLowRecords

		totalStats.HighMinutes += s.HourlyStats[currentIndex].HighMinutes
		totalStats.HighRecords += s.HourlyStats[currentIndex].HighRecords

		totalStats.VeryHighMinutes += s.HourlyStats[currentIndex].VeryHighMinutes
		totalStats.VeryHighRecords += s.HourlyStats[currentIndex].VeryHighRecords

		totalStats.TotalGlucose += s.HourlyStats[currentIndex].TotalGlucose
		totalStats.TotalMinutes += s.HourlyStats[currentIndex].TotalMinutes
		totalStats.TotalRecords += s.HourlyStats[currentIndex].TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (s CGMStats) CalculatePeriod(i int, totalStats *CGMHourlyStat) {
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
	lastRecordTime := s.HourlyStats[len(s.HourlyStats)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	// TODO move
	//s.LastData = &lastRecordTime
	//s.FirstData = s.HourlyStats[0].Date

	s.TotalHours = len(s.HourlyStats)

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
	if s.Periods == nil {
		s.Periods = make(map[string]*CGMPeriod)
	}

	s.Periods[strconv.Itoa(i)+"d"] = &CGMPeriod{
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
