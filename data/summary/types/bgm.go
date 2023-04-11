package types

import (
	"errors"
	"strconv"
	"time"

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

type BGMPeriods map[string]BGMPeriod

type BGMStats struct {
	Periods    BGMPeriods                             `json:"periods" bson:"periods"`
	Buckets    Buckets[BGMBucketData, *BGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours int                                    `json:"totalHours" bson:"totalHours"`
}

func (*BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (*BGMStats) GetDeviceDataType() string {
	return DeviceDataTypeBGM
}

func (s *BGMStats) Init() {
	s.Buckets = make(Buckets[BGMBucketData, *BGMBucketData], 0)
	s.Periods = make(map[string]BGMPeriod)
	s.TotalHours = 0
}

func (s *BGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *BGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *BGMStats) Update(userData any) error {
	var err error
	userDataTyped := userData.([]*glucoseDatum.Glucose)
	err = AddData(&s.Buckets, userDataTyped)
	if err != nil {
		return err
	}

	s.CalculateSummary()

	return nil
}

func (B *BGMBucketData) CalculateStats(r interface{}, _ *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("BGM record for calculation is not compatible with Glucose type")
	}
	var normalizedValue float64

	normalizedValue = *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromString(summaryGlucoseUnits))

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
	var totalStats = &BGMBucketData{}

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
}

func (s *BGMStats) CalculatePeriod(i int, totalStats *BGMBucketData) {
	var timeInTargetPercent *float64
	var timeInLowPercent *float64
	var timeInVeryLowPercent *float64
	var timeInHighPercent *float64
	var timeInVeryHighPercent *float64
	var averageGlucose *Glucose

	// remove partial hour (data end) from total time for more accurate TimeBGMUse
	totalMinutes := float64(i * 24 * 60)
	lastRecordTime := s.Buckets[len(s.Buckets)-1].LastRecordTime
	nextHour := time.Date(lastRecordTime.Year(), lastRecordTime.Month(), lastRecordTime.Day(),
		lastRecordTime.Hour()+1, 0, 0, 0, lastRecordTime.Location())
	totalMinutes = totalMinutes - nextHour.Sub(lastRecordTime).Minutes()

	s.TotalHours = len(s.Buckets)

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

		HasAverageGlucose: averageGlucose != nil,
		AverageGlucose:    averageGlucose,

		HasTotalRecords: true,
		TotalRecords:    pointer.FromAny(totalStats.TotalRecords),

		HasAverageDailyRecords: true,
		AverageDailyRecords:    pointer.FromAny(float64(totalStats.TotalRecords) / float64(i)),

		HasTimeInTargetPercent: timeInTargetPercent != nil,
		TimeInTargetPercent:    timeInTargetPercent,

		HasTimeInTargetRecords: true,
		TimeInTargetRecords:    pointer.FromAny(totalStats.TargetRecords),

		HasTimeInLowPercent: timeInLowPercent != nil,
		TimeInLowPercent:    timeInLowPercent,

		HasTimeInLowRecords: true,
		TimeInLowRecords:    pointer.FromAny(totalStats.LowRecords),

		HasTimeInVeryLowPercent: timeInVeryLowPercent != nil,
		TimeInVeryLowPercent:    timeInVeryLowPercent,

		HasTimeInVeryLowRecords: true,
		TimeInVeryLowRecords:    pointer.FromAny(totalStats.VeryLowRecords),

		HasTimeInHighPercent: timeInHighPercent != nil,
		TimeInHighPercent:    timeInHighPercent,

		HasTimeInHighRecords: true,
		TimeInHighRecords:    pointer.FromAny(totalStats.HighRecords),

		HasTimeInVeryHighPercent: timeInVeryHighPercent != nil,
		TimeInVeryHighPercent:    timeInVeryHighPercent,

		HasTimeInVeryHighRecords: true,
		TimeInVeryHighRecords:    pointer.FromAny(totalStats.VeryHighRecords),
	}
}
