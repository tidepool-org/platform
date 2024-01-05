package types

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"

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
	HasTimeCGMUsePercent   bool     `json:"hasTimeCGMUsePercent" bson:"hasTimeCGMUsePercent"`
	TimeCGMUsePercent      *float64 `json:"timeCGMUsePercent" bson:"timeCGMUsePercent"`
	TimeCGMUsePercentDelta *float64 `json:"timeCGMUsePercentDelta" bson:"timeCGMUsePercentDelta"`

	HasTimeCGMUseMinutes   bool `json:"hasTimeCGMUseMinutes" bson:"hasTimeCGMUseMinutes"`
	TimeCGMUseMinutes      *int `json:"timeCGMUseMinutes" bson:"timeCGMUseMinutes"`
	TimeCGMUseMinutesDelta *int `json:"timeCGMUseMinutesDelta" bson:"timeCGMUseMinutesDelta"`

	HasTimeCGMUseRecords   bool `json:"hasTimeCGMUseRecords" bson:"hasTimeCGMUseRecords"`
	TimeCGMUseRecords      *int `json:"timeCGMUseRecords" bson:"timeCGMUseRecords"`
	TimeCGMUseRecordsDelta *int `json:"timeCGMUseRecordsDelta" bson:"timeCGMUseRecordsDelta"`

	HasAverageGlucoseMmol   bool     `json:"hasAverageGlucoseMmol" bson:"hasAverageGlucoseMmol"`
	AverageGlucoseMmol      *float64 `json:"averageGlucoseMmol" bson:"averageGlucoseMmol"`
	AverageGlucoseMmolDelta *float64 `json:"averageGlucoseMmolDelta" bson:"averageGlucoseMmolDelta"`

	HasGlucoseManagementIndicator   bool     `json:"hasGlucoseManagementIndicator" bson:"hasGlucoseManagementIndicator"`
	GlucoseManagementIndicator      *float64 `json:"glucoseManagementIndicator" bson:"glucoseManagementIndicator"`
	GlucoseManagementIndicatorDelta *float64 `json:"glucoseManagementIndicatorDelta" bson:"glucoseManagementIndicatorDelta"`

	HasTotalRecords   bool `json:"hasTotalRecords" bson:"hasTotalRecords"`
	TotalRecords      *int `json:"totalRecords" bson:"totalRecords"`
	TotalRecordsDelta *int `json:"totalRecordsDelta" bson:"totalRecordsDelta"`

	HasAverageDailyRecords   bool     `json:"hasAverageDailyRecords" bson:"hasAverageDailyRecords"`
	AverageDailyRecords      *float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`
	AverageDailyRecordsDelta *float64 `json:"averageDailyRecordsDelta" bson:"averageDailyRecordsDelta"`

	HasTimeInTargetPercent   bool     `json:"hasTimeInTargetPercent" bson:"hasTimeInTargetPercent"`
	TimeInTargetPercent      *float64 `json:"timeInTargetPercent" bson:"timeInTargetPercent"`
	TimeInTargetPercentDelta *float64 `json:"timeInTargetPercentDelta" bson:"timeInTargetPercentDelta"`

	HasTimeInTargetMinutes   bool `json:"hasTimeInTargetMinutes" bson:"hasTimeInTargetMinutes"`
	TimeInTargetMinutes      *int `json:"timeInTargetMinutes" bson:"timeInTargetMinutes"`
	TimeInTargetMinutesDelta *int `json:"timeInTargetMinutesDelta" bson:"timeInTargetMinutesDelta"`

	HasTimeInTargetRecords   bool `json:"hasTimeInTargetRecords" bson:"hasTimeInTargetRecords"`
	TimeInTargetRecords      *int `json:"timeInTargetRecords" bson:"timeInTargetRecords"`
	TimeInTargetRecordsDelta *int `json:"timeInTargetRecordsDelta" bson:"timeInTargetRecordsDelta"`

	HasTimeInLowPercent   bool     `json:"hasTimeInLowPercent" bson:"hasTimeInLowPercent"`
	TimeInLowPercent      *float64 `json:"timeInLowPercent" bson:"timeInLowPercent"`
	TimeInLowPercentDelta *float64 `json:"timeInLowPercentDelta" bson:"timeInLowPercentDelta"`

	HasTimeInLowMinutes   bool `json:"hasTimeInLowMinutes" bson:"hasTimeInLowMinutes"`
	TimeInLowMinutes      *int `json:"timeInLowMinutes" bson:"timeInLowMinutes"`
	TimeInLowMinutesDelta *int `json:"timeInLowMinutesDelta" bson:"timeInLowMinutesDelta"`

	HasTimeInLowRecords   bool `json:"hasTimeInLowRecords" bson:"hasTimeInLowRecords"`
	TimeInLowRecords      *int `json:"timeInLowRecords" bson:"timeInLowRecords"`
	TimeInLowRecordsDelta *int `json:"timeInLowRecordsDelta" bson:"timeInLowRecordsDelta"`

	HasTimeInVeryLowPercent   bool     `json:"hasTimeInVeryLowPercent" bson:"hasTimeInVeryLowPercent"`
	TimeInVeryLowPercent      *float64 `json:"timeInVeryLowPercent" bson:"timeInVeryLowPercent"`
	TimeInVeryLowPercentDelta *float64 `json:"timeInVeryLowPercentDelta" bson:"timeInVeryLowPercentDelta"`

	HasTimeInVeryLowMinutes   bool `json:"hasTimeInVeryLowMinutes" bson:"hasTimeInVeryLowMinutes"`
	TimeInVeryLowMinutes      *int `json:"timeInVeryLowMinutes" bson:"timeInVeryLowMinutes"`
	TimeInVeryLowMinutesDelta *int `json:"timeInVeryLowMinutesDelta" bson:"timeInVeryLowMinutesDelta"`

	HasTimeInVeryLowRecords   bool `json:"hasTimeInVeryLowRecords" bson:"hasTimeInVeryLowRecords"`
	TimeInVeryLowRecords      *int `json:"timeInVeryLowRecords" bson:"timeInVeryLowRecords"`
	TimeInVeryLowRecordsDelta *int `json:"timeInVeryLowRecordsDelta" bson:"timeInVeryLowRecordsDelta"`

	HasTimeInAnyLowPercent   bool     `json:"hasTimeInAnyLowPercent" bson:"hasTimeInAnyLowPercent"`
	TimeInAnyLowPercent      *float64 `json:"timeInAnyLowPercent" bson:"timeInAnyLowPercent"`
	TimeInAnyLowPercentDelta *float64 `json:"timeInAnyLowPercentDelta" bson:"timeInAnyLowPercentDelta"`

	HasTimeInAnyLowMinutes   bool `json:"hasTimeInAnyLowMinutes" bson:"hasTimeInAnyLowMinutes"`
	TimeInAnyLowMinutes      *int `json:"timeInAnyLowMinutes" bson:"timeInAnyLowMinutes"`
	TimeInAnyLowMinutesDelta *int `json:"timeInAnyLowMinutesDelta" bson:"timeInAnyLowMinutesDelta"`

	HasTimeInAnyLowRecords   bool `json:"hasTimeInAnyLowRecords" bson:"hasTimeInAnyLowRecords"`
	TimeInAnyLowRecords      *int `json:"timeInAnyLowRecords" bson:"timeInAnyLowRecords"`
	TimeInAnyLowRecordsDelta *int `json:"timeInAnyLowRecordsDelta" bson:"timeInAnyLowRecordsDelta"`

	HasTimeInHighPercent   bool     `json:"hasTimeInHighPercent" bson:"hasTimeInHighPercent"`
	TimeInHighPercent      *float64 `json:"timeInHighPercent" bson:"timeInHighPercent"`
	TimeInHighPercentDelta *float64 `json:"timeInHighPercentDelta" bson:"timeInHighPercentDelta"`

	HasTimeInHighMinutes   bool `json:"hasTimeInHighMinutes" bson:"hasTimeInHighMinutes"`
	TimeInHighMinutes      *int `json:"timeInHighMinutes" bson:"timeInHighMinutes"`
	TimeInHighMinutesDelta *int `json:"timeInHighMinutesDelta" bson:"timeInHighMinutesDelta"`

	HasTimeInHighRecords   bool `json:"hasTimeInHighRecords" bson:"hasTimeInHighRecords"`
	TimeInHighRecords      *int `json:"timeInHighRecords" bson:"timeInHighRecords"`
	TimeInHighRecordsDelta *int `json:"timeInHighRecordsDelta" bson:"timeInHighRecordsDelta"`

	HasTimeInVeryHighPercent   bool     `json:"hasTimeInVeryHighPercent" bson:"hasTimeInVeryHighPercent"`
	TimeInVeryHighPercent      *float64 `json:"timeInVeryHighPercent" bson:"timeInVeryHighPercent"`
	TimeInVeryHighPercentDelta *float64 `json:"timeInVeryHighPercentDelta" bson:"timeInVeryHighPercentDelta"`

	HasTimeInVeryHighMinutes   bool `json:"hasTimeInVeryHighMinutes" bson:"hasTimeInVeryHighMinutes"`
	TimeInVeryHighMinutes      *int `json:"timeInVeryHighMinutes" bson:"timeInVeryHighMinutes"`
	TimeInVeryHighMinutesDelta *int `json:"timeInVeryHighMinutesDelta" bson:"timeInVeryHighMinutesDelta"`

	HasTimeInVeryHighRecords   bool `json:"hasTimeInVeryHighRecords" bson:"hasTimeInVeryHighRecords"`
	TimeInVeryHighRecords      *int `json:"timeInVeryHighRecords" bson:"timeInVeryHighRecords"`
	TimeInVeryHighRecordsDelta *int `json:"timeInVeryHighRecordsDelta" bson:"timeInVeryHighRecordsDelta"`

	HasTimeInAnyHighPercent   bool     `json:"hasTimeInAnyHighPercent" bson:"hasTimeInAnyHighPercent"`
	TimeInAnyHighPercent      *float64 `json:"timeInAnyHighPercent" bson:"timeInAnyHighPercent"`
	TimeInAnyHighPercentDelta *float64 `json:"timeInAnyHighPercentDelta" bson:"timeInAnyHighPercentDelta"`

	HasTimeInAnyHighMinutes   bool `json:"hasTimeInAnyHighMinutes" bson:"hasTimeInAnyHighMinutes"`
	TimeInAnyHighMinutes      *int `json:"timeInAnyHighMinutes" bson:"timeInAnyHighMinutes"`
	TimeInAnyHighMinutesDelta *int `json:"timeInAnyHighMinutesDelta" bson:"timeInAnyHighMinutesDelta"`

	HasTimeInAnyHighRecords   bool `json:"hasTimeInAnyHighRecords" bson:"hasTimeInAnyHighRecords"`
	TimeInAnyHighRecords      *int `json:"timeInAnyHighRecords" bson:"timeInAnyHighRecords"`
	TimeInAnyHighRecordsDelta *int `json:"timeInAnyHighRecordsDelta" bson:"timeInAnyHighRecordsDelta"`
}

type CGMPeriods map[string]*CGMPeriod

type CGMStats struct {
	Periods       CGMPeriods                             `json:"periods" bson:"periods"`
	OffsetPeriods CGMPeriods                             `json:"offsetPeriods" bson:"offsetPeriods"`
	Buckets       Buckets[CGMBucketData, *CGMBucketData] `json:"buckets" bson:"buckets"`
	TotalHours    int                                    `json:"totalHours" bson:"totalHours"`
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataType() string {
	return continuous.Type
}

func (s *CGMStats) Init() {
	s.Buckets = make(Buckets[CGMBucketData, *CGMBucketData], 0)
	s.Periods = make(map[string]*CGMPeriod)
	s.OffsetPeriods = make(map[string]*CGMPeriod)
	s.TotalHours = 0
}

func (s *CGMStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *CGMStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *CGMStats) ClearInvalidatedBuckets(status *UserLastUpdated) {
	offset := int(status.EarliestModified.Sub(s.Buckets[0].Date).Hours())

	for i := offset; i < len(s.Buckets); i++ {
		s.Buckets[i] = nil
	}
	s.Buckets = s.Buckets[offset:]
}

func (s *CGMStats) Update(ctx context.Context, cursor *mongo.Cursor) error {
	var userData []*glucoseDatum.Glucose = nil
	var err error

	count := 0
	for cursor.Next(ctx) {
		count++
		if userData == nil {
			userData = make([]*glucoseDatum.Glucose, 0, cursor.RemainingBatchLength())
		}

		r := glucoseDatum.Glucose{}
		if err = cursor.Decode(&r); err != nil {
			return fmt.Errorf("unable to decode userData: %w", err)
		}
		userData = append(userData, &r)

		// we call AddData before each network call to the db to reduce thrashing
		if cursor.RemainingBatchLength() != 0 {
			err = AddData(&s.Buckets, userData)
			if err != nil {
				return err
			}
			userData = nil
		}
	}

	// add the final partial batch
	if userData != nil {
		err = AddData(&s.Buckets, userData)
		if err != nil {
			return err
		}
		userData = nil
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
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := &CGMBucketData{}
	totalOffsetStats := &CGMBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		currentIndex := len(s.Buckets) - 1 - i

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if i == stopPoints[nextStopPoint]*24 {
				s.CalculatePeriod(stopPoints[nextStopPoint], false, totalStats)
				nextStopPoint++
			}

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

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(stopPoints) > nextOffsetStopPoint {
			if i == stopPoints[nextOffsetStopPoint]*24*2 {
				s.CalculatePeriod(stopPoints[nextOffsetStopPoint], true, totalOffsetStats)
				nextOffsetStopPoint++
				totalOffsetStats = &CGMBucketData{}
			}
			totalOffsetStats.TargetMinutes += s.Buckets[currentIndex].Data.TargetMinutes
			totalOffsetStats.TargetRecords += s.Buckets[currentIndex].Data.TargetRecords

			totalOffsetStats.LowMinutes += s.Buckets[currentIndex].Data.LowMinutes
			totalOffsetStats.LowRecords += s.Buckets[currentIndex].Data.LowRecords

			totalOffsetStats.VeryLowMinutes += s.Buckets[currentIndex].Data.VeryLowMinutes
			totalOffsetStats.VeryLowRecords += s.Buckets[currentIndex].Data.VeryLowRecords

			totalOffsetStats.HighMinutes += s.Buckets[currentIndex].Data.HighMinutes
			totalOffsetStats.HighRecords += s.Buckets[currentIndex].Data.HighRecords

			totalOffsetStats.VeryHighMinutes += s.Buckets[currentIndex].Data.VeryHighMinutes
			totalOffsetStats.VeryHighRecords += s.Buckets[currentIndex].Data.VeryHighRecords

			totalOffsetStats.TotalGlucose += s.Buckets[currentIndex].Data.TotalGlucose
			totalOffsetStats.TotalMinutes += s.Buckets[currentIndex].Data.TotalMinutes
			totalOffsetStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], false, totalStats)
	}
	for i := nextOffsetStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], true, totalOffsetStats)
		totalOffsetStats = &CGMBucketData{}
	}

	s.TotalHours = len(s.Buckets)

	s.CalculateDelta()
}

func (s *CGMStats) CalculateDelta() {
	// We do this as a separate pass through the periods as the amount of tracking required to reverse the iteration
	// and fill this in during the period calculation would likely nullify any benefits, at least with the current
	// approach.

	for k := range s.Periods {
		if s.Periods[k].TimeCGMUsePercent != nil && s.OffsetPeriods[k].TimeCGMUsePercent != nil {
			delta := *s.Periods[k].TimeCGMUsePercent - *s.OffsetPeriods[k].TimeCGMUsePercent

			s.Periods[k].TimeCGMUsePercentDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeCGMUsePercentDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeCGMUseRecords != nil && s.OffsetPeriods[k].TimeCGMUseRecords != nil {
			delta := *s.Periods[k].TimeCGMUseRecords - *s.OffsetPeriods[k].TimeCGMUseRecords

			s.Periods[k].TimeCGMUseRecordsDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeCGMUseRecordsDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].TimeCGMUseMinutes != nil && s.OffsetPeriods[k].TimeCGMUseMinutes != nil {
			delta := *s.Periods[k].TimeCGMUseMinutes - *s.OffsetPeriods[k].TimeCGMUseMinutes

			s.Periods[k].TimeCGMUseMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeCGMUseMinutesDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].AverageGlucoseMmol != nil && s.OffsetPeriods[k].AverageGlucoseMmol != nil {
			delta := *s.Periods[k].AverageGlucoseMmol - *s.OffsetPeriods[k].AverageGlucoseMmol

			s.Periods[k].AverageGlucoseMmolDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].AverageGlucoseMmolDelta = pointer.FromAny(-delta)
		}

		if s.Periods[k].GlucoseManagementIndicator != nil && s.OffsetPeriods[k].GlucoseManagementIndicator != nil {
			delta := *s.Periods[k].GlucoseManagementIndicator - *s.OffsetPeriods[k].GlucoseManagementIndicator

			s.Periods[k].GlucoseManagementIndicatorDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].GlucoseManagementIndicatorDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInTargetMinutes != nil && s.OffsetPeriods[k].TimeInTargetMinutes != nil {
			delta := *s.Periods[k].TimeInTargetMinutes - *s.OffsetPeriods[k].TimeInTargetMinutes

			s.Periods[k].TimeInTargetMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInTargetMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInLowMinutes != nil && s.OffsetPeriods[k].TimeInLowMinutes != nil {
			delta := *s.Periods[k].TimeInLowMinutes - *s.OffsetPeriods[k].TimeInLowMinutes

			s.Periods[k].TimeInLowMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInLowMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInVeryLowMinutes != nil && s.OffsetPeriods[k].TimeInVeryLowMinutes != nil {
			delta := *s.Periods[k].TimeInVeryLowMinutes - *s.OffsetPeriods[k].TimeInVeryLowMinutes

			s.Periods[k].TimeInVeryLowMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryLowMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInAnyLowMinutes != nil && s.OffsetPeriods[k].TimeInAnyLowMinutes != nil {
			delta := *s.Periods[k].TimeInAnyLowMinutes - *s.OffsetPeriods[k].TimeInAnyLowMinutes

			s.Periods[k].TimeInAnyLowMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyLowMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInHighMinutes != nil && s.OffsetPeriods[k].TimeInHighMinutes != nil {
			delta := *s.Periods[k].TimeInHighMinutes - *s.OffsetPeriods[k].TimeInHighMinutes

			s.Periods[k].TimeInHighMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInHighMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInVeryHighMinutes != nil && s.OffsetPeriods[k].TimeInVeryHighMinutes != nil {
			delta := *s.Periods[k].TimeInVeryHighMinutes - *s.OffsetPeriods[k].TimeInVeryHighMinutes

			s.Periods[k].TimeInVeryHighMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInVeryHighMinutesDelta = pointer.FromAny(-delta)
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

		if s.Periods[k].TimeInAnyHighMinutes != nil && s.OffsetPeriods[k].TimeInAnyHighMinutes != nil {
			delta := *s.Periods[k].TimeInAnyHighMinutes - *s.OffsetPeriods[k].TimeInAnyHighMinutes

			s.Periods[k].TimeInAnyHighMinutesDelta = pointer.FromAny(delta)
			s.OffsetPeriods[k].TimeInAnyHighMinutesDelta = pointer.FromAny(-delta)
		}
	}
}

func (s *CGMStats) CalculatePeriod(i int, offset bool, totalStats *CGMBucketData) {
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

		HasTimeInAnyLowMinutes: true,
		TimeInAnyLowMinutes:    pointer.FromAny(totalStats.LowMinutes + totalStats.VeryLowMinutes),

		HasTimeInAnyLowRecords: true,
		TimeInAnyLowRecords:    pointer.FromAny(totalStats.LowRecords + totalStats.VeryLowRecords),

		HasTimeInHighMinutes: true,
		TimeInHighMinutes:    pointer.FromAny(totalStats.HighMinutes),

		HasTimeInHighRecords: true,
		TimeInHighRecords:    pointer.FromAny(totalStats.HighRecords),

		HasTimeInVeryHighMinutes: true,
		TimeInVeryHighMinutes:    pointer.FromAny(totalStats.VeryHighMinutes),

		HasTimeInVeryHighRecords: true,
		TimeInVeryHighRecords:    pointer.FromAny(totalStats.VeryHighRecords),

		HasTimeInAnyHighMinutes: true,
		TimeInAnyHighMinutes:    pointer.FromAny(totalStats.LowMinutes + totalStats.VeryLowMinutes),

		HasTimeInAnyHighRecords: true,
		TimeInAnyHighRecords:    pointer.FromAny(totalStats.HighRecords + totalStats.VeryHighRecords),
	}

	if totalStats.TotalRecords != 0 {
		realMinutes := CalculateRealMinutes(i, s.Buckets[len(s.Buckets)-1].LastRecordTime, s.Buckets[len(s.Buckets)-1].Data.LastRecordDuration)
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

			newPeriod.HasTimeInAnyLowPercent = true
			newPeriod.TimeInAnyLowPercent = pointer.FromAny(float64(totalStats.VeryLowRecords+totalStats.LowRecords) / float64(totalStats.TotalRecords))

			newPeriod.HasTimeInHighPercent = true
			newPeriod.TimeInHighPercent = pointer.FromAny(float64(totalStats.HighMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInVeryHighPercent = true
			newPeriod.TimeInVeryHighPercent = pointer.FromAny(float64(totalStats.VeryHighMinutes) / float64(totalStats.TotalMinutes))

			newPeriod.HasTimeInAnyHighPercent = true
			newPeriod.TimeInAnyHighPercent = pointer.FromAny(float64(totalStats.VeryHighRecords+totalStats.HighRecords) / float64(totalStats.TotalRecords))

		}

		newPeriod.HasAverageGlucoseMmol = true
		newPeriod.AverageGlucoseMmol = pointer.FromAny(totalStats.TotalGlucose / float64(totalStats.TotalMinutes))

		// we only add GMI if cgm use >70%, otherwise clear it
		if *newPeriod.TimeCGMUsePercent > 0.7 {
			newPeriod.HasGlucoseManagementIndicator = true
			newPeriod.GlucoseManagementIndicator = pointer.FromAny(CalculateGMI(*newPeriod.AverageGlucoseMmol))
		}
	}

	if offset {
		s.OffsetPeriods[strconv.Itoa(i)+"d"] = newPeriod
	} else {
		s.Periods[strconv.Itoa(i)+"d"] = newPeriod
	}

}
