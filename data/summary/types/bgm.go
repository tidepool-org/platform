package types

import (
	"errors"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"

	"github.com/tidepool-org/platform/data/blood/glucose"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

type BGMBucketData struct {
	TargetRecords   int `json:"targetRecords" bson:"targetRecords"`
	AverageReadings int `json:"averageReadings" bson:"averageReadings"`
	LowRecords      int `json:"lowRecords" bson:"lowRecords"`
	VeryLowRecords  int `json:"veryLowRecords" bson:"veryLowRecords"`
	HighRecords     int `json:"highRecords" bson:"highRecords"`
	VeryHighRecords int `json:"veryHighRecords" bson:"veryHighRecords"`

	TotalGlucose float64 `json:"totalGlucose" bson:"totalGlucose"`
	TotalRecords int     `json:"totalRecords" bson:"totalRecords"`
}

type BGMPeriod struct {
	HasAverageGlucose bool     `json:"hasAverageGlucose" bson:"hasAverageGlucose"`
	AverageGlucose    *Glucose `json:"averageGlucose" bson:"averageGlucose"`

	HasTotalRecords bool `json:"hasTotalRecords" bson:"hasTotalRecords"`
	TotalRecords    *int `json:"totalRecords" bson:"totalRecords"`

	HasAverageDailyRecords bool     `json:"hasAverageDailyRecords" bson:"hasAverageDailyRecords"`
	AverageDailyRecords    *float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`

	HasTimeInTargetPercent bool     `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	TimeInTargetPercent    *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`

	HasTimeInTargetRecords bool `json:"hasTimeInTargetRecords" bson:"hasTimeInTargetRecords"`
	TimeInTargetRecords    *int `json:"timeInTargetRecords" bson:"timeInTargetRecords"`

	HasTimeInLowPercent bool     `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	TimeInLowPercent    *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`

	HasTimeInLowRecords bool `json:"hasTimeInLowRecords" bson:"hasTimeInLowRecords"`
	TimeInLowRecords    *int `json:"timeInLowRecords" bson:"timeInLowRecords"`

	HasTimeInVeryLowPercent bool     `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`
	TimeInVeryLowPercent    *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`

	HasTimeInVeryLowRecords bool `json:"hasTimeInVeryLowRecords" bson:"hasTimeInVeryLowRecords"`
	TimeInVeryLowRecords    *int `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`

	HasTimeInHighPercent bool     `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	TimeInHighPercent    *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`

	HasTimeInHighRecords bool `json:"hasTimeInHighRecords" bson:"hasTimeInHighRecords"`
	TimeInHighRecords    *int `json:"timeInHighRecords" bson:"timeInHighRecords"`

	HasTimeInVeryHighPercent bool     `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	TimeInVeryHighPercent    *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`

	HasTimeInVeryHighRecords bool `json:"hasTimeInVeryHighRecords" bson:"hasTimeInVeryHighRecords"`
	TimeInVeryHighRecords    *int `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
}

type BGMPeriods map[string]*BGMPeriod

type BGMStats struct {
	Periods    BGMPeriods                             `json:"periods" bson:"periods"`
	Buckets    Buckets[BGMBucketData, *BGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours int                                    `json:"totalHours" bson:"totalHours"`
}

func (*BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (*BGMStats) GetDeviceDataType() string {
	return selfmonitored.Type
}

func (s *BGMStats) Init() {
	s.Buckets = make(Buckets[BGMBucketData, *BGMBucketData], 0)
	s.Periods = make(map[string]*BGMPeriod)
	s.TotalHours = 0
}

func (s *BGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *BGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *BGMStats) Update(userData any) error {
	userDataTyped, ok := userData.([]*glucoseDatum.Glucose)
	if !ok {
		return errors.New("BGM records for calculation is not compatible with Glucose type")
	}

	err := AddData(&s.Buckets, userDataTyped)
	if err != nil {
		return err
	}

	s.CalculateSummary()

	return nil
}

func (B *BGMBucketData) CalculateStats(r any, _ *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("BGM record for calculation is not compatible with Glucose type")
	}

	normalizedValue := *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromAny(glucose.MmolL))

	if normalizedValue < veryLowBloodGlucose {
		B.VeryLowRecords++
	} else if normalizedValue > veryHighBloodGlucose {
		B.VeryHighRecords++
	} else if normalizedValue < lowBloodGlucose {
		B.LowRecords++
	} else if normalizedValue > highBloodGlucose {
		B.HighRecords++
	} else {
		B.TargetRecords++
	}

	B.TotalRecords++
	B.TotalGlucose += normalizedValue

	return false, nil
}

func (s *BGMStats) CalculateSummary() {
	// count backwards through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	var nextStopPoint int
	var totalStats = &BGMBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		if i == stopPoints[nextStopPoint]*24 {
			s.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
			nextStopPoint++
		}

		currentIndex := len(s.Buckets) - 1 - i
		totalStats.TargetRecords += s.Buckets[currentIndex].Data.TargetRecords
		totalStats.LowRecords += s.Buckets[currentIndex].Data.LowRecords
		totalStats.VeryLowRecords += s.Buckets[currentIndex].Data.VeryLowRecords
		totalStats.HighRecords += s.Buckets[currentIndex].Data.HighRecords
		totalStats.VeryHighRecords += s.Buckets[currentIndex].Data.VeryHighRecords

		totalStats.TotalGlucose += s.Buckets[currentIndex].Data.TotalGlucose
		totalStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], totalStats)
	}

	s.TotalHours = len(s.Buckets)
}

func (s *BGMStats) CalculatePeriod(i int, totalStats *BGMBucketData) {
	newPeriod := &BGMPeriod{
		HasTotalRecords: true,
		TotalRecords:    pointer.FromAny(totalStats.TotalRecords),

		HasAverageDailyRecords: true,
		AverageDailyRecords:    pointer.FromAny(float64(totalStats.TotalRecords) / float64(i)),

		HasTimeInTargetRecords: true,
		TimeInTargetRecords:    pointer.FromAny(totalStats.TargetRecords),

		HasTimeInLowRecords: true,
		TimeInLowRecords:    pointer.FromAny(totalStats.LowRecords),

		HasTimeInVeryLowRecords: true,
		TimeInVeryLowRecords:    pointer.FromAny(totalStats.VeryLowRecords),

		HasTimeInHighRecords: true,
		TimeInHighRecords:    pointer.FromAny(totalStats.HighRecords),

		HasTimeInVeryHighRecords: true,
		TimeInVeryHighRecords:    pointer.FromAny(totalStats.VeryHighRecords),
	}

	if totalStats.TotalRecords != 0 {
		newPeriod.HasTimeInTargetPercent = true
		newPeriod.TimeInTargetPercent = pointer.FromAny(float64(totalStats.TargetRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInLowPercent = true
		newPeriod.TimeInLowPercent = pointer.FromAny(float64(totalStats.LowRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInVeryLowPercent = true
		newPeriod.TimeInVeryLowPercent = pointer.FromAny(float64(totalStats.VeryLowRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInHighPercent = true
		newPeriod.TimeInHighPercent = pointer.FromAny(float64(totalStats.HighRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInVeryHighPercent = true
		newPeriod.TimeInVeryHighPercent = pointer.FromAny(float64(totalStats.VeryHighRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasAverageGlucose = true
		newPeriod.AverageGlucose = &Glucose{
			Value: totalStats.TotalGlucose / float64(totalStats.TotalRecords),
			Units: glucose.MmolL,
		}
	}

	s.Periods[strconv.Itoa(i)+"d"] = newPeriod
}
