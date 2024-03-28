package types_test

import (
	"context"
	"fmt"
	"time"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/test"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"

	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/summary/types"

	"github.com/tidepool-org/platform/data/types/blood/glucose"

	"github.com/tidepool-org/platform/pointer"
)

func NewDataSetDataRealtime(t string, startTime time.Time, hours float64, realtime bool) []*glucose.Glucose {
	requiredRecords := int(hours * 2)
	typ := pointer.FromAny(t)

	var dataSetData = make([]*glucose.Glucose, requiredRecords)
	var glucoseValue = inTargetBloodGlucose
	var deviceId = "SummaryTestDevice"
	var uploadId = test.RandomSetID()

	// generate X hours of data
	for count := 0; count < requiredRecords; count += 1 {
		datumTime := startTime.Add(time.Duration(count-requiredRecords) * time.Minute * 30)

		datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceId, &uploadId)
		datum.Value = pointer.FromFloat64(glucoseValue)

		if realtime {
			datum.CreatedTime = pointer.FromAny(datumTime.Add(5 * time.Minute))
			datum.ModifiedTime = pointer.FromAny(datumTime.Add(10 * time.Minute))
		}

		dataSetData[count] = datum
	}

	return dataSetData
}

var _ = Describe("Continuous Summary", func() {
	var userId string
	var datumTime time.Time
	var logger log.Logger
	var ctx context.Context
	var err error
	var dataSetCGMData []*glucose.Glucose

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		userId = userTest.RandomID()
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("Create Summary", func() {
		It("Correctly initializes a summary", func() {
			summary := types.Create[*types.ContinuousStats](userId)
			Expect(summary).ToNot(BeNil())
			Expect(summary.Type).To(Equal("continuous"))

			Expect(summary.UserID).To(Equal(userId))
			Expect(summary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userContinuousSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]

		Context("AddData Bucket Testing", func() {
			It("Returns correct hour count when given 2 weeks", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 336, inTargetBloodGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct hour count when given 1 week", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct hour count when given 3 weeks", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 504, inTargetBloodGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct records when given >60d of data", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)

				dataSetCGMData = NewDataSetCGMDataRanges(datumTime, 5, NewDataRangesSingle(lowBloodGlucose-0.5))
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(10))

				dataSetCGMData = NewDataSetCGMDataRanges(datumTime.Add(1*time.Hour), 1, NewDataRangesSingle(highBloodGlucose+0.5))
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(10))

				dataSetCGMData = NewDataSetCGMDataRanges(datumTime.Add(24*60*time.Hour), 1, NewDataRangesSingle(inTargetBloodGlucose-0.5))
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Stats.Buckets[0].Data.HighRecords).To(Equal(10))

				for i := 0; i < len(userContinuousSummary.Stats.Buckets); i++ {
					Expect(userContinuousSummary.Stats.Buckets[i]).ToNot(BeNil())
				}
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(1440))
			})

			It("Returns correct records when given data a full 60d ahead of previous data", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)

				dataSetCGMData = NewDataSetCGMDataRanges(datumTime, 1, NewDataRangesSingle(lowBloodGlucose-0.5))
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataRanges(datumTime.Add(24*62*time.Hour), 1, NewDataRangesSingle(inTargetBloodGlucose-0.5))
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < len(userContinuousSummary.Stats.Buckets); i++ {
					Expect(userContinuousSummary.Stats.Buckets[i]).ToNot(BeNil())
				}
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(1))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := lowBloodGlucose
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userContinuousSummary.Stats.Buckets)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", inTargetBloodGlucose*12*5, 0.001))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetCGMData = NewDataSetCGMDataAvg(secondDatumTime, 168, secondRequestedAvgGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(528)) // 22 days

				By("check total glucose and dates for second batch")
				newHourlyStatsLen = len(userContinuousSummary.Stats.Buckets)
				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetCGMData)/12 // 12 per day, need length without the gap
				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", secondRequestedAvgGlucose*12*5))

					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("check total glucose and dates for gap")
				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64(0)))
				}
			})

			It("Returns correct stats when given multiple batches in a day", func() {
				var incrementalDatumTime time.Time
				var lastRecordTime time.Time
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 144, inTargetBloodGlucose)
				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(144))

				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetCGMData = NewDataSetCGMDataAvg(incrementalDatumTime, 1, float64(i))

					err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMData)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(144 + i))
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
				}

				for i := 144; i < len(userContinuousSummary.Stats.Buckets); i++ {
					f := fmt.Sprintf("hour %d", i)
					By(f)
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*5)
					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", float64((i-143)*12*5), 0.001))

					averageGlucoseMmol := userContinuousSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userContinuousSummary.Stats.Buckets[i].Data.TotalMinutes)
					Expect(averageGlucoseMmol).To(BeNumerically("~", i-143))
				}
			})

			It("Returns correct hourly stats for days uploaded in reverse", func() {
				var lastRecordTime time.Time
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)

				// Datasets use +1 and +2 offset to allow for checking via iteration
				dataSetCGMDataOne := NewDataSetCGMDataAvg(datumTime.AddDate(0, 0, -2), 24, inTargetBloodGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(datumTime.AddDate(0, 0, -1), 24, inTargetBloodGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(datumTime, 24, inTargetBloodGlucose+2)

				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMDataThree)
				Expect(err).ToNot(HaveOccurred())

				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMDataTwo)
				Expect(err).ToNot(HaveOccurred())

				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetCGMDataOne)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(72))

				for i := len(userContinuousSummary.Stats.Buckets) - 1; i >= 0; i-- {
					By(fmt.Sprintf("hour %d", i+1))
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userContinuousSummary.Stats.Buckets)-i-1) - 5*time.Minute)
					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}
			})
		})

		Context("ClearInvalidatedBuckets", func() {
			var dataStore types.DeviceDataFetcher

			BeforeEach(func() {
				config := storeStructuredMongoTest.NewConfig()
				store, err := dataStoreMongo.NewStore(config)
				Expect(err).ToNot(HaveOccurred())
				dataStore = store.NewDataRepository()
			})

			It("trims the correct buckets", func() {
				var dataSetCGMDataCursor types.DeviceDataCursor
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 10, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userContinuousSummary.Stats.Update(ctx, dataSetCGMDataCursor, dataStore)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(10))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(-5 * time.Hour))

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(5))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1]).ToNot(BeNil())

				Expect(firstData).To(Equal(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1].LastRecordTime))
			})

			It("trims the all buckets with data beyond the beginning of the buckets", func() {
				var dataSetCGMDataCursor types.DeviceDataCursor
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 10, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userContinuousSummary.Stats.Update(ctx, dataSetCGMDataCursor, dataStore)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(10))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(-15 * time.Hour))

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(0))

				Expect(firstData.IsZero()).To(BeTrue())
			})

			It("doesnt trim if only modified in the future", func() {
				var dataSetCGMDataCursor types.DeviceDataCursor
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(datumTime, 10, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userContinuousSummary.Stats.Update(ctx, dataSetCGMDataCursor, dataStore)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(10))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(time.Hour))

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1]).ToNot(BeNil())

				Expect(firstData).To(Equal(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1].LastRecordTime))
			})

			It("doesnt trim if only modified on the same hour, but after the bucket time", func() {
				var dataSetCGMDataCursor types.DeviceDataCursor
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				midDatumTime := datumTime.Add(30 * time.Minute)
				dataSetCGMData = NewDataSetCGMDataAvg(midDatumTime, 9, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userContinuousSummary.Stats.Update(ctx, dataSetCGMDataCursor, dataStore)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(10))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(midDatumTime.Add(10 * time.Minute))

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1]).ToNot(BeNil())

				Expect(firstData).To(Equal(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1].LastRecordTime))
			})

			It("trims if modified on the same hour, and before the bucket time", func() {
				var dataSetCGMDataCursor types.DeviceDataCursor
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				midDatumTime := datumTime.Add(30 * time.Minute)
				dataSetCGMData = NewDataSetCGMDataAvg(midDatumTime, 9, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userContinuousSummary.Stats.Update(ctx, dataSetCGMDataCursor, dataStore)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(10))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(10))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(midDatumTime.Add(-10 * time.Minute))

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(9))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1]).ToNot(BeNil())

				Expect(firstData).To(Equal(userContinuousSummary.Stats.Buckets[len(userContinuousSummary.Stats.Buckets)-1].LastRecordTime))
			})

			It("successfully does nothing if there are no buckets", func() {
				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(0))
				Expect(userContinuousSummary.Stats.TotalHours).To(Equal(0))

				firstData := userContinuousSummary.Stats.ClearInvalidatedBuckets(datumTime)

				// we have the right length
				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(0))

				Expect(firstData.IsZero()).To(BeTrue())
			})
		})
	})
})
