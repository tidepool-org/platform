package types

import (
	"errors"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

type CGMBucketData struct {
	LastRecordDuration int `json:"LastRecordDuration" bson:"LastRecordDuration"`

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
	HasTimeCGMUsePercent bool     `json:"hasTimeCGMUsePercent" bson:"hasTimeCGMUsePercent"`
	TimeCGMUsePercent    *float64 `json:"timeCGMUsePercent" bson:"timeCGMUsePercent"`

	HasTimeCGMUseMinutes bool `json:"hasTimeCGMUseMinutes" bson:"hasTimeCGMUseMinutes"`
	TimeCGMUseMinutes    *int `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`

	HasTimeCGMUseRecords bool `json:"hasTimeCGMUseRecords" bson:"hasTimeCGMUseRecords"`
	TimeCGMUseRecords    *int `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`

	HasAverageGlucose bool     `json:"hasAverageGlucose" bson:"hasAverageGlucose"`
	AverageGlucose    *Glucose `json:"averageGlucose" bson:"averageGlucose"`

	HasGlucoseManagementIndicator bool     `json:"hasGlucoseManagementIndicator" bson:"hasGlucoseManagementIndicator"`
	GlucoseManagementIndicator    *float64 `json:"glucoseManagementIndicator" bson:"glucoseManagementIndicator"`

	HasTotalRecords bool `json:"hasTotalRecords" bson:"hasTotalRecords"`
	TotalRecords    *int `json:"totalRecords" bson:"totalRecords"`

	HasAverageDailyRecords bool     `json:"hasAverageDailyRecords" bson:"hasAverageDailyRecords"`
	AverageDailyRecords    *float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`

	HasTimeInTargetPercent bool     `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	TimeInTargetPercent    *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`

	HasTimeInTargetMinutes bool `json:"hasTimeInTargetMinutes" bson:"hasTimeInTargetMinutes"`
	TimeInTargetMinutes    *int `json:"timeInTargetMinutes" bson:"timeInTargetMinutes"`

	HasTimeInTargetRecords bool `json:"hasTimeInTargetRecords" bson:"hasTimeInTargetRecords"`
	TimeInTargetRecords    *int `json:"timeInTargetRecords" bson:"timeInTargetRecords"`

	HasTimeInLowPercent bool     `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	TimeInLowPercent    *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`

	HasTimeInLowMinutes bool `json:"hasTimeInLowMinutes" bson:"hasTimeInLowMinutes"`
	TimeInLowMinutes    *int `json:"timeInLowMinutes" bson:"timeInLowMinutes"`

	HasTimeInLowRecords bool `json:"hasTimeInLowRecords" bson:"hasTimeInLowRecords"`
	TimeInLowRecords    *int `json:"timeInLowRecords" bson:"timeInLowRecords"`

	HasTimeInVeryLowPercent bool     `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`
	TimeInVeryLowPercent    *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`

	HasTimeInVeryLowMinutes bool `json:"hasTimeInVeryLowMinutes" bson:"hasTimeInVeryLowMinutes"`
	TimeInVeryLowMinutes    *int `json:"timeInVeryLowMinutes" bson:"timeInVeryLowMinutes"`

	HasTimeInVeryLowRecords bool `json:"hasTimeInVeryLowRecords" bson:"hasTimeInVeryLowRecords"`
	TimeInVeryLowRecords    *int `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`

	HasTimeInHighPercent bool     `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	TimeInHighPercent    *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`

	HasTimeInHighMinutes bool `json:"hasTimeInHighMinutes" bson:"hasTimeInHighMinutes"`
	TimeInHighMinutes    *int `json:"timeInHighMinutes" bson:"timeInHighMinutes"`

	HasTimeInHighRecords bool `json:"hasTimeInHighRecords" bson:"hasTimeInHighRecords"`
	TimeInHighRecords    *int `json:"timeInHighRecords" bson:"timeInHighRecords"`

	HasTimeInVeryHighPercent bool     `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	TimeInVeryHighPercent    *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`

	HasTimeInVeryHighMinutes bool `json:"hasTimeInVeryHighMinutes" bson:"hasTimeInVeryHighMinutes"`
	TimeInVeryHighMinutes    *int `json:"timeInVeryHighMinutes" bson:"timeInVeryHighMinutes"`

	HasTimeInVeryHighRecords bool `json:"hasTimeInVeryHighRecords" bson:"hasTimeInVeryHighRecords"`
	TimeInVeryHighRecords    *int `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
}

type CGMPeriods map[string]*CGMPeriod

type CGMStats struct {
	Periods    CGMPeriods                             `json:"periods" bson:"periods"`
	Buckets    Buckets[CGMBucketData, *CGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours int                                    `json:"totalHours" bson:"totalHours"`
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataType() string {
	return DeviceDataTypeCGM
}

func (s *CGMStats) Init() {
	s.Buckets = make(Buckets[CGMBucketData, *CGMBucketData], 0)
	s.Periods = make(map[string]*CGMPeriod)
	s.TotalHours = 0
}

func (s *CGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *CGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *CGMStats) Update(userData any) error {
	userDataTyped, ok := userData.([]*glucoseDatum.Glucose)
	if !ok {
		return errors.New("CGM records for calculation is not compatible with Glucose type")
	}

	err := AddData(&s.Buckets, userDataTyped)
	if err != nil {
		return err
	}

	s.CalculateSummary()

	return nil
}

func (B *CGMBucketData) CalculateStats(r any, lastRecordTime *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("CGM record for calculation is not compatible with Glucose type")
	}

	// this is a new bucket, use current record as duration reference
	if B.LastRecordDuration == 0 {
		B.LastRecordDuration = GetDuration(dataRecord)
	}

	// calculate blackoutWindow based on duration of previous value
	blackoutWindow := time.Duration(B.LastRecordDuration)*time.Minute - 10*time.Second

	// Skip record unless we are beyond the blackout window
	if dataRecord.Time.Sub(*lastRecordTime) > blackoutWindow {
		normalizedValue := *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromAny(glucose.MmolL))
		duration := GetDuration(dataRecord)

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
		B.LastRecordDuration = duration

		return false, nil
	}

	return true, nil
}

func (s *CGMStats) CalculateSummary() {
	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	var nextStopPoint int
	var totalStats = &CGMBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			s.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex := len(s.Buckets) - 1 - i
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

	s.TotalHours = len(s.Buckets)
}

func (s *CGMStats) CalculatePeriod(i int, totalStats *CGMBucketData) {
	newPeriod := &CGMPeriod{
		HasTimeCGMUseMinutes: true,
		TimeCGMUseMinutes:    pointer.FromAny(totalStats.TotalMinutes),

		HasTimeCGMUseRecords: true,
		TimeCGMUseRecords:    pointer.FromAny(totalStats.TotalRecords),

		HasTotalRecords: true,
		TotalRecords:    pointer.FromAny(totalStats.TotalRecords),

		HasAverageDailyRecords: true,
		AverageDailyRecords:    pointer.FromAny(float64(totalStats.TotalRecords) / float64(i)),

		HasTimeInTargetMinutes: true,
		TimeInTargetMinutes:    pointer.FromAny(totalStats.TargetMinutes),

		HasTimeInTargetRecords: true,
		TimeInTargetRecords:    pointer.FromAny(totalStats.TargetRecords),

		HasTimeInLowMinutes: true,
		TimeInLowMinutes:    pointer.FromAny(totalStats.LowMinutes),

		HasTimeInLowRecords: true,
		TimeInLowRecords:    pointer.FromAny(totalStats.LowRecords),

		HasTimeInVeryLowMinutes: true,
		TimeInVeryLowMinutes:    pointer.FromAny(totalStats.VeryLowMinutes),

		HasTimeInVeryLowRecords: true,
		TimeInVeryLowRecords:    pointer.FromAny(totalStats.VeryLowRecords),

		HasTimeInHighMinutes: true,
		TimeInHighMinutes:    pointer.FromAny(totalStats.HighMinutes),

		HasTimeInHighRecords: true,
		TimeInHighRecords:    pointer.FromAny(totalStats.HighRecords),

		HasTimeInVeryHighMinutes: true,
		TimeInVeryHighMinutes:    pointer.FromAny(totalStats.VeryHighMinutes),

		HasTimeInVeryHighRecords: true,
		TimeInVeryHighRecords:    pointer.FromAny(totalStats.VeryHighRecords),
	}

	if totalStats.TotalRecords != 0 {
		realMinutes := CalculateRealMinutes(i, s.Buckets[len(s.Buckets)-1].LastRecordTime)
		newPeriod.HasTimeCGMUsePercent = true
		newPeriod.TimeCGMUsePercent = pointer.FromAny(float64(totalStats.TotalMinutes) / realMinutes)

		// if we are storing under 1d, apply 70% rule to TimeIn*
		// if we are storing over 1d, check for 24h cgm use
		if (i <= 1 && *newPeriod.TimeCGMUsePercent > 0.7) || (i > 1 && totalStats.TotalMinutes > 1440) {
			newPeriod.HasTimeInTargetPercent = true
			newPeriod.TimeInTargetPercent = pointer.FromAny(float64(totalStats.TargetMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInLowPercent = true
			newPeriod.TimeInLowPercent = pointer.FromAny(float64(totalStats.LowMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInVeryLowPercent = true
			newPeriod.TimeInVeryLowPercent = pointer.FromAny(float64(totalStats.VeryLowMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInHighPercent = true
			newPeriod.TimeInHighPercent = pointer.FromAny(float64(totalStats.HighMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInVeryHighPercent = true
			newPeriod.TimeInVeryHighPercent = pointer.FromAny(float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalMinutes))
		}

		newPeriod.HasAverageGlucose = true
		newPeriod.AverageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalMinutes),
			Units: glucose.MmolL,
		}

		// we only add GMI if cgm use >70%, otherwise clear it
		if *newPeriod.TimeCGMUsePercent > 0.7 {
			newPeriod.HasGlucoseManagementIndicator = true
			newPeriod.GlucoseManagementIndicator = pointer.FromAny(CalculateGMI(newPeriod.AverageGlucose.Value))
		}
	}

	s.Periods[strconv.Itoa(i)+"d"] = newPeriod
}
