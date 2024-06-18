package types

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/summary/fetcher"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

type ContinuousBucketData struct {
	TotalRecords int `json:"totalRecords" bson:"totalRecords"`

	// RealtimeRecords is the total count of records which were both uploaded within 24h of the record creation
	// and from a continuous dataset
	RealtimeRecords int `json:"realtimeRecords" bson:"realtimeRecords"`

	// DeferredRecords is the total count of records which are in continuous datasets, but not uploaded within 24h
	DeferredRecords int `json:"deferredRecords" bson:"deferredRecords"`
}

type ContinuousPeriod struct {
	TotalRecords *int `json:"totalRecords" bson:"totalRecords"`

	AverageDailyRecords *float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`

	RealtimeRecords *int     `json:"realtimeRecords" bson:"realtimeRecords"`
	RealtimePercent *float64 `json:"realtimeRecordsPercent" bson:"realtimeRecordsPercent"`
	DeferredRecords *int     `json:"deferredRecords" bson:"deferredRecords"`
	DeferredPercent *float64 `json:"deferredPercent" bson:"deferredPercent"`
}

type ContinuousPeriods map[string]*ContinuousPeriod

type ContinuousStats struct {
	Periods    ContinuousPeriods                                      `json:"periods" bson:"periods"`
	Buckets    []*Bucket[*ContinuousBucketData, ContinuousBucketData] `json:"buckets" bson:"buckets"`
	TotalHours int                                                    `json:"totalHours" bson:"totalHours"`
}

func (*ContinuousStats) GetType() string {
	return SummaryTypeContinuous
}

func (*ContinuousStats) GetDeviceDataTypes() []string {
	return DeviceDataTypes
}

func (s *ContinuousStats) Init() {
	s.Buckets = make([]*Bucket[*ContinuousBucketData, ContinuousBucketData], 0)
	s.Periods = make(map[string]*ContinuousPeriod)
	s.TotalHours = 0
}

func (s *ContinuousStats) GetBucketsLen() int {
	return len(s.Buckets)
}

func (s *ContinuousStats) GetBucketDate(i int) time.Time {
	return s.Buckets[i].Date
}

func (s *ContinuousStats) ClearInvalidatedBuckets(earliestModified time.Time) (firstData time.Time) {
	if len(s.Buckets) == 0 {
		return
	} else if earliestModified.After(s.Buckets[len(s.Buckets)-1].LastRecordTime) {
		return s.Buckets[len(s.Buckets)-1].LastRecordTime
	} else if earliestModified.Before(s.Buckets[0].Date) || earliestModified.Equal(s.Buckets[0].Date) {
		// we are before all existing buckets, remake for GC
		s.Buckets = make([]*Bucket[*ContinuousBucketData, ContinuousBucketData], 0)
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

func (s *ContinuousStats) Update(ctx context.Context, cursor fetcher.DeviceDataCursor) error {
	hasMoreData := true
	for hasMoreData {
		userData, err := cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
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

func (B *ContinuousBucketData) CalculateStats(r any, _ *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("continuous record for calculation is not compatible with Glucose type")
	}

	if dataRecord.CreatedTime.Sub(*dataRecord.Time).Hours() < 24 {
		B.RealtimeRecords++
	} else {
		B.DeferredRecords++
	}

	B.TotalRecords++

	return false, nil
}

func (s *ContinuousStats) CalculateSummary() {
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	totalStats := &ContinuousBucketData{}

	for i := 0; i < len(s.Buckets); i++ {
		currentIndex := len(s.Buckets) - 1 - i

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if i == stopPoints[nextStopPoint]*24 {
				s.CalculatePeriod(stopPoints[nextStopPoint], totalStats)
				nextStopPoint++
			}

			totalStats.TotalRecords += s.Buckets[currentIndex].Data.TotalRecords

			totalStats.DeferredRecords += s.Buckets[currentIndex].Data.DeferredRecords
			totalStats.RealtimeRecords += s.Buckets[currentIndex].Data.RealtimeRecords
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], totalStats)
	}

	s.TotalHours = len(s.Buckets)
}

func (s *ContinuousStats) CalculatePeriod(i int, totalStats *ContinuousBucketData) {
	newPeriod := &ContinuousPeriod{
		TotalRecords:        pointer.FromAny(totalStats.TotalRecords),
		AverageDailyRecords: pointer.FromAny(float64(totalStats.TotalRecords) / float64(i)),
		RealtimeRecords:     pointer.FromAny(totalStats.RealtimeRecords),
		DeferredRecords:     pointer.FromAny(totalStats.DeferredRecords),
	}

	if totalStats.TotalRecords != 0 {
		newPeriod.RealtimePercent = pointer.FromAny(float64(totalStats.RealtimeRecords) / float64(totalStats.TotalRecords))
		newPeriod.DeferredPercent = pointer.FromAny(float64(totalStats.DeferredRecords) / float64(totalStats.TotalRecords))
	}

	s.Periods[strconv.Itoa(i)+"d"] = newPeriod
}

func (s *ContinuousStats) GetNumberOfDaysWithRealtimeData(startTime time.Time, endTime time.Time) (count int) {
	loc1 := startTime.Location()
	loc2 := endTime.Location()

	startOffset := int(startTime.Sub(s.Buckets[0].Date.In(loc1)).Hours())
	// cap to start of list
	if startOffset < 0 {
		startOffset = 0
	}

	endOffset := int(endTime.Sub(s.Buckets[0].Date.In(loc2)).Hours())
	// cap to end of list
	if endOffset > len(s.Buckets) {
		endOffset = len(s.Buckets)
	}

	for i := startOffset; i < endOffset; i++ {
		if s.Buckets[i].Data.RealtimeRecords > 0 {
			count += 1
			i += 23 - s.Buckets[i].Date.In(loc1).Hour()
			continue
		}
	}

	return count
}
