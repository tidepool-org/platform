package test_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("Glucose", func() {
	var bucketTime time.Time
	var err error
	var userId string

	BeforeEach(func() {
		now := time.Now()
		userId = "1234"
		bucketTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	})

	Context("Ranges", func() {
		It("range.Update", func() {
			glucoseRange := types.Range{}

			By("adding 5 minutes of 5mmol")
			glucoseRange.Update(5, 5, true)
			Expect(glucoseRange.Glucose).To(Equal(5.0 * 5.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(5))
			Expect(glucoseRange.Variance).To(Equal(0.0))

			By("adding 1 minute of 10mmol")
			glucoseRange.Update(10, 1, true)
			Expect(glucoseRange.Glucose).To(Equal(5.0*5.0 + 10.0))
			Expect(glucoseRange.Records).To(Equal(2))
			Expect(glucoseRange.Minutes).To(Equal(6))
			Expect(glucoseRange.Variance).To(Equal(20.833333333333336))
		})

		It("range.Update without minutes", func() {
			glucoseRange := types.Range{}

			By("adding 1 record of 5mmol")
			glucoseRange.Update(5, 0, true)
			Expect(glucoseRange.Glucose).To(Equal(5.0))
			Expect(glucoseRange.Records).To(Equal(1))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))

			By("adding 1 record of 10mmol")
			glucoseRange.Update(10, 0, true)
			Expect(glucoseRange.Glucose).To(Equal(15.0))
			Expect(glucoseRange.Records).To(Equal(2))
			Expect(glucoseRange.Minutes).To(Equal(0))
			Expect(glucoseRange.Variance).To(Equal(0.0))
		})

		It("range.Add", func() {
			firstRange := types.Range{
				Glucose:  5,
				Minutes:  5,
				Records:  5,
				Percent:  5,
				Variance: 5,
			}

			secondRange := types.Range{
				Glucose:  10,
				Minutes:  10,
				Records:  10,
				Percent:  10,
				Variance: 10,
			}

			firstRange.Add(&secondRange)

			Expect(firstRange.Glucose).To(Equal(15.0))
			Expect(firstRange.Minutes).To(Equal(15))
			Expect(firstRange.Records).To(Equal(15))
			Expect(firstRange.Variance).To(Equal(15.0))

			// expect percent untouched, we don't handle percent on add
			Expect(firstRange.Percent).To(Equal(5.0))
		})

	})

	Context("bucket.Update", func() {
		var userBucket *types.Bucket[*types.GlucoseBucket, types.GlucoseBucket]
		var cgmDatum data.Datum

		It("With a fresh bucket", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.IsModified()).To(BeTrue())

			Expect(userBucket.Data.Target.Records).To(Equal(userBucket.Data.Total.Records))
			Expect(userBucket.Data.Target.Minutes).To(Equal(userBucket.Data.Total.Minutes))
		})

		It("With a bgm value", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeBGM)
			bgmDatum := NewGlucoseWithValue(selfmonitored.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(bgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(0))
			Expect(userBucket.IsModified()).To(BeTrue())

			Expect(userBucket.Data.Target.Records).To(Equal(userBucket.Data.Total.Records))
			Expect(userBucket.Data.Target.Minutes).To(Equal(userBucket.Data.Total.Minutes))
		})

		It("With a smbg value in a cgm bucket", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			bgmDatum := NewGlucoseWithValue(selfmonitored.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(bgmDatum)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("record for cgm calculation is of invald type smbg"))
		})

		It("With a cbg value in a bgm bucket", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeBGM)
			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(cgmDatum)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("record for bgm calculation is of invald type cbg"))
		})

		It("With two values in a range", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)

			By("Inserting the first data")

			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)
			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.IsModified()).To(BeTrue())

			Expect(userBucket.Data.Target.Records).To(Equal(userBucket.Data.Total.Records))
			Expect(userBucket.Data.Target.Minutes).To(Equal(userBucket.Data.Total.Minutes))

			secondDatumTime := datumTime.Add(5 * time.Minute)
			cgmDatum = NewGlucoseWithValue(continuous.Type, secondDatumTime, InTargetBloodGlucose)

			By("Inserting the second data")

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.LastData).To(Equal(secondDatumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(2))
			Expect(userBucket.Data.Target.Minutes).To(Equal(10))
			Expect(userBucket.IsModified()).To(BeTrue())

			Expect(userBucket.Data.Target.Records).To(Equal(userBucket.Data.Total.Records))
			Expect(userBucket.Data.Target.Minutes).To(Equal(userBucket.Data.Total.Minutes))

		})

		It("With values in all ranges", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)

			ranges := map[float64]*types.Range{
				VeryLowBloodGlucose - 0.1:     &userBucket.Data.VeryLow,
				LowBloodGlucose - 0.1:         &userBucket.Data.Low,
				InTargetBloodGlucose + 0.1:    &userBucket.Data.Target,
				HighBloodGlucose + 0.1:        &userBucket.Data.High,
				ExtremeHighBloodGlucose + 0.1: &userBucket.Data.ExtremeHigh,
			}

			expectedGlucose := 0.0
			expectedMinutes := 0
			expectedRecords := 0

			expectedAnyLowGlucose := 0.0
			expectedAnyLowMinutes := 0
			expectedAnyLowRecords := 0

			expectedAnyHighGlucose := 0.0
			expectedAnyHighMinutes := 0
			expectedAnyHighRecords := 0

			expectedVeryHighGlucose := 0.0
			expectedVeryHighMinutes := 0
			expectedVeryHighRecords := 0

			for k, v := range ranges {
				By(fmt.Sprintf("Add a value of %f", k))
				Expect(v.Records).To(BeZero())
				Expect(v.Glucose).To(BeZero())
				Expect(v.Minutes).To(BeZero())

				cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, k)
				err = userBucket.Update(cgmDatum)
				Expect(err).ToNot(HaveOccurred())

				Expect(v.Records).To(Equal(1))
				Expect(v.Minutes).To(Equal(5))

				expectedGlucose += k * 5
				expectedMinutes += 5
				expectedRecords++
				Expect(userBucket.Data.Total.Records).To(Equal(expectedRecords))
				Expect(userBucket.Data.Total.Glucose).To(Equal(expectedGlucose))
				Expect(userBucket.Data.Total.Minutes).To(Equal(expectedMinutes))

				if k < LowBloodGlucose {
					expectedAnyLowGlucose += k * 5
					expectedAnyLowMinutes += 5
					expectedAnyLowRecords++
				}
				Expect(userBucket.Data.AnyLow.Records).To(Equal(expectedAnyLowRecords))
				Expect(userBucket.Data.AnyLow.Minutes).To(Equal(expectedAnyLowMinutes))

				if k > HighBloodGlucose {
					expectedAnyHighGlucose += k * 5
					expectedAnyHighMinutes += 5
					expectedAnyHighRecords++
				}
				Expect(userBucket.Data.AnyHigh.Records).To(Equal(expectedAnyHighRecords))
				Expect(userBucket.Data.AnyHigh.Minutes).To(Equal(expectedAnyHighMinutes))

				if k > VeryHighBloodGlucose {
					expectedVeryHighGlucose += k * 5
					expectedVeryHighMinutes += 5
					expectedVeryHighRecords++
				}
				Expect(userBucket.Data.VeryHigh.Records).To(Equal(expectedVeryHighRecords))
				Expect(userBucket.Data.VeryHigh.Minutes).To(Equal(expectedVeryHighMinutes))

				// we should check that total gets variance
			}
		})
	})

	Context("bucketsByTime.Update", func() {
		var userBuckets types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]
		var cgmDatums []data.Datum

		It("With no existing buckets", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]{}
			cgmDatums = []data.Datum{NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)}

			err = userBuckets.Update(userId, types.SummaryTypeCGM, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
		})

		It("Adding to existing buckets", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]{}
			cgmDatums = []data.Datum{NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)}

			err = userBuckets.Update(userId, types.SummaryTypeCGM, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))

			err = userBuckets.Update(userId, types.SummaryTypeCGM, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(2))
		})

		It("Adding to two different buckets at once", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]{}
			cgmDatums = []data.Datum{
				NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose),
				NewGlucoseWithValue(continuous.Type, datumTime.Add(time.Hour), LowBloodGlucose-0.1),
			}

			err = userBuckets.Update(userId, types.SummaryTypeCGM, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))
			Expect(userBuckets[bucketTime.Add(time.Hour)].Data.Low.Records).To(Equal(1))
		})

		It("Adding two records to the same bucket at once", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]{}
			cgmDatums = []data.Datum{
				NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose),
				NewGlucoseWithValue(continuous.Type, datumTime, LowBloodGlucose-0.1),
			}

			err = userBuckets.Update(userId, types.SummaryTypeCGM, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))
			Expect(userBuckets[bucketTime].Data.Low.Records).To(Equal(1))
		})
	})

	Context("period", func() {
		var period types.GlucosePeriod

		It("Add single bucket to an empty period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.GlucosePeriod{}

			bucketOne := types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))
		})

		It("Add duplicate buckets to a period", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.GlucosePeriod{}

			bucketOne := types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))

			err = period.Update(bucketOne)
			Expect(err).To(HaveOccurred())
		})

		It("Add three buckets to an empty period on 2 different days, 3 different hours", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			period = types.GlucosePeriod{}

			bucketOne := types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = bucketOne.Update(NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			bucketTwo := types.NewBucket[*types.GlucoseBucket](userId, bucketTime.Add(time.Hour), types.SummaryTypeCGM)
			err = bucketTwo.Update(NewGlucoseWithValue(continuous.Type, datumTime.Add(time.Hour), InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			bucketThree := types.NewBucket[*types.GlucoseBucket](userId, bucketTime.Add(24*time.Hour), types.SummaryTypeCGM)
			err = bucketThree.Update(NewGlucoseWithValue(continuous.Type, datumTime.Add(24*time.Hour), InTargetBloodGlucose))
			Expect(err).ToNot(HaveOccurred())

			err = period.Update(bucketOne)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(1))
			Expect(period.HoursWithData).To(Equal(1))
			Expect(period.DaysWithData).To(Equal(1))

			err = period.Update(bucketTwo)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(2))
			Expect(period.HoursWithData).To(Equal(2))
			Expect(period.DaysWithData).To(Equal(1))

			err = period.Update(bucketThree)
			Expect(err).ToNot(HaveOccurred())
			Expect(period.Target.Records).To(Equal(3))
			Expect(period.HoursWithData).To(Equal(3))
			Expect(period.DaysWithData).To(Equal(2))
		})

		It("Finalize a 1d period", func() {
			period = types.GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 24, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(1)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 24.0))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(types.CalculateGMI(InTargetBloodGlucose)))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 7d period", func() {
			period = types.GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 24*5, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(7)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(1.0))
			Expect(period.AnyLow.Percent).To(Equal(1.0))
			Expect(period.AnyHigh.Percent).To(Equal(1.0))
			Expect(period.Target.Percent).To(Equal(1.0))
			Expect(period.Low.Percent).To(Equal(1.0))
			Expect(period.High.Percent).To(Equal(1.0))
			Expect(period.VeryLow.Percent).To(Equal(1.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(1.0))

			Expect(period.AverageDailyRecords).To(Equal((12.0 * 24.0) * 5 / 7))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(types.CalculateGMI(InTargetBloodGlucose)))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 1d period with insufficient data", func() {
			period = types.GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 16, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(1)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(0.0))
			Expect(period.AnyLow.Percent).To(Equal(0.0))
			Expect(period.AnyHigh.Percent).To(Equal(0.0))
			Expect(period.Target.Percent).To(Equal(0.0))
			Expect(period.Low.Percent).To(Equal(0.0))
			Expect(period.High.Percent).To(Equal(0.0))
			Expect(period.VeryLow.Percent).To(Equal(0.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(0.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 16.0))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(0.0))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Finalize a 7d period with insufficient data", func() {
			period = types.GlucosePeriod{}
			buckets := CreateGlucoseBuckets(bucketTime, 23, 12, true)

			for i := range buckets {
				err = period.Update(buckets[i])
				Expect(err).ToNot(HaveOccurred())
			}

			period.Finalize(7)

			// data is generated at 100% per range
			Expect(period.VeryHigh.Percent).To(Equal(0.0))
			Expect(period.AnyLow.Percent).To(Equal(0.0))
			Expect(period.AnyHigh.Percent).To(Equal(0.0))
			Expect(period.Target.Percent).To(Equal(0.0))
			Expect(period.Low.Percent).To(Equal(0.0))
			Expect(period.High.Percent).To(Equal(0.0))
			Expect(period.VeryLow.Percent).To(Equal(0.0))
			Expect(period.ExtremeHigh.Percent).To(Equal(0.0))

			Expect(period.AverageDailyRecords).To(Equal(12.0 * 23.0 / 7))
			Expect(period.AverageGlucose).To(Equal(InTargetBloodGlucose))
			Expect(period.GlucoseManagementIndicator).To(Equal(0.0))

			// we only validate these are set here, as this requires more specific validation
			Expect(period.StandardDeviation).ToNot(Equal(0.0))
			Expect(period.CoefficientOfVariation).ToNot(Equal(0.0))
		})

		It("Update a finalized period", func() {
			period = types.GlucosePeriod{}
			period.Finalize(14)

			bucket := types.NewBucket[*types.GlucoseBucket](userId, bucketTime, types.SummaryTypeCGM)
			err = period.Update(bucket)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("GlucoseStats", func() {
		var logger log.Logger
		var ctx context.Context

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		It("Init", func() {
			s := types.GlucoseStats{}
			s.Init()

			Expect(s.Periods).ToNot(BeNil())
			Expect(s.OffsetPeriods).ToNot(BeNil())
		})

		Context("CalculateSummary", func() {

			It("CalculateSummary 1d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(0))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(0))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(0))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 2d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 48, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 2))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(0))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 2))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(0))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 2))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 7d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*7, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(0))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(0))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 14d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*14, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(24 * 7))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(0))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 28d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*28, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(24 * 7))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(24 * 14))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 28))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 30d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*30, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(24 * 7))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(24 * 14))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 30))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(0))
			})

			It("CalculateSummary 60d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*60, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(24 * 7))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(24 * 14))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 30))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(24 * 30))
			})

			It("CalculateSummary 61d", func() {
				s := types.GlucoseStats{}
				s.Init()

				buckets := CreateGlucoseBuckets(bucketTime, 24*61, 1, true)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods).To(Not(BeNil()))
				Expect(s.OffsetPeriods).To(Not(BeNil()))

				Expect(s.Periods["1d"].Total.Records).To(Equal(24))
				Expect(s.OffsetPeriods["1d"].Total.Records).To(Equal(24))

				Expect(s.Periods["7d"].Total.Records).To(Equal(24 * 7))
				Expect(s.OffsetPeriods["7d"].Total.Records).To(Equal(24 * 7))

				Expect(s.Periods["14d"].Total.Records).To(Equal(24 * 14))
				Expect(s.OffsetPeriods["14d"].Total.Records).To(Equal(24 * 14))

				Expect(s.Periods["30d"].Total.Records).To(Equal(24 * 30))
				Expect(s.OffsetPeriods["30d"].Total.Records).To(Equal(24 * 30))
			})
		})

		Context("CalculateDelta", func() {

			It("CalculateDelta populates all values", func() {
				// This validates a large block of easy to typo function calls in CalculateDelta, apologies to whoever has
				// to update this.
				s := types.GlucoseStats{
					Periods: types.GlucosePeriods{"1d": &types.GlucosePeriod{
						GlucoseRanges: types.GlucoseRanges{
							Total: types.Range{
								Glucose:  0,
								Minutes:  0,
								Records:  0,
								Percent:  0,
								Variance: 0,
							},
							VeryLow: types.Range{
								Glucose:  1,
								Minutes:  1,
								Records:  1,
								Percent:  1,
								Variance: 1,
							},
							Low: types.Range{
								Glucose:  2,
								Minutes:  2,
								Records:  2,
								Percent:  2,
								Variance: 2,
							},
							Target: types.Range{
								Glucose:  3,
								Minutes:  3,
								Records:  3,
								Percent:  3,
								Variance: 3,
							},
							High: types.Range{
								Glucose:  4,
								Minutes:  4,
								Records:  4,
								Percent:  4,
								Variance: 4,
							},
							VeryHigh: types.Range{
								Glucose:  5,
								Minutes:  5,
								Records:  5,
								Percent:  5,
								Variance: 5,
							},
							ExtremeHigh: types.Range{
								Glucose:  6,
								Minutes:  6,
								Records:  6,
								Percent:  6,
								Variance: 6,
							},
							AnyLow: types.Range{
								Glucose:  7,
								Minutes:  7,
								Records:  7,
								Percent:  7,
								Variance: 7,
							},
							AnyHigh: types.Range{
								Glucose:  8,
								Minutes:  8,
								Records:  8,
								Percent:  8,
								Variance: 8,
							},
						},
						HoursWithData:              0,
						DaysWithData:               1,
						AverageGlucose:             2,
						GlucoseManagementIndicator: 3,
						CoefficientOfVariation:     4,
						StandardDeviation:          5,
						AverageDailyRecords:        6,
					}},
					OffsetPeriods: types.GlucosePeriods{"1d": &types.GlucosePeriod{
						GlucoseRanges: types.GlucoseRanges{
							Total: types.Range{
								Glucose:  99,
								Minutes:  98,
								Records:  97,
								Percent:  96,
								Variance: 95,
							},
							VeryLow: types.Range{
								Glucose:  89,
								Minutes:  88,
								Records:  87,
								Percent:  86,
								Variance: 85,
							},
							Low: types.Range{
								Glucose:  79,
								Minutes:  78,
								Records:  77,
								Percent:  76,
								Variance: 75,
							},
							Target: types.Range{
								Glucose:  69,
								Minutes:  68,
								Records:  67,
								Percent:  66,
								Variance: 65,
							},
							High: types.Range{
								Glucose:  59,
								Minutes:  58,
								Records:  57,
								Percent:  56,
								Variance: 55,
							},
							VeryHigh: types.Range{
								Glucose:  49,
								Minutes:  48,
								Records:  47,
								Percent:  46,
								Variance: 45,
							},
							ExtremeHigh: types.Range{
								Glucose:  39,
								Minutes:  38,
								Records:  37,
								Percent:  36,
								Variance: 35,
							},
							AnyLow: types.Range{
								Glucose:  29,
								Minutes:  28,
								Records:  27,
								Percent:  26,
								Variance: 25,
							},
							AnyHigh: types.Range{
								Glucose:  19,
								Minutes:  18,
								Records:  17,
								Percent:  16,
								Variance: 15,
							},
						},
						HoursWithData:              99,
						DaysWithData:               98,
						AverageGlucose:             97,
						GlucoseManagementIndicator: 96,
						CoefficientOfVariation:     95,
						StandardDeviation:          94,
						AverageDailyRecords:        93,
					}},
				}

				s.CalculateDelta()

				expectedDelta := types.GlucosePeriod{
					GlucoseRanges: types.GlucoseRanges{
						Total: types.Range{
							Minutes: -98,
							Records: -97,
							Percent: -96,
						},
						VeryLow: types.Range{
							Minutes: -87,
							Records: -86,
							Percent: -85,
						},
						Low: types.Range{
							Minutes: -76,
							Records: -75,
							Percent: -74,
						},
						Target: types.Range{
							Minutes: -65,
							Records: -64,
							Percent: -63,
						},
						High: types.Range{
							Minutes: -54,
							Records: -53,
							Percent: -52,
						},
						VeryHigh: types.Range{
							Minutes: -43,
							Records: -42,
							Percent: -41,
						},
						ExtremeHigh: types.Range{
							Minutes: -32,
							Records: -31,
							Percent: -30,
						},
						AnyLow: types.Range{
							Minutes: -21,
							Records: -20,
							Percent: -19,
						},
						AnyHigh: types.Range{
							Minutes: -10,
							Records: -9,
							Percent: -8,
						},
					},
					HoursWithData:              -99,
					DaysWithData:               -97,
					AverageGlucose:             -95,
					GlucoseManagementIndicator: -93,
					CoefficientOfVariation:     -91,
					StandardDeviation:          -89,
					AverageDailyRecords:        -87,
				}

				opts := cmpopts.IgnoreUnexported(types.GlucosePeriod{})
				Expect(*(s.Periods["1d"].Delta)).To(BeComparableTo(expectedDelta, opts))

			})

			It("CalculateDelta 1d", func() {
				s := types.GlucoseStats{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -1), 24, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods["1d"].Delta).To(BeNil())
				Expect(s.OffsetPeriods["1d"].Delta).To(BeNil())

				s.CalculateDelta()

				Expect(s.Periods["1d"].Delta.Total.Records).To(Equal(-24))
				Expect(s.OffsetPeriods["1d"].Delta.Total.Records).To(Equal(24))
			})

			It("CalculateDelta 7d", func() {
				s := types.GlucoseStats{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*7, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -7), 24*7, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods["7d"].Delta).To(BeNil())
				Expect(s.OffsetPeriods["7d"].Delta).To(BeNil())

				s.CalculateDelta()

				Expect(s.Periods["7d"].Delta.Total.Records).To(Equal(-24 * 7))
				Expect(s.OffsetPeriods["7d"].Delta.Total.Records).To(Equal(24 * 7))
			})

			It("CalculateDelta 14d", func() {
				s := types.GlucoseStats{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*14, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -14), 24*14, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods["14d"].Delta).To(BeNil())
				Expect(s.OffsetPeriods["14d"].Delta).To(BeNil())

				s.CalculateDelta()

				Expect(s.Periods["14d"].Delta.Total.Records).To(Equal(-24 * 14))
				Expect(s.OffsetPeriods["14d"].Delta.Total.Records).To(Equal(24 * 14))
			})

			It("CalculateDelta 30d", func() {
				s := types.GlucoseStats{}
				s.Init()

				bucketsOne := CreateGlucoseBuckets(bucketTime, 24*30, 1, true)
				bucketsTwo := CreateGlucoseBuckets(bucketTime.AddDate(0, 0, -30), 24*30, 2, true)
				buckets := append(bucketsOne, bucketsTwo...)
				bucketsCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(buckets), nil, nil)
				Expect(err).ToNot(HaveOccurred())

				err = s.CalculateSummary(ctx, bucketsCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(s.Periods["30d"].Delta).To(BeNil())
				Expect(s.OffsetPeriods["30d"].Delta).To(BeNil())

				s.CalculateDelta()

				Expect(s.Periods["30d"].Delta.Total.Records).To(Equal(-24 * 30))
				Expect(s.OffsetPeriods["30d"].Delta.Total.Records).To(Equal(24 * 30))
			})
		})
	})
})
