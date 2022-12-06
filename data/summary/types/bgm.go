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

type BGMHourlyStats []BGMHourlyStat

func (s BGMHourlyStat) GetDate() time.Time {
	return s.Date
}

func (s BGMHourlyStat) SetDate(t time.Time) {
	s.Date = t
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

type BGMPeriods map[string]BGMPeriod

type BGMStats struct {
	Periods     BGMPeriods     `json:"periods" bson:"periods"`
	HourlyStats BGMHourlyStats `json:"hourlyStats" bson:"hourlyStats"`
	TotalHours  int            `json:"totalHours" bson:"totalHours"`
}

func (BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (s BGMStats) Init() {
	s.HourlyStats = make([]BGMHourlyStat, 0)
	s.Periods = make(map[string]BGMPeriod)
	s.TotalHours = 0
}

func (s BGMStats) CalculateStats(userDataInterface interface{}) error {
	userData := userDataInterface.([]*glucoseDatum.Glucose)
	var normalizedValue float64
	var recordTime time.Time
	var lastHour time.Time
	var currentHour time.Time
	var err error
	var newStat *BGMHourlyStat

	for _, r := range userData {
		recordTime = *r.Time
		if err != nil {
			return errors.Wrap(err, "cannot parse time in record")
		}

		// truncate time is not timezone/DST safe here, even if we do expect UTC
		currentHour = time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(),
			recordTime.Hour(), 0, 0, 0, recordTime.Location())

		// store newStat for the day, if we are now on the next day
		if !lastHour.IsZero() && !currentHour.Equal(lastHour) {
			err = AddStats(s.HourlyStats, *newStat)
			if err != nil {
				return err
			}
			newStat = nil
		}

		if newStat == nil {
			// pull newStat if they already exist
			// NOTE we search the entire list, not just the last entry, in case we are given backfilled data
			if len(s.HourlyStats) > 0 {
				for i := len(s.HourlyStats) - 1; i >= 0; i-- {
					if s.HourlyStats[i].Date.Equal(currentHour) {
						newStat = &s.HourlyStats[i]
						break
					}

					// we already passed our date, give up
					if s.HourlyStats[i].Date.After(currentHour) {
						break
					}
				}
			}

			if newStat == nil {
				newStat = CreateHourlyStat[BGMHourlyStat](currentHour)
			}
		}

		lastHour = currentHour

		// if on fresh day, pull LastRecordTime from last day if possible
		if newStat.LastRecordTime.IsZero() && len(s.HourlyStats) > 0 {
			newStat.LastRecordTime = s.HourlyStats[len(s.HourlyStats)-1].LastRecordTime
		}

		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(summaryGlucoseUnits))

		if normalizedValue <= veryLowBloodGlucose {
			newStat.VeryLowRecords++
		} else if normalizedValue >= veryHighBloodGlucose {
			newStat.VeryHighRecords++
		} else if normalizedValue <= lowBloodGlucose {
			newStat.LowRecords++
		} else if normalizedValue >= highBloodGlucose {
			newStat.HighRecords++
		} else {
			newStat.TargetRecords++
		}

		newStat.TotalRecords++
		newStat.TotalGlucose += normalizedValue
		newStat.LastRecordTime = recordTime
	}

	// store
	err = AddStats(s.HourlyStats, *newStat)
	if err != nil {
		return err
	}

	return nil
}

func (s BGMStats) CalculateSummary() {
	totalStats := CreateHourlyStat[BGMHourlyStat](time.Time{})

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
		s.Periods = make(map[string]BGMPeriod)
	}

	s.Periods[strconv.Itoa(i)+"d"] = BGMPeriod{
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
