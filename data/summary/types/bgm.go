package types

import (
	"context"
	"errors"
	"github.com/tidepool-org/platform/data/summary"
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
	HasAverageGlucoseMmol   bool     `json:"hasAverageGlucoseMmol" bson:"hasAverageGlucoseMmol"`
	AverageGlucoseMmol      *float64 `json:"averageGlucoseMmol" bson:"averageGlucoseMmol"`
	AverageGlucoseMmolDelta *float64 `json:"averageGlucoseMmolDelta" bson:"averageGlucoseMmolDelta"`

	HasTotalRecords   bool `json:"hasTotalRecords" bson:"hasTotalRecords"`
	TotalRecords      *int `json:"totalRecords" bson:"totalRecords"`
	TotalRecordsDelta *int `json:"totalRecordsDelta" bson:"totalRecordsDelta"`

	HasAverageDailyRecords   bool     `json:"hasAverageDailyRecords" bson:"hasAverageDailyRecords"`
	AverageDailyRecords      *float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`
	AverageDailyRecordsDelta *float64 `json:"averageDailyRecordsDelta" bson:"averageDailyRecordsDelta"`

	HasTimeInTargetPercent   bool     `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	TimeInTargetPercent      *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`
	TimeInTargetPercentDelta *float64 `json:"timeInTargetPercentDelta" bson:"timeInTargetPercentDelta"`

	HasTimeInTargetRecords   bool `json:"hasTimeInTargetRecords" bson:"hasTimeInTargetRecords"`
	TimeInTargetRecords      *int `json:"timeInTargetRecords" bson:"timeInTargetRecords"`
	TimeInTargetRecordsDelta *int `json:"timeInTargetRecordsDelta" bson:"timeInTargetRecordsDelta"`

	HasTimeInLowPercent   bool     `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	TimeInLowPercent      *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`
	TimeInLowPercentDelta *float64 `json:"timeInLowPercentDelta" bson:"timeInLowPercentDelta"`

	HasTimeInLowRecords   bool `json:"hasTimeInLowRecords" bson:"hasTimeInLowRecords"`
	TimeInLowRecords      *int `json:"timeInLowRecords" bson:"timeInLowRecords"`
	TimeInLowRecordsDelta *int `json:"timeInLowRecordsDelta" bson:"timeInLowRecordsDelta"`

	HasTimeInVeryLowPercent   bool     `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`
	TimeInVeryLowPercent      *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`
	TimeInVeryLowPercentDelta *float64 `json:"timeInVeryLowPercentDelta" bson:"timeInVeryLowPercentDelta"`

	HasTimeInVeryLowRecords   bool `json:"hasTimeInVeryLowRecords" bson:"hasTimeInVeryLowRecords"`
	TimeInVeryLowRecords      *int `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`
	TimeInVeryLowRecordsDelta *int `json:"timeInVeryLowRecordsDelta" bson:"timeInVeryLowRecordsDelta"`

	HasTimeInAnyLowPercent   bool     `json:"hasTimeInAnyLowPercent" bson:"hasTimeInAnyLowPercent"`
	TimeInAnyLowPercent      *float64 `json:"timeInAnyLowPercent" bson:"timeInAnyLowPercent"`
	TimeInAnyLowPercentDelta *float64 `json:"timeInAnyLowPercentDelta" bson:"timeInAnyLowPercentDelta"`

	HasTimeInAnyLowRecords   bool `json:"hasTimeInAnyLowRecords" bson:"hasTimeInAnyLowRecords"`
	TimeInAnyLowRecords      *int `json:"timeInAnyLowRecords" bson:"timeInAnyLowRecords"`
	TimeInAnyLowRecordsDelta *int `json:"timeInAnyLowRecordsDelta" bson:"timeInAnyLowRecordsDelta"`

	HasTimeInHighPercent   bool     `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	TimeInHighPercent      *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`
	TimeInHighPercentDelta *float64 `json:"timeInHighPercentDelta" bson:"timeInHighPercentDelta"`

	HasTimeInHighRecords   bool `json:"hasTimeInHighRecords" bson:"hasTimeInHighRecords"`
	TimeInHighRecords      *int `json:"timeInHighRecords" bson:"timeInHighRecords"`
	TimeInHighRecordsDelta *int `json:"timeInHighRecordsDelta" bson:"timeInHighRecordsDelta"`

	HasTimeInVeryHighPercent   bool     `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	TimeInVeryHighPercent      *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`
	TimeInVeryHighPercentDelta *float64 `json:"timeInVeryHighPercentDelta" bson:"timeInVeryHighPercentDelta"`

	HasTimeInVeryHighRecords   bool `json:"hasTimeInVeryHighRecords" bson:"hasTimeInVeryHighRecords"`
	TimeInVeryHighRecords      *int `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
	TimeInVeryHighRecordsDelta *int `json:"timeInVeryHighRecordsDelta" bson:"timeInVeryHighRecordsDelta"`

	HasTimeInAnyHighPercent   bool     `json:"hasTimeInAnyHighPercent" bson:"hasTimeInAnyHighPercent"`
	TimeInAnyHighPercent      *float64 `json:"timeInAnyHighPercent" bson:"timeInAnyHighPercent"`
	TimeInAnyHighPercentDelta *float64 `json:"timeInAnyHighPercentDelta" bson:"timeInAnyHighPercentDelta"`

	HasTimeInAnyHighRecords   bool `json:"hasTimeInAnyHighRecords" bson:"hasTimeInAnyHighRecords"`
	TimeInAnyHighRecords      *int `json:"timeInAnyHighRecords" bson:"timeInAnyHighRecords"`
	TimeInAnyHighRecordsDelta *int `json:"timeInAnyHighRecordsDelta" bson:"timeInAnyHighRecordsDelta"`
}

type BGMPeriods map[string]*BGMPeriod

type BGMStats struct {
	Periods       BGMPeriods                               `json:"periods" bson:"periods"`
	OffsetPeriods BGMPeriods                               `json:"offsetPeriods" bson:"offsetPeriods"`
	Buckets       []*Bucket[*BGMBucketData, BGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours    int                                      `json:"totalHours" bson:"totalHours"`
}

func (*BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (*BGMStats) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}

func (s *BGMStats) Init() {
	s.Buckets = make([]*Bucket[*BGMBucketData, BGMBucketData], 0)
	s.Periods = make(map[string]*BGMPeriod)
	s.OffsetPeriods = make(map[string]*BGMPeriod)
	s.TotalHours = 0
}

func (s *BGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *BGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *BGMStats) ClearInvalidatedBuckets(earliestModified time.Time) (firstData time.Time) {
	if len(s.Buckets) == 0 {
		return
	} else if earliestModified.After(s.Buckets[len(s.Buckets)-1].LastRecordTime) {
		return s.Buckets[len(s.Buckets)-1].LastRecordTime
	} else if earliestModified.Before(s.Buckets[0].Date) || earliestModified.Equal(s.Buckets[0].Date) {
		// we are before all existing buckets, remake for GC
		s.Buckets = make([]*Bucket[*BGMBucketData, BGMBucketData], 0)
		return
	}

	offset := len(s.Buckets) - (int(s.Buckets[len(s.Buckets)-1].Date.Sub(earliestModified.UTC().Truncate(time.Hour)).Hours()) + 1)

	for i := offset; i < len(s.Buckets); i++ {
		s.Buckets[i] = nil
	}
	s.Buckets = s.Buckets[:offset]

	if len(s.Buckets) > 0 {
		return s.Buckets[len(s.Buckets)-1].LastRecordTime
	}
	return
}

func (s *BGMStats) Update(ctx context.Context, cursor summary.DeviceDataCursor) error {
	for cursor.Next(ctx) {
		userData, err := cursor.GetNextBatch(ctx)
		if err != nil {
			return err
		}
		if len(userData) > 0 {
			err = AddData(&s.Buckets, userData)
			if err != nil {
				return err
			}
		}
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
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := &BGMBucketData{}
	totalOffsetStats := &BGMBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		currentIndex := len(s.Buckets) - 1 - i

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if i == stopPoints[nextStopPoint]*24 {
				s.CalculatePeriod(stopPoints[nextStopPoint], false, totalStats)
				nextStopPoint++
			}

			totalStats.TargetRecords += s.Buckets[currentIndex].Data.TargetRecords
			totalStats.LowRecords += s.Buckets[currentIndex].Data.LowRecords
			totalStats.VeryLowRecords += s.Buckets[currentIndex].Data.VeryLowRecords
			totalStats.HighRecords += s.Buckets[currentIndex].Data.HighRecords
			totalStats.VeryHighRecords += s.Buckets[currentIndex].Data.VeryHighRecords

			totalStats.TotalGlucose += s.Buckets[currentIndex].Data.TotalGlucose
			totalStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords
		}

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(stopPoints) > nextOffsetStopPoint {
			if i == stopPoints[nextOffsetStopPoint]*24*2 {
				s.CalculatePeriod(stopPoints[nextOffsetStopPoint], true, totalOffsetStats)
				nextOffsetStopPoint++
				totalOffsetStats = &BGMBucketData{}
			}

			totalOffsetStats.TargetRecords += s.Buckets[currentIndex].Data.TargetRecords
			totalOffsetStats.LowRecords += s.Buckets[currentIndex].Data.LowRecords
			totalOffsetStats.VeryLowRecords += s.Buckets[currentIndex].Data.VeryLowRecords
			totalOffsetStats.HighRecords += s.Buckets[currentIndex].Data.HighRecords
			totalOffsetStats.VeryHighRecords += s.Buckets[currentIndex].Data.VeryHighRecords

			totalOffsetStats.TotalGlucose += s.Buckets[currentIndex].Data.TotalGlucose
			totalOffsetStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], false, totalStats)
	}
	for i := nextOffsetStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], true, totalOffsetStats)
		totalOffsetStats = &BGMBucketData{}
	}

	s.TotalHours = len(s.Buckets)

	s.CalculateDelta()
}

func (s *BGMStats) CalculateDelta() {
	// We do this as a separate pass through the periods as the amount of tracking required to reverse the iteration
	// and fill this in during the period calculation would likely nullify any benefits, at least with the current
	// approach.

	for k := range s.Periods {
		if s.Periods[k].AverageGlucoseMmol != nil && s.OffsetPeriods[k].AverageGlucoseMmol != nil {
			delta := *s.Periods[k].AverageGlucoseMmol - *s.OffsetPeriods[k].AverageGlucoseMmol

			s.Periods[k].AverageGlucoseMmolDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].AverageGlucoseMmolDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TotalRecords != nil && s.OffsetPeriods[k].TotalRecords != nil {
			delta := *s.Periods[k].TotalRecords - *s.OffsetPeriods[k].TotalRecords

			s.Periods[k].TotalRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TotalRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].AverageDailyRecords != nil && s.OffsetPeriods[k].AverageDailyRecords != nil {
			delta := *s.Periods[k].AverageDailyRecords - *s.OffsetPeriods[k].AverageDailyRecords

			s.Periods[k].AverageDailyRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].AverageDailyRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInTargetPercent != nil && s.OffsetPeriods[k].TimeInTargetPercent != nil {
			delta := *s.Periods[k].TimeInTargetPercent - *s.OffsetPeriods[k].TimeInTargetPercent

			s.Periods[k].TimeInTargetPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInTargetPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInTargetRecords != nil && s.OffsetPeriods[k].TimeInTargetRecords != nil {
			delta := *s.Periods[k].TimeInTargetRecords - *s.OffsetPeriods[k].TimeInTargetRecords

			s.Periods[k].TimeInTargetRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInTargetRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInLowPercent != nil && s.OffsetPeriods[k].TimeInLowPercent != nil {
			delta := *s.Periods[k].TimeInLowPercent - *s.OffsetPeriods[k].TimeInLowPercent

			s.Periods[k].TimeInLowPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInLowPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInLowRecords != nil && s.OffsetPeriods[k].TimeInLowRecords != nil {
			delta := *s.Periods[k].TimeInLowRecords - *s.OffsetPeriods[k].TimeInLowRecords

			s.Periods[k].TimeInLowRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInLowRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInVeryLowPercent != nil && s.OffsetPeriods[k].TimeInVeryLowPercent != nil {
			delta := *s.Periods[k].TimeInVeryLowPercent - *s.OffsetPeriods[k].TimeInVeryLowPercent

			s.Periods[k].TimeInVeryLowPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryLowPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInVeryLowRecords != nil && s.OffsetPeriods[k].TimeInVeryLowRecords != nil {
			delta := *s.Periods[k].TimeInVeryLowRecords - *s.OffsetPeriods[k].TimeInVeryLowRecords

			s.Periods[k].TimeInVeryLowRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryLowRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInAnyLowPercent != nil && s.OffsetPeriods[k].TimeInAnyLowPercent != nil {
			delta := *s.Periods[k].TimeInAnyLowPercent - *s.OffsetPeriods[k].TimeInAnyLowPercent

			s.Periods[k].TimeInAnyLowPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyLowPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInAnyLowRecords != nil && s.OffsetPeriods[k].TimeInAnyLowRecords != nil {
			delta := *s.Periods[k].TimeInAnyLowRecords - *s.OffsetPeriods[k].TimeInAnyLowRecords

			s.Periods[k].TimeInAnyLowRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyLowRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInHighPercent != nil && s.OffsetPeriods[k].TimeInHighPercent != nil {
			delta := *s.Periods[k].TimeInHighPercent - *s.OffsetPeriods[k].TimeInHighPercent

			s.Periods[k].TimeInHighPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInHighPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInHighRecords != nil && s.OffsetPeriods[k].TimeInHighRecords != nil {
			delta := *s.Periods[k].TimeInHighRecords - *s.OffsetPeriods[k].TimeInHighRecords

			s.Periods[k].TimeInHighRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInHighRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInVeryHighPercent != nil && s.OffsetPeriods[k].TimeInVeryHighPercent != nil {
			delta := *s.Periods[k].TimeInVeryHighPercent - *s.OffsetPeriods[k].TimeInVeryHighPercent

			s.Periods[k].TimeInVeryHighPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryHighPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInVeryHighRecords != nil && s.OffsetPeriods[k].TimeInVeryHighRecords != nil {
			delta := *s.Periods[k].TimeInVeryHighRecords - *s.OffsetPeriods[k].TimeInVeryHighRecords

			s.Periods[k].TimeInVeryHighRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryHighRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInAnyHighPercent != nil && s.OffsetPeriods[k].TimeInAnyHighPercent != nil {
			delta := *s.Periods[k].TimeInAnyHighPercent - *s.OffsetPeriods[k].TimeInAnyHighPercent

			s.Periods[k].TimeInAnyHighPercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyHighPercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeInAnyHighRecords != nil && s.OffsetPeriods[k].TimeInAnyHighRecords != nil {
			delta := *s.Periods[k].TimeInAnyHighRecords - *s.OffsetPeriods[k].TimeInAnyHighRecords

			s.Periods[k].TimeInAnyHighRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyHighRecordsDelta = pointer.FromAny(-delta)
		}
	}
}

func (s *BGMStats) CalculatePeriod(i int, offset bool, totalStats *BGMBucketData) {
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

		HasTimeInAnyLowRecords: true,
		TimeInAnyLowRecords:    pointer.FromAny(totalStats.VeryLowRecords + totalStats.LowRecords),

		HasTimeInHighRecords: true,
		TimeInHighRecords:    pointer.FromAny(totalStats.HighRecords),

		HasTimeInVeryHighRecords: true,
		TimeInVeryHighRecords:    pointer.FromAny(totalStats.VeryHighRecords),

		HasTimeInAnyHighRecords: true,
		TimeInAnyHighRecords:    pointer.FromAny(totalStats.VeryHighRecords + totalStats.HighRecords),
	}

	if totalStats.TotalRecords != 0 {
		newPeriod.HasTimeInTargetPercent = true
		newPeriod.TimeInTargetPercent = pointer.FromAny(float64(totalStats.TargetRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInLowPercent = true
		newPeriod.TimeInLowPercent = pointer.FromAny(float64(totalStats.LowRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInVeryLowPercent = true
		newPeriod.TimeInVeryLowPercent = pointer.FromAny(float64(totalStats.VeryLowRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInAnyLowPercent = true
		newPeriod.TimeInAnyLowPercent = pointer.FromAny(float64(totalStats.VeryLowRecords+totalStats.LowRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInHighPercent = true
		newPeriod.TimeInHighPercent = pointer.FromAny(float64(totalStats.HighRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInVeryHighPercent = true
		newPeriod.TimeInVeryHighPercent = pointer.FromAny(float64(totalStats.VeryHighRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasTimeInAnyHighPercent = true
		newPeriod.TimeInAnyHighPercent = pointer.FromAny(float64(totalStats.VeryHighRecords+totalStats.HighRecords) / float64(totalStats.TotalRecords))

		newPeriod.HasAverageGlucoseMmol = true
		newPeriod.AverageGlucoseMmol = pointer.FromAny(totalStats.TotalGlucose / float64(totalStats.TotalRecords))
	}

	if offset {
		s.OffsetPeriods[strconv.Itoa(i)+"d"] = newPeriod
	} else {
		s.Periods[strconv.Itoa(i)+"d"] = newPeriod
	}
}
