package test_test

//import (
//	"fmt"
//	"time"
//
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//
//	"github.com/tidepool-org/platform/data"
//	"github.com/tidepool-org/platform/data/summary/types"
//	summaryTest "github.com/tidepool-org/platform/data/summary/types/test"
//	"github.com/tidepool-org/platform/data/test"
//	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
//	"github.com/tidepool-org/platform/pointer"
//	userTest "github.com/tidepool-org/platform/user/test"
//)
//
//func NewDataSetDataRealtime(typ string, startTime time.Time, hours float64, realtime bool) []data.Datum {
//	requiredRecords := int(hours * 2)
//	dataSetData := make([]data.Datum, requiredRecords)
//	deviceId := "SummaryTestDevice"
//	uploadId := test.RandomSetID()
//
//	glucoseValue := pointer.FromAny(inTargetBloodGlucose)
//
//	// generate X hours of data
//	for count := 0; count < requiredRecords; count += 1 {
//		datumTime := startTime.Add(time.Duration(count-requiredRecords) * time.Minute * 30)
//
//		datum := NewGlucose(&typ, &units, &datumTime, &deviceId, &uploadId)
//		datum.Value = glucoseValue
//
//		if realtime {
//			datum.CreatedTime = pointer.FromAny(datumTime.Add(5 * time.Minute))
//			datum.ModifiedTime = pointer.FromAny(datumTime.Add(10 * time.Minute))
//		}
//
//		dataSetData[count] = datum
//	}
//
//	return dataSetData
//}
//
//var _ = Describe("Continuous Summary", func() {
//	var userId string
//	var datumTime time.Time
//	var err error
//	var dataSetContinuousData []data.Datum
//
//	BeforeEach(func() {
//		userId = userTest.RandomID()
//		datumTime = time.Now().UTC().Truncate(24 * time.Hour)
//	})
//
//	Context("Create Summary", func() {
//		It("Correctly initializes a summary", func() {
//			summary := types.Create[*types.ContinuousStats](userId)
//			Expect(summary).ToNot(BeNil())
//			Expect(summary.Type).To(Equal("continuous"))
//
//			Expect(summary.UserID).To(Equal(userId))
//			Expect(summary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
//		})
//	})
//
//	Context("Summary calculations requiring datasets", func() {
//		var userContinuousSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//
//		Context("AddData Bucket Testing", func() {
//			It("Returns correct hour count when given 2 weeks", func() {
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 336, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(336))
//			})
//
//			It("Returns correct hour count when given 1 week", func() {
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 168, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(168))
//			})
//
//			It("Returns correct hour count when given 3 weeks", func() {
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 504, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(504))
//			})
//
//			It("Returns correct records when given >60d of data", func() {
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 5, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userContinuousSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(2))
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime.Add(1*time.Hour), 1, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userContinuousSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(2))
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime.Add(24*60*time.Hour), 1, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userContinuousSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(2))
//
//				for i := 0; i < len(userContinuousSummary.Stats.Buckets); i++ {
//					Expect(userContinuousSummary.Stats.Buckets[i]).ToNot(BeNil())
//				}
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(1440))
//			})
//
//			It("Returns correct records when given data a full 60d ahead of previous data", func() {
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 1, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//				Expect(err).ToNot(HaveOccurred())
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime.Add(24*62*time.Hour), 1, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//				Expect(err).ToNot(HaveOccurred())
//
//				for i := 0; i < len(userContinuousSummary.Stats.Buckets); i++ {
//					Expect(userContinuousSummary.Stats.Buckets[i]).ToNot(BeNil())
//				}
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(1))
//			})
//
//			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
//				var lastRecordTime time.Time
//				var hourlyStatsLen int
//				var newHourlyStatsLen int
//				secondDatumTime := datumTime.AddDate(0, 0, 15)
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 168, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(168))
//
//				By("check total glucose and dates for first batch")
//				hourlyStatsLen = len(userContinuousSummary.Stats.Buckets)
//				for i := hourlyStatsLen - 1; i >= 0; i-- {
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(2))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 30*time.Minute)
//					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, secondDatumTime, 168, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(528)) // 22 days
//
//				By("check total glucose and dates for second batch")
//				newHourlyStatsLen = len(userContinuousSummary.Stats.Buckets)
//				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetContinuousData)/2 // 2 per day, need length without the gap
//				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(2))
//
//					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 30*time.Minute)
//					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("check total glucose and dates for gap")
//				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
//				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(0))
//				}
//			})
//
//			It("Returns correct stats when given multiple batches in a day", func() {
//				var incrementalDatumTime time.Time
//				var lastRecordTime time.Time
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//
//				dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, datumTime, 144, true)
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(144))
//
//				for i := 1; i <= 24; i++ {
//					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
//					dataSetContinuousData = NewDataSetDataRealtime(continuous.Type, incrementalDatumTime, 1, true)
//
//					err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousData)
//
//					Expect(err).ToNot(HaveOccurred())
//					Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(144 + i))
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(2))
//				}
//
//				for i := 144; i < len(userContinuousSummary.Stats.Buckets); i++ {
//					f := fmt.Sprintf("hour %d", i)
//					By(f)
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(2))
//
//					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*30)
//					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//			})
//
//			It("Returns correct hourly stats for days uploaded in reverse", func() {
//				var lastRecordTime time.Time
//				userContinuousSummary = types.Create[*types.ContinuousStats](userId)
//
//				// Datasets use +1 and +2 offset to allow for checking via iteration
//				dataSetContinuousDataOne := NewDataSetDataRealtime(continuous.Type, datumTime.AddDate(0, 0, -2), 24, true)
//				dataSetContinuousDataTwo := NewDataSetDataRealtime(continuous.Type, datumTime.AddDate(0, 0, -1), 24, true)
//				dataSetContinuousDataThree := NewDataSetDataRealtime(continuous.Type, datumTime, 24, true)
//
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousDataThree)
//				Expect(err).ToNot(HaveOccurred())
//
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousDataTwo)
//				Expect(err).ToNot(HaveOccurred())
//
//				err = types.AddData(&userContinuousSummary.Stats.Buckets, dataSetContinuousDataOne)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userContinuousSummary.Stats.Buckets)).To(Equal(72))
//
//				for i := len(userContinuousSummary.Stats.Buckets) - 1; i >= 0; i-- {
//					By(fmt.Sprintf("hour %d", i+1))
//					Expect(userContinuousSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(2))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userContinuousSummary.Stats.Buckets)-i-1) - 30*time.Minute)
//					Expect(userContinuousSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//			})
//		})
//
//		Context("GetPatientsWithRealtimeData", func() {
//
//			It("with some realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 15)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(count).To(Equal(15))
//			})
//
//			It("with no realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 0)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(count).To(Equal(0))
//			})
//
//			It("with 60d of realtime data", func() {
//				endTime := time.Now().UTC().Truncate(time.Hour * 24)
//				startTime := endTime.AddDate(0, 0, -30)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId, startTime, endTime, 60)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(30))
//			})
//
//			It("with a week of realtime data, with a non-utc, non-dst timezone", func() {
//				loc1 := time.FixedZone("suffering", 12*3600)
//				loc2 := time.FixedZone("pain", 12*3600)
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//
//			It("with a week of realtime data, with a non-utc, dst timezone", func() {
//				loc1 := time.FixedZone("suffering", 12*3600)
//				loc2 := time.FixedZone("pain", 13*3600)
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//
//			It("with a week of realtime data, with a non-utc, dst timezone backwards", func() {
//				loc1 := time.FixedZone("pain", 13*3600)
//				loc2 := time.FixedZone("sadness", 12*3600)
//
//				lastWeekInNZ := time.Now().In(loc2)
//
//				endTime := time.Date(lastWeekInNZ.Year(), lastWeekInNZ.Month(), lastWeekInNZ.Day(), 23, 59, 59, 0, loc2)
//				startTime := endTime.AddDate(0, 0, -2)
//				startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc1)
//
//				userContinuousSummary = summaryTest.NewRealtimeSummary(userId,
//					startTime.AddDate(0, 0, -2),
//					endTime.AddDate(0, 0, 2),
//					7)
//
//				count := userContinuousSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(count).To(Equal(3))
//			})
//		})
//	})
//})
