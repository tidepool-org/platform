package types

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/summary/fetcher"

	"github.com/tidepool-org/platform/data"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
)

// This is a good example of what a summary type requires, as it does not share as many pieces as CGM/BGM

type ContinuousStats struct {
	Periods    ContinuousPeriods `json:"periods" bson:"periods"`
	TotalHours int               `json:"totalHours" bson:"totalHours"`
}

func (*ContinuousStats) GetType() string {
	return SummaryTypeContinuous
}

func (*ContinuousStats) GetDeviceDataTypes() []string {
	return DeviceDataTypes
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

func (CR *ContinuousRanges) Add(new *ContinuousRanges) {
	CR.Realtime.Add(&new.Realtime)
	CR.Total.Add(&new.Total)
	CR.Deferred.Add(&new.Deferred)
}

func (CR *ContinuousRanges) Finalize() {
	CR.Realtime.Percent = float64(CR.Realtime.Records) / float64(CR.Total.Records)
	CR.Deferred.Percent = float64(CR.Deferred.Records) / float64(CR.Total.Records)
}

type ContinuousBucket struct {
	BucketShared     `json:",inline" bson:",inline"`
	ContinuousRanges `json:",inline" bson:",inline"`
}

func (B *ContinuousBucket) Add(bucket *ContinuousBucket) {
	B.Add(bucket)
}

func (B *ContinuousBucket) Update(r data.Datum, shared *BucketShared) error {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return errors.New("cgm or bgm record for calculation is not compatible with Glucose type")
	}

	// TODO validate record type matches bucket type

	if dataRecord.CreatedTime.Sub(*dataRecord.Time).Hours() < 24 {
		B.Realtime.Records++
	} else {
		B.Deferred.Records++
	}

	B.Total.Records++

	return nil
}

type ContinuousPeriod struct {
	ContinuousRanges `json:",inline" bson:",inline"`

	AverageDailyRecords float64 `json:"averageDailyRecords" bson:"averageDailyRecords"`
}

func (P ContinuousPeriod) Update(bucket *Bucket[*ContinuousBucket, ContinuousBucket]) {
	if bucket.Data.Total.Records == 0 {
		return
	}

	P.Add(&bucket.Data.ContinuousRanges)
}

func (P ContinuousPeriod) Finalize(days int) {
	P.ContinuousRanges.Finalize()
	P.AverageDailyRecords = float64(P.Total.Records) / float64(days)
}

type ContinuousPeriods map[string]*ContinuousPeriod

func (s *ContinuousStats) Init() {
	s.Periods = make(map[string]*ContinuousPeriod)
	s.TotalHours = 0
}

func (s *ContinuousStats) Update(ctx context.Context, shared SummaryShared, bucketsFetcher BucketFetcher[*ContinuousBucket, ContinuousBucket], cursor fetcher.DeviceDataCursor) error {
	// move all of this to a generic method? fetcher interface?

	hasMoreData := true
	var buckets BucketsByTime[*ContinuousBucket, ContinuousBucket]
	var err error
	var userData []data.Datum
	var startTime time.Time
	var endTime time.Time

	for hasMoreData {
		userData, err = cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			startTime = userData[0].GetTime().UTC().Truncate(time.Hour)
			endTime = userData[len(userData)].GetTime().UTC().Truncate(time.Hour)
			buckets, err = bucketsFetcher.GetBuckets(ctx, shared.UserID, startTime, endTime)
			if err != nil {
				return err
			}

			err = buckets.Update(shared.UserID, shared.Type, userData)
			if err != nil {
				return err
			}

			err = bucketsFetcher.WriteModifiedBuckets(ctx, shared.Dates.LastUpdatedDate, buckets)
			if err != nil {
				return err
			}

		}
	}

	allBuckets, err := bucketsFetcher.GetAllBuckets(ctx, shared.UserID)
	if err != nil {
		return err
	}
	defer allBuckets.Close(ctx)

	err = s.CalculateSummary(ctx, allBuckets)
	if err != nil {
		return err
	}

	return nil
}

func (s *ContinuousStats) CalculateSummary(ctx context.Context, buckets fetcher.AnyCursor) error {
	// count backwards (newest first) through hourly stats, stopping at 1d, 7d, 14d, 30d,
	// currently only supports day precision
	nextStopPoint := 0
	totalStats := ContinuousPeriod{}
	var err error
	var stopPoints []time.Time

	bucket := &Bucket[*ContinuousBucket, ContinuousBucket]{}

	for buckets.Next(ctx) {
		if err = buckets.Decode(bucket); err != nil {
			return err
		}

		// We should have the newest (last) bucket here, use its date for breakpoints
		if stopPoints == nil {
			stopPoints, _ = calculateStopPoints(bucket.Time)
		}

		if bucket.Data.Total.Records == 0 {
			panic("bucket exists with 0 records")
		}

		s.TotalHours++

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			totalStats.Update(bucket)

			if bucket.Time.Before(stopPoints[nextStopPoint]) {
				s.CalculatePeriod(periodLengths[nextStopPoint], false, totalStats)
				nextStopPoint++
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
