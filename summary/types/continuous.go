package types

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
)

// This is a good example of what a summary type requires, as it does not share as many pieces as CGM/BGM

type ContinuousPeriods map[string]*ContinuousPeriod

func (*ContinuousPeriods) GetType() string {
	return SummaryTypeContinuous
}

func (*ContinuousPeriods) GetDeviceDataTypes() []string {
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

func (rs *ContinuousRanges) Add(new *ContinuousRanges) {
	rs.Realtime.Add(&new.Realtime)
	rs.Total.Add(&new.Total)
	rs.Deferred.Add(&new.Deferred)
}

func (rs *ContinuousRanges) Finalize() {
	rs.Realtime.Percent = float64(rs.Realtime.Records) / float64(rs.Total.Records)
	rs.Deferred.Percent = float64(rs.Deferred.Records) / float64(rs.Total.Records)
}

type ContinuousBucket struct {
	ContinuousRanges `json:",inline" bson:",inline"`
}

func (b *ContinuousBucket) Update(r data.Datum, _ *time.Time) (bool, error) {
	record, err := NewGlucose(r)
	if err != nil {
		return false, err
	}

	// NOTE we do not call range.update here, as we only require a single field of a range
	if record.GetCreatedTime().Sub(*record.GetTime()).Hours() < 24 {
		b.Realtime.Records++
	} else {
		b.Deferred.Records++
	}

	b.Total.Records++

	return true, nil
}

type ContinuousPeriod struct {
	ContinuousRanges `json:",inline" bson:",inline"`

	AverageDailyRecords float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`

	state CalcState
}

func (p *ContinuousPeriod) Update(bucket *Bucket[*ContinuousBucket, ContinuousBucket]) error {
	if p.state.Final {
		return errors.New("period has been finalized, cannot add any data")
	}

	if bucket.Data.Total.Records == 0 {
		return nil
	}

	if p.state.FirstCountedHour.IsZero() && p.state.FirstCountedHour.IsZero() {
		p.state.FirstCountedHour = bucket.Time
		p.state.LastCountedHour = bucket.Time
	} else {
		if bucket.Time.Before(p.state.FirstCountedHour) {
			p.state.FirstCountedHour = bucket.Time
		} else if bucket.Time.After(p.state.LastCountedHour) {
			p.state.LastCountedHour = bucket.Time
		} else {
			return fmt.Errorf("bucket of time %s is within the existing period range of %s - %s",
				bucket.Time, p.state.FirstCountedHour, p.state.LastCountedHour)
		}
	}

	p.Add(&bucket.Data.ContinuousRanges)

	return nil
}

func (p *ContinuousPeriod) Finalize(days int) {
	if p.state.Final != false {
		return
	}
	p.state.Final = true

	p.ContinuousRanges.Finalize()
	p.AverageDailyRecords = float64(p.Total.Records) / float64(days)
}

func (s *ContinuousPeriods) Init() {
	*s = make(map[string]*ContinuousPeriod)
}

func (s *ContinuousPeriods) Update(ctx context.Context, bucketsCursor *mongo.Cursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d
	period := ContinuousPeriod{}

	var stopPoints []time.Time
	nextStopPoint := 0

	previousBucketTime := time.Time{}

	for bucketsCursor.Next(ctx) {
		bucket := &Bucket[*ContinuousBucket, ContinuousBucket]{}
		if err := bucketsCursor.Decode(bucket); err != nil {
			return err
		}

		if !previousBucketTime.IsZero() && bucket.Time.Compare(previousBucketTime) >= 0 {
			return fmt.Errorf("bucket with date %s is equal or later than to the last added bucket with date %s, "+
				"buckets must be in reverse order and unique", bucket.Time, previousBucketTime)
		}
		previousBucketTime = bucket.Time

		// We should have the newest (last) bucket here, use its date for breakpoints
		if stopPoints == nil {
			stopPoints, _ = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		if len(stopPoints) > nextStopPoint && bucket.Time.Compare(stopPoints[nextStopPoint]) <= 0 {
			s.CalculatePeriod(periodLengths[nextStopPoint], false, period)
			nextStopPoint++
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			if err := period.Update(bucket); err != nil {
				return err
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(periodLengths[i], false, period)
	}

	return nil
}

func (s *ContinuousPeriods) CalculatePeriod(i int, _ bool, period ContinuousPeriod) {
	// We don't make a copy of period, as the struct has no pointers... right? you didn't add any pointers right?
	period.Finalize(i)
	(*s)[strconv.Itoa(i)+"d"] = &period
}
