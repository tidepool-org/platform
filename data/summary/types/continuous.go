package types

import (
	"context"
	"errors"
	"fmt"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

// This is a good example of what a summary type requires, as it does not share as many pieces as CGM/BGM

type ContinuousStats struct {
	Periods ContinuousPeriods `json:"periods" bson:"periods"`
}

func (*ContinuousStats) GetType() string {
	return SummaryTypeContinuous
}

func (*ContinuousStats) GetDeviceDataTypes() []string {
	return []string{continuous.Type, selfmonitored.Type}
}

type ContinuousRanges struct {
	// Realtime is the total count of records which were both uploaded within 24h of the record creation
	// and from a continuous dataset
	Realtime Range `json:"realtime" bson:"realtime"`

	// Deferred is the total count of records which are in continuous datasets, but not uploaded within 24h
	Deferred Range `json:"deferred" bson:"deferred"`

	// Total is the total count of all records, regardless of when they were created or uploaded
	Total Range `json:"total" bson:"total"`
}

func (c *ContinuousRanges) Add(new *ContinuousRanges) {
	c.Realtime.Add(&new.Realtime)
	c.Total.Add(&new.Total)
	c.Deferred.Add(&new.Deferred)
}

func (c *ContinuousRanges) Finalize() {
	c.Realtime.Percent = float64(c.Realtime.Records) / float64(c.Total.Records)
	c.Deferred.Percent = float64(c.Deferred.Records) / float64(c.Total.Records)
}

type ContinuousBucket struct {
	ContinuousRanges `json:",inline" bson:",inline"`
}

func (B *ContinuousBucket) Update(r data.Datum, _ *BucketShared) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("cgm or bgm record for calculation is not compatible with Glucose type")
	}

	// TODO validate record type matches bucket type

	// NOTE we do not call range.update here, as we only require a single field of a range
	if dataRecord.CreatedTime.Sub(*dataRecord.Time).Hours() < 24 {
		B.Realtime.Records++
	} else {
		B.Deferred.Records++
	}

	B.Total.Records++

	return true, nil
}

type ContinuousPeriod struct {
	ContinuousRanges `json:",inline" bson:",inline"`

	AverageDailyRecords float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`

	state CalcState
}

func (P *ContinuousPeriod) Update(bucket *Bucket[*ContinuousBucket, ContinuousBucket]) error {
	if P.state.Final {
		return errors.New("period has been finalized, cannot add any data")
	}

	if bucket.Data.Total.Records == 0 {
		return nil
	}

	if P.state.FirstCountedHour.IsZero() && P.state.FirstCountedHour.IsZero() {
		P.state.FirstCountedHour = bucket.Time
		P.state.LastCountedHour = bucket.Time
	} else {
		if bucket.Time.Before(P.state.FirstCountedHour) {
			P.state.FirstCountedHour = bucket.Time
		} else if bucket.Time.After(P.state.LastCountedHour) {
			P.state.LastCountedHour = bucket.Time
		} else {
			return fmt.Errorf("bucket of time %s is within the existing period range of %s - %s",
				bucket.Time, P.state.FirstCountedHour, P.state.LastCountedHour)
		}
	}

	P.Add(&bucket.Data.ContinuousRanges)

	return nil
}

func (P *ContinuousPeriod) Finalize(days int) {
	if P.state.Final != false {
		return
	}
	P.state.Final = true

	P.ContinuousRanges.Finalize()
	P.AverageDailyRecords = float64(P.Total.Records) / float64(days)
}

type ContinuousPeriods map[string]*ContinuousPeriod

func (s *ContinuousStats) Init() {
	s.Periods = make(map[string]*ContinuousPeriod)
}

func (s *ContinuousStats) Update(ctx context.Context, bucketsCursor *mongo.Cursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d,
	// currently only supports day precision
	nextStopPoint := 0
	totalStats := ContinuousPeriod{}
	var err error
	var stopPoints []time.Time

	bucket := &Bucket[*ContinuousBucket, ContinuousBucket]{}

	for bucketsCursor.Next(ctx) {
		if err = bucketsCursor.Decode(bucket); err != nil {
			return err
		}

		// We should have the newest (last) bucket here, use its date for breakpoints
		if stopPoints == nil {
			stopPoints, _ = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		if len(stopPoints) > nextStopPoint && bucket.Time.Compare(stopPoints[nextStopPoint]) <= 0 {
			s.CalculatePeriod(periodLengths[nextStopPoint], false, totalStats)
			nextStopPoint++
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			err = totalStats.Update(bucket)
			if err != nil {
				return err
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(periodLengths[i], false, totalStats)
	}

	return nil
}

func (s *ContinuousStats) CalculatePeriod(i int, _ bool, period ContinuousPeriod) {
	// We don't make a copy of period, as the struct has no pointers... right? you didn't add any pointers right?
	period.Finalize(i)
	s.Periods[strconv.Itoa(i)+"d"] = &period
}
