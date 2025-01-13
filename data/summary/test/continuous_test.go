package test_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("Continuous", func() {
	var userId string
	var bucketTime time.Time
	var err error

	BeforeEach(func() {
		now := time.Now()
		userId = "1234"
		bucketTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
	})

	Context("Range", func() {
		// range has no direct functions for continuous, but if it does, tests here.
	})

	Context("Ranges", func() {
		It("ranges.Add", func() {
			firstRange := types.ContinuousRanges{
				Realtime: types.Range{
					Records: 5,
				},
				Deferred: types.Range{
					Records: 10,
				},
				Total: types.Range{
					Records: 12,
				},
			}

			secondRange := types.ContinuousRanges{
				Realtime: types.Range{
					Records: 3,
				},
				Deferred: types.Range{
					Records: 11,
				},
				Total: types.Range{
					Records: 13,
				},
			}
			firstRange.Add(&secondRange)

			Expect(firstRange.Realtime.Records).To(Equal(8))
			Expect(firstRange.Deferred.Records).To(Equal(21))
			Expect(firstRange.Total.Records).To(Equal(25))
		})

		It("ranges.Finalize", func() {
			continuousRange := types.ContinuousRanges{
				Realtime: types.Range{
					Records: 5,
				},
				Deferred: types.Range{
					Records: 10,
				},
				Total: types.Range{
					Records: 10,
				},
			}

			continuousRange.Finalize()

			Expect(continuousRange.Realtime.Percent).To(Equal(0.5))
			Expect(continuousRange.Deferred.Percent).To(Equal(1.0))

		})
	})

	Context("bucket.Update", func() {
		var userBucket *types.Bucket[*types.ContinuousBucket, types.ContinuousBucket]
		var continuousDatum data.Datum

		It("With a realtime value", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			continuousDatum = NewRealtimeGlucose(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(continuousDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Total.Records).To(Equal(1))
			Expect(userBucket.Data.Deferred.Records).To(Equal(0))
			Expect(userBucket.Data.Realtime.Records).To(Equal(1))
			Expect(userBucket.IsModified()).To(BeTrue())

			err = userBucket.Update(continuousDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.Data.Total.Records).To(Equal(2))
			Expect(userBucket.Data.Deferred.Records).To(Equal(0))
			Expect(userBucket.Data.Realtime.Records).To(Equal(2))
		})

		It("With a deferred value", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			continuousDatum = NewDeferredGlucose(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(continuousDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Total.Records).To(Equal(1))
			Expect(userBucket.Data.Deferred.Records).To(Equal(1))
			Expect(userBucket.Data.Realtime.Records).To(Equal(0))
			Expect(userBucket.IsModified()).To(BeTrue())

			err = userBucket.Update(continuousDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.Data.Total.Records).To(Equal(2))
			Expect(userBucket.Data.Deferred.Records).To(Equal(2))
			Expect(userBucket.Data.Realtime.Records).To(Equal(0))

		})
	})

	Context("period", func() {
		var period types.ContinuousPeriod

		It("Add single bucket to an empty period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.ContinuousPeriod{}

			bucketOne := types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewRealtimeGlucose(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Realtime.Records).To(Equal(1))
			Expect(period.Deferred.Records).To(Equal(0))
			Expect(period.Total.Records).To(Equal(1))
		})

		It("Add duplicate buckets to a period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.ContinuousPeriod{}

			bucketOne := types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewRealtimeGlucose(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Realtime.Records).To(Equal(1))

			err = period.Update(bucketOne)
			Expect(err).To(HaveOccurred())
			Expect(period.Realtime.Records).To(Equal(1))
		})

		It("Add two buckets to an empty period on 2 different hours", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.ContinuousPeriod{}

			bucketOne := types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewRealtimeGlucose(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			bucketTwo := types.NewBucket[*types.ContinuousBucket](userId, bucketTime.Add(time.Hour), types.SummaryTypeCGM)
			err = bucketTwo.Update(NewRealtimeGlucose(continuous.Type, datumTime.Add(time.Hour), InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Realtime.Records).To(Equal(1))
			Expect(period.Deferred.Records).To(Equal(0))
			Expect(period.Total.Records).To(Equal(1))

			err = period.Update(bucketTwo)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Realtime.Records).To(Equal(2))
			Expect(period.Deferred.Records).To(Equal(0))
			Expect(period.Total.Records).To(Equal(2))
		})

		It("Finalize a 1d period", func() {
			period = types.ContinuousPeriod{}
			buckets := CreateContinuousBuckets(bucketTime, 24, 12)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(1)

			// data is generated at 100% per range
			Expect(period.Realtime.Percent).To(Equal(1.0))
			Expect(period.Deferred.Percent).To(Equal(1.0))
			Expect(period.AverageDailyRecords).To(Equal(12.0 * 24.0))
		})

		It("Finalize a 7d period", func() {
			period = types.ContinuousPeriod{}
			buckets := CreateContinuousBuckets(bucketTime, 24*5, 12)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(7)

			// data is generated at 100% per range
			Expect(period.Realtime.Percent).To(Equal(1.0))
			Expect(period.Deferred.Percent).To(Equal(1.0))
			Expect(period.AverageDailyRecords).To(Equal((12.0 * 24.0) * 5 / 7))
		})

		It("Update a finalized period", func() {
			period = types.ContinuousPeriod{}
			period.Finalize(14)

			bucket := types.NewBucket[*types.ContinuousBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = period.Update(bucket)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ContinuousStats", func() {
		var logger log.Logger
		var ctx context.Context

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		It("Init", func() {
			s := types.ContinuousStats{}
			s.Init()

			Expect(s.Periods).ToNot(BeNil())
		})

		Context("CalculateSummary", func() {

			It("CalculateSummary 1d", func() {
				s := types.ContinuousStats{}
				s.Init()

				buckets := CreateContinuousBuckets(bucketTime, 24, 1)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.Periods["7d"].Total.Records).To(Equal(24))
				Expect(s.Periods["14d"].Total.Records).To(Equal(24))
				Expect(s.Periods["30d"].Total.Records).To(Equal(24))
			})

			It("CalculateSummary 2d", func() {
				s := types.ContinuousStats{}
				s.Init()

				buckets := CreateContinuousBuckets(bucketTime, 48, 1)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 2))
				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 2))
				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 2))
			})

			It("CalculateSummary 7d", func() {
				s := types.ContinuousStats{}
				s.Init()

				buckets := CreateContinuousBuckets(bucketTime, 24*7, 1)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 7))
				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 7))
			})

			It("CalculateSummary 14d", func() {
				s := types.ContinuousStats{}
				s.Init()

				buckets := CreateContinuousBuckets(bucketTime, 24*14, 1)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 14))
			})

			It("CalculateSummary 30d", func() {
				s := types.ContinuousStats{}
				s.Init()

				buckets := CreateContinuousBuckets(bucketTime, 24*30, 1)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 30))
			})
		})
	})
})
