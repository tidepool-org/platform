package types

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"strconv"
	"time"
)

type CGMBucketData struct {
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

type CGMPeriods map[string]CGMPeriod

type CGMStats struct {
	Periods    CGMPeriods                             `json:"periods" bson:"periods"`
	Buckets    Buckets[CGMBucketData, *CGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours int                                    `json:"totalHours" bson:"totalHours"`
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (s *CGMStats) Init() {
	s.Buckets = make(Buckets[CGMBucketData, *CGMBucketData], 0)
	s.Periods = make(map[string]CGMPeriod)
	s.TotalHours = 0
}

func (s *CGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *CGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *CGMStats) Update(userData any) error {
	userDataTyped := userData.([]*glucoseDatum.Glucose)
	err := AddData(s.Buckets, userDataTyped)
	if err != nil {
		return err
	}

	s.CalculateSummary()

	return nil
}

func (B *CGMBucketData) CalculateStats(r interface{}, lastRecordTime *time.Time) error {
	dataRecord := r.(*glucoseDatum.Glucose)
	var normalizedValue float64
	var duration int

	// duration has never been calculated, use current record's duration for this cycle
	if duration == 0 {
		duration = GetDuration(dataRecord)
	}

	// calculate blackoutWindow based on duration of previous value
	blackoutWindow := time.Duration(duration)*time.Minute - 3*time.Second

	// if we are too close to the previous value, skip
	if dataRecord.Time.Sub(*lastRecordTime) > blackoutWindow {
		normalizedValue = *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromString(summaryGlucoseUnits))
		duration = GetDuration(dataRecord)

		if normalizedValue < veryLowBloodGlucose {
			B.VeryLowMinutes += duration
			B.VeryLowRecords++
		} else if normalizedValue > veryHighBloodGlucose {
			B.VeryHighMinutes += duration
			B.VeryHighRecords++
		} else if normalizedValue < lowBloodGlucose {
			B.LowMinutes += duration
			B.LowRecords++
		} else if normalizedValue > highBloodGlucose {
			B.HighMinutes += duration
			B.HighRecords++
		} else {
			B.TargetMinutes += duration
			B.TargetRecords++
		}

		B.TotalMinutes += duration
		B.TotalRecords++
		B.TotalGlucose += normalizedValue * float64(duration)
	}

	return nil
}

func (s *CGMStats) CalculateSummary() {
	var totalStats *CGMBucketData
	s.TotalHours = len(s.Buckets)

	// ensure periods exists, just in case
	if s.Periods == nil {
		s.Periods = make(map[string]CGMPeriod)
	}

	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	stopPoints := []int{1, 7, 14, 30}
	var nextStopPoint int
	var currentIndex int

	for i := 0; i < len(s.Buckets); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			s.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex = len(s.Buckets) - 1 - i
		totalStats.TargetMinutes += s.Buckets[currentIndex].Data.TargetMinutes
		totalStats.TargetRecords += s.Buckets[currentIndex].Data.TargetRecords

		totalStats.LowMinutes += s.Buckets[currentIndex].Data.LowMinutes
		totalStats.LowRecords += s.Buckets[currentIndex].Data.LowRecords

		totalStats.VeryLowMinutes += s.Buckets[currentIndex].Data.VeryLowMinutes
		totalStats.VeryLowRecords += s.Buckets[currentIndex].Data.VeryLowRecords

		totalStats.HighMinutes += s.Buckets[currentIndex].Data.HighMinutes
		totalStats.HighRecords += s.Buckets[currentIndex].Data.HighRecords

		totalStats.VeryHighMinutes += s.Buckets[currentIndex].Data.VeryHighMinutes
		totalStats.VeryHighRecords += s.Buckets[currentIndex].Data.VeryHighRecords

		totalStats.TotalGlucose += s.Buckets[currentIndex].Data.TotalGlucose
		totalStats.TotalMinutes += s.Buckets[currentIndex].Data.TotalMinutes
		totalStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], totalStats)
	}
}

func (s *CGMStats) CalculatePeriod(i int, totalStats *CGMBucketData) {
	var timeCGMUsePercent *float64
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var glucoseManagementIndicator *float64
	var realMinutes float64
	var averageGlucose *Glucose

	if totalStats.TotalRecords != 0 {
		realMinutes = CalculateRealMinutes(i, s.Buckets[len(s.Buckets)-1].LastRecordTime)
		timeCGMUsePercent = pointer.FromFloat64(float64(totalStats.TotalMinutes) / realMinutes)
		// if we are storing under 1d, apply 70% rule to TimeIn*
		// if we are storing over 1d, check for 24h cgm use
		if (i <= 1 && *timeCGMUsePercent > 0.7) || (i > 1 && totalStats.TotalMinutes > 1440) {
			timeInTargetPercent = pointer.FromFloat64(float64(totalStats.TargetMinutes) / float64(totalStats.TotalMinutes))
			timeInLowPercent = pointer.FromFloat64(float64(totalStats.LowMinutes) / float64(totalStats.TotalMinutes))
			timeInVeryLowPercent = pointer.FromFloat64(float64(totalStats.VeryLowMinutes) / float64(totalStats.TotalMinutes))
			timeInHighPercent = pointer.FromFloat64(float64(totalStats.HighMinutes) / float64(totalStats.TotalMinutes))
			timeInVeryHighPercent = pointer.FromFloat64(float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalMinutes))
		}

		averageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalMinutes),
			Units: summaryGlucoseUnits,
		}

		// we only add GMI if cgm use >70%, otherwise clear it
		if *timeCGMUsePercent > 0.7 {
			glucoseManagementIndicator = pointer.FromFloat64(CalculateGMI(averageGlucose.Value))
		}
	}

	s.Periods[strconv.Itoa(i)+"d"] = CGMPeriod{
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
