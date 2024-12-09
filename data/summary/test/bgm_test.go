package test_test

//import (
//	"context"
//	"fmt"
//	"math/rand"
//	"strconv"
//	"time"
//
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	"go.mongodb.org/mongo-driver/mongo"
//
//	"github.com/tidepool-org/platform/data"
//	"github.com/tidepool-org/platform/data/summary"
//	"github.com/tidepool-org/platform/data/summary/fetcher"
//	"github.com/tidepool-org/platform/data/summary/types"
//	"github.com/tidepool-org/platform/data/test"
//	"github.com/tidepool-org/platform/log"
//	logTest "github.com/tidepool-org/platform/log/test"
//	"github.com/tidepool-org/platform/pointer"
//	userTest "github.com/tidepool-org/platform/user/test"
//)
//
//func BGMCursorFetcher(c *mongo.Cursor) fetcher.DeviceDataCursor {
//	return fetcher.NewDefaultCursor(c, summary.CreateGlucoseDatum)
//}
//
//func NewDataSetBGMDataAvg(deviceId string, startTime time.Time, hours float64, reqAvg float64) []data.Datum {
//	requiredRecords := int(hours * 6)
//	typ := "smbg"
//	dataSetData := make([]data.Datum, requiredRecords)
//	uploadId := test.RandomSetID()
//
//	// generate X hours of data
//	for count := 0; count < requiredRecords; count += 2 {
//		randValue := 1 + rand.Float64()*(reqAvg-1)
//		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}
//
//		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
//		for i, glucoseValue := range glucoseValues {
//			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 10)
//
//			datum := NewGlucose(&typ, &units, &datumTime, &deviceId, &uploadId)
//			datum.Value = pointer.FromAny(glucoseValue)
//
//			dataSetData[requiredRecords-count-i-1] = datum
//		}
//	}
//
//	return dataSetData
//}
//
//func NewDataSetBGMDataRanges(deviceId string, startTime time.Time, hours float64, ranges DataRanges) []data.Datum {
//	requiredRecords := int(hours * 6)
//	typ := "smbg"
//	dataSetData := make([]data.Datum, requiredRecords)
//	uploadId := test.RandomSetID()
//
//	glucoseBrackets := [6][2]float64{
//		{ranges.Min, ranges.VeryLow - ranges.Padding},
//		{ranges.VeryLow, ranges.Low - ranges.Padding},
//		{ranges.Low, ranges.High - ranges.Padding},
//		{ranges.High, ranges.VeryHigh - ranges.Padding},
//		{ranges.VeryHigh, ranges.ExtremeHigh - ranges.Padding},
//		{ranges.ExtremeHigh, ranges.Max},
//	}
//
//	// generate requiredRecords of data
//	for count := 0; count < requiredRecords; count += 6 {
//		for i, bracket := range glucoseBrackets {
//			datumTime := startTime.Add(-time.Duration(count+i+1) * time.Minute * 10)
//
//			datum := NewGlucose(&typ, &units, &datumTime, &deviceId, &uploadId)
//			datum.Value = pointer.FromAny(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())
//
//			dataSetData[requiredRecords-count-i-1] = datum
//		}
//	}
//
//	return dataSetData
//}
//
//var _ = Describe("BGM Summary", func() {
//	var userId string
//	var datumTime time.Time
//	var deviceId string
//	var logger log.Logger
//	var ctx context.Context
//	var err error
//	var dataSetBGMData []data.Datum
//
//	BeforeEach(func() {
//		logger = logTest.NewLogger()
//		ctx = log.NewContextWithLogger(context.Background(), logger)
//		userId = userTest.RandomID()
//		deviceId = "SummaryTestDevice"
//		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
//	})
//
//	Context("CreateBGMSummary", func() {
//		It("Correctly initializes a bgm summary", func() {
//			summary := types.Create[*types.BGMStats](userId)
//			Expect(summary).To(Not(BeNil()))
//			Expect(summary.Type).To(Equal("bgm"))
//
//			Expect(summary.UserID).To(Equal(userId))
//			Expect(summary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
//		})
//	})
//
//	Context("Summary calculations requiring datasets", func() {
//		var userBGMSummary *types.Summary[*types.BGMStats, types.BGMStats]
//		var periodKeys = []string{"1d", "7d", "14d", "30d"}
//		var periodInts = []int{1, 7, 14, 30}
//
//		Context("AddData Bucket Testing", func() {
//			It("Returns correct hour count when given 2 weeks", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 336, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(336))
//			})
//
//			It("Returns correct hour count when given 1 week", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))
//			})
//
//			It("Returns correct hour count when given 3 weeks", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 504, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(504))
//			})
//
//			It("Returns correct record count when given overlapping records", func() {
//				// NOTE CGM would filter these, we are testing that they don't get filtered here
//				var doubledBGMData = make([]data.Datum, 288*2)
//
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
//				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)
//
//				// interlace the lists
//				for i := 0; i < len(dataSetBGMData); i += 1 {
//					doubledBGMData[i*2] = dataSetBGMData[i]
//					doubledBGMData[i*2+1] = dataSetBGMDataTwo[i]
//				}
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(6))
//			})
//
//			It("Returns correct record count when given overlapping records across multiple calculations", func() {
//				// NOTE CGM would filter these, we are testing that they don't get filtered here
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(12))
//			})
//
//			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
//				var lastRecordTime time.Time
//				var hourlyStatsLen int
//				var newHourlyStatsLen int
//				secondDatumTime := datumTime.AddDate(0, 0, 15)
//				secondRequestedAvgGlucose := lowBloodGlucose
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))
//
//				By("check total glucose and dates for first batch")
//				hourlyStatsLen = len(userBGMSummary.Stats.Buckets)
//				for i := hourlyStatsLen - 1; i >= 0; i-- {
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", inTargetBloodGlucose*6, 0.001))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 10*time.Minute)
//					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, secondDatumTime, 168, secondRequestedAvgGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(528)) // 22 days
//
//				By("check total glucose and dates for second batch")
//				newHourlyStatsLen = len(userBGMSummary.Stats.Buckets)
//				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetBGMData)/12 // 12 per day, need length without the gap
//				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", secondRequestedAvgGlucose*6))
//
//					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 10*time.Minute)
//					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("check total glucose and dates for gap")
//				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
//				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64(0)))
//				}
//			})
//
//			It("Returns correct stats when given multiple batches in a day", func() {
//				var incrementalDatumTime time.Time
//				var lastRecordTime time.Time
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 144, inTargetBloodGlucose)
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(144))
//
//				for i := 1; i <= 24; i++ {
//					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
//					dataSetBGMData = NewDataSetBGMDataAvg(deviceId, incrementalDatumTime, 1, float64(i))
//
//					err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//					Expect(err).ToNot(HaveOccurred())
//					Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(144 + i))
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))
//				}
//
//				for i := 144; i < len(userBGMSummary.Stats.Buckets); i++ {
//					f := fmt.Sprintf("hour %d", i)
//					By(f)
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))
//
//					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*10)
//					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64((i - 143) * 6)))
//
//					averageGlucoseMmol := userBGMSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userBGMSummary.Stats.Buckets[i].Data.TotalRecords)
//					Expect(averageGlucoseMmol).To(BeNumerically("~", i-143, 0.005))
//				}
//			})
//
//			It("Returns correct daily stats for days with different averages", func() {
//				var expectedTotalGlucose float64
//				var lastRecordTime time.Time
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				// Datasets use +1 and +2 offset to allow for checking via iteration
//				dataSetBGMDataOne := NewDataSetBGMDataAvg(deviceId, datumTime.AddDate(0, 0, -2), 24, inTargetBloodGlucose)
//				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceId, datumTime.AddDate(0, 0, -1), 24, inTargetBloodGlucose+1)
//				dataSetBGMDataThree := NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose+2)
//				dataSetBGMData = append(dataSetBGMDataOne, dataSetBGMDataTwo...)
//				dataSetBGMData = append(dataSetBGMData, dataSetBGMDataThree...)
//
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(72))
//
//				for i := len(userBGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userBGMSummary.Stats.Buckets)-i-1) - 10*time.Minute)
//					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//
//					expectedTotalGlucose = (inTargetBloodGlucose + float64(i/24)) * 6
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", expectedTotalGlucose, 0.001))
//				}
//			})
//
//			It("Returns correct hourly stats for hours with different Time in Range", func() {
//				var lastRecordTime time.Time
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
//				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
//				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
//				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
//				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)
//				extremeHighRange := NewDataRangesSingle(extremeHighBloodGlucose + 0.5)
//
//				dataSetBGMDataOne := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-5*time.Hour), 1, veryLowRange)
//				dataSetBGMDataTwo := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-4*time.Hour), 1, lowRange)
//				dataSetBGMDataThree := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-3*time.Hour), 1, inRange)
//				dataSetBGMDataFour := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-2*time.Hour), 1, highRange)
//				dataSetBGMDataFive := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-1*time.Hour), 1, veryHighRange)
//				dataSetBGMDataSix := NewDataSetBGMDataRanges(deviceId, datumTime, 1, extremeHighRange)
//
//				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataOne)
//				Expect(err).ToNot(HaveOccurred())
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataTwo)
//				Expect(err).ToNot(HaveOccurred())
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataThree)
//				Expect(err).ToNot(HaveOccurred())
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataFour)
//				Expect(err).ToNot(HaveOccurred())
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataFive)
//				Expect(err).ToNot(HaveOccurred())
//				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataSix)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(6))
//
//				By("check record counters for insurance")
//				for i := len(userBGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userBGMSummary.Stats.Buckets)-i-1) - time.Minute*10)
//					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("very low records")
//				Expect(userBGMSummary.Stats.Buckets[0].Data.VeryLowRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.TargetRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.HighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.VeryHighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[0].Data.ExtremeHighRecords).To(Equal(0))
//
//				By("low records")
//				Expect(userBGMSummary.Stats.Buckets[1].Data.VeryLowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[1].Data.LowRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[1].Data.TargetRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[1].Data.HighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[1].Data.VeryHighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[1].Data.ExtremeHighRecords).To(Equal(0))
//
//				By("in-range records")
//				Expect(userBGMSummary.Stats.Buckets[2].Data.VeryLowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[2].Data.LowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[2].Data.TargetRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[2].Data.HighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[2].Data.VeryHighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[2].Data.ExtremeHighRecords).To(Equal(0))
//
//				By("high records")
//				Expect(userBGMSummary.Stats.Buckets[3].Data.VeryLowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[3].Data.LowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[3].Data.TargetRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[3].Data.HighRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[3].Data.VeryHighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[3].Data.ExtremeHighRecords).To(Equal(0))
//
//				By("very high records")
//				Expect(userBGMSummary.Stats.Buckets[4].Data.VeryLowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[4].Data.LowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[4].Data.TargetRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[4].Data.HighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[4].Data.VeryHighRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[4].Data.ExtremeHighRecords).To(Equal(0))
//
//				By("extreme high records")
//				Expect(userBGMSummary.Stats.Buckets[5].Data.VeryLowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[5].Data.LowRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[5].Data.TargetRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[5].Data.HighRecords).To(Equal(0))
//				Expect(userBGMSummary.Stats.Buckets[5].Data.VeryHighRecords).To(Equal(6))
//				Expect(userBGMSummary.Stats.Buckets[5].Data.ExtremeHighRecords).To(Equal(6))
//			})
//		})
//
//		Context("CalculateDelta", func() {
//			It("Returns correct deltas for periods", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				for i, period := range periodKeys {
//					vFloat := float64(i) * 7.5
//					vFloatTwo := vFloat * 2
//					vInt := i * 7
//					vIntTwo := vInt * 2
//
//					userBGMSummary.Stats.Periods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    pointer.FromAny(vFloat),
//						TotalRecords:          pointer.FromAny(vInt),
//						AverageDailyRecords:   pointer.FromAny(vFloat),
//						TimeInTargetPercent:   pointer.FromAny(vFloat),
//						TimeInTargetRecords:   pointer.FromAny(vInt),
//						TimeInLowPercent:      pointer.FromAny(vFloat),
//						TimeInLowRecords:      pointer.FromAny(vInt),
//						TimeInVeryLowPercent:  pointer.FromAny(vFloat),
//						TimeInVeryLowRecords:  pointer.FromAny(vInt),
//						TimeInHighPercent:     pointer.FromAny(vFloat),
//						TimeInHighRecords:     pointer.FromAny(vInt),
//						TimeInVeryHighPercent: pointer.FromAny(vFloat),
//						TimeInVeryHighRecords: pointer.FromAny(vInt),
//					}
//
//					userBGMSummary.Stats.OffsetPeriods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    pointer.FromAny(vFloatTwo),
//						TotalRecords:          pointer.FromAny(vIntTwo),
//						AverageDailyRecords:   pointer.FromAny(vFloatTwo),
//						TimeInTargetPercent:   pointer.FromAny(vFloatTwo),
//						TimeInTargetRecords:   pointer.FromAny(vIntTwo),
//						TimeInLowPercent:      pointer.FromAny(vFloatTwo),
//						TimeInLowRecords:      pointer.FromAny(vIntTwo),
//						TimeInVeryLowPercent:  pointer.FromAny(vFloatTwo),
//						TimeInVeryLowRecords:  pointer.FromAny(vIntTwo),
//						TimeInHighPercent:     pointer.FromAny(vFloatTwo),
//						TimeInHighRecords:     pointer.FromAny(vIntTwo),
//						TimeInVeryHighPercent: pointer.FromAny(vFloatTwo),
//						TimeInVeryHighRecords: pointer.FromAny(vIntTwo),
//					}
//				}
//
//				userBGMSummary.Stats.CalculateDelta()
//
//				for i, period := range periodKeys {
//					floatDiff := float64(i)*7.5 - float64(i)*7.5*2
//					intDiff := i*7 - i*7*2
//
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(Equal(-floatDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TotalRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(Equal(-intDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(Equal(-floatDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(Equal(-floatDiff))
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(Equal(-intDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(Equal(-floatDiff))
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(Equal(-intDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(Equal(-floatDiff))
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(Equal(-intDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(Equal(-floatDiff))
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(Equal(-intDiff))
//
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(Equal(floatDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(Equal(-floatDiff))
//					Expect(*userBGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(Equal(intDiff))
//					Expect(*userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(Equal(-intDiff))
//				}
//			})
//
//			It("Returns correct nil deltas with nil latest stats", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				for _, period := range periodKeys {
//					userBGMSummary.Stats.Periods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    nil,
//						TotalRecords:          nil,
//						AverageDailyRecords:   nil,
//						TimeInTargetPercent:   nil,
//						TimeInTargetRecords:   nil,
//						TimeInLowPercent:      nil,
//						TimeInLowRecords:      nil,
//						TimeInVeryLowPercent:  nil,
//						TimeInVeryLowRecords:  nil,
//						TimeInHighPercent:     nil,
//						TimeInHighRecords:     nil,
//						TimeInVeryHighPercent: nil,
//						TimeInVeryHighRecords: nil,
//					}
//
//					userBGMSummary.Stats.OffsetPeriods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    pointer.FromAny(1.0),
//						TotalRecords:          pointer.FromAny(1),
//						AverageDailyRecords:   pointer.FromAny(1.0),
//						TimeInTargetPercent:   pointer.FromAny(1.0),
//						TimeInTargetRecords:   pointer.FromAny(1),
//						TimeInLowPercent:      pointer.FromAny(1.0),
//						TimeInLowRecords:      pointer.FromAny(1),
//						TimeInVeryLowPercent:  pointer.FromAny(1.0),
//						TimeInVeryLowRecords:  pointer.FromAny(1),
//						TimeInHighPercent:     pointer.FromAny(1.0),
//						TimeInHighRecords:     pointer.FromAny(1),
//						TimeInVeryHighPercent: pointer.FromAny(1.0),
//						TimeInVeryHighRecords: pointer.FromAny(1),
//					}
//				}
//
//				userBGMSummary.Stats.CalculateDelta()
//
//				for _, period := range periodKeys {
//					Expect(userBGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TotalRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(BeNil())
//				}
//			})
//
//			It("Returns correct nil deltas with nil offset stats", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				for _, period := range periodKeys {
//					userBGMSummary.Stats.Periods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    pointer.FromAny(1.0),
//						TotalRecords:          pointer.FromAny(1),
//						AverageDailyRecords:   pointer.FromAny(1.0),
//						TimeInTargetPercent:   pointer.FromAny(1.0),
//						TimeInTargetRecords:   pointer.FromAny(1),
//						TimeInLowPercent:      pointer.FromAny(1.0),
//						TimeInLowRecords:      pointer.FromAny(1),
//						TimeInVeryLowPercent:  pointer.FromAny(1.0),
//						TimeInVeryLowRecords:  pointer.FromAny(1),
//						TimeInHighPercent:     pointer.FromAny(1.0),
//						TimeInHighRecords:     pointer.FromAny(1),
//						TimeInVeryHighPercent: pointer.FromAny(1.0),
//						TimeInVeryHighRecords: pointer.FromAny(1),
//					}
//
//					userBGMSummary.Stats.OffsetPeriods[period] = &types.BGMPeriod{
//						AverageGlucoseMmol:    nil,
//						TotalRecords:          nil,
//						AverageDailyRecords:   nil,
//						TimeInTargetPercent:   nil,
//						TimeInTargetRecords:   nil,
//						TimeInLowPercent:      nil,
//						TimeInLowRecords:      nil,
//						TimeInVeryLowPercent:  nil,
//						TimeInVeryLowRecords:  nil,
//						TimeInHighPercent:     nil,
//						TimeInHighRecords:     nil,
//						TimeInVeryHighPercent: nil,
//						TimeInVeryHighRecords: nil,
//					}
//				}
//
//				userBGMSummary.Stats.CalculateDelta()
//
//				for _, period := range periodKeys {
//					Expect(userBGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TotalRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(BeNil())
//
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(BeNil())
//					Expect(userBGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(BeNil())
//				}
//			})
//		})
//
//		Context("CalculateSummary/Update", func() {
//			var newDatumTime time.Time
//			var dataSetBGMDataCursor *mongo.Cursor
//
//			It("Returns correct time in range for stats", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				ranges := NewDataRanges()
//				dataSetBGMData = NewDataSetBGMDataRanges(deviceId, datumTime, 720, ranges)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))
//
//				stopPoints := []int{1, 7, 14, 30}
//				for _, v := range stopPoints {
//					periodKey := strconv.Itoa(v) + "d"
//
//					f := fmt.Sprintf("period %s", periodKey)
//					By(f)
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInTargetRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInTargetRecords).To(Equal(24 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInTargetPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInTargetPercent).To(Equal(1.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryLowRecords).To(Equal(24 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryLowPercent).To(Equal(1.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInAnyLowRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInAnyLowRecords).To(Equal(24 * 2 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInAnyLowPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInAnyLowPercent).To(Equal(2.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInLowRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInLowRecords).To(Equal(24 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInLowPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInLowPercent).To(Equal(1.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInHighRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInHighRecords).To(Equal(24 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInHighPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInHighPercent).To(Equal(1.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryHighPercent).To(Equal(2.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInExtremeHighRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInExtremeHighRecords).To(Equal(24 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInExtremeHighPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInExtremeHighPercent).To(Equal(1.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInAnyHighRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInAnyHighRecords).To(Equal(36 * 2 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInAnyHighPercent).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInAnyHighPercent).To(Equal(3.0 / 6.0))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasTotalRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].TotalRecords).To(Equal(144 * v))
//
//					Expect(userBGMSummary.Stats.Periods[periodKey].HasAverageDailyRecords).To(BeTrue())
//					Expect(*userBGMSummary.Stats.Periods[periodKey].AverageDailyRecords).To(BeNumerically("==", 144))
//				}
//			})
//
//			It("Returns correct average glucose for stats", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(Equal(inTargetBloodGlucose))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly calculated summary with no rolling", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", inTargetBloodGlucose, 0.001))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly calculated summary with rolling low to high record counts", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				newDatumTime = datumTime.AddDate(0, 0, 30)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 1, lowBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//
//				// start the actual test
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 720, highBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(721))
//
//				userBGMSummary.Stats.CalculateSummary()
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(721))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly calculated summary with rolling high to low record counts", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				newDatumTime = datumTime.Add(time.Duration(23) * time.Hour)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, lowBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.005))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//
//				// start the actual test
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 23, highBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(743))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(743))
//
//				for i, period := range periodKeys {
//					expectedAverage := ExpectedAverage(periodInts[i]*24, 23, highBloodGlucose, lowBloodGlucose)
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", expectedAverage, 0.005))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly non-rolling summary with two 30 day windows", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				newDatumTime = datumTime.AddDate(0, 0, 61)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, lowBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(24))
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//
//				// start the actual test
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 168, highBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1440))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1440)) // 60 days
//
//				for _, period := range periodKeys {
//					Expect(*userBGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))
//					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly moving offset periods", func() {
//				// Here we generate 5 1d datasets, and add them in a specific order:
//				// -59d -- target glucose
//				// -27d -- veryHigh glucose
//				// -13d -- veryLow glucose
//				//  -1d -- high glucose
//				//   0d -- low glucose
//
//				// This should result in:
//				//  1d regular  -- low, 144 readings (from 0d)
//				//  1d offset   -- high, 144 readings (from 1d)
//				//  7d regular  -- (high+low)/2, 144*2 (288) readings (from 0d + 1d)
//				//  7d offset   -- veryLow, 144 readings (from 14d)
//				// 14d regular -- (high+low+veryLow)/3, 144*3 (432) readings (from 1d + 2d + 14d)
//				// 14d offset  -- veryHigh, 144 readings (from 28d)
//				// 30d regular -- (high+veryHigh+low+veryLow)/4, 144*4 (576) readings (from 1d + 2d + 14d + 28d)
//				// 30d offset  -- target, 144 readings (from 60d)
//
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//
//				newDatumTimeOne := datumTime.AddDate(0, 0, -59)
//				newDatumTimeTwo := datumTime.AddDate(0, 0, -27)
//				newDatumTimeThree := datumTime.AddDate(0, 0, -13)
//				newDatumTimeFour := datumTime.AddDate(0, 0, -1)
//				newDatumTimeFive := datumTime
//
//				dataSetBGMDataOne := NewDataSetBGMDataAvg(deviceId, newDatumTimeOne, 24, inTargetBloodGlucose)
//				dataSetBGMDataOneCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMDataOne), nil, nil)
//				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceId, newDatumTimeTwo, 24, veryHighBloodGlucose)
//				dataSetBGMDataTwoCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMDataTwo), nil, nil)
//				dataSetBGMDataThree := NewDataSetBGMDataAvg(deviceId, newDatumTimeThree, 24, veryLowBloodGlucose)
//				dataSetBGMDataThreeCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMDataThree), nil, nil)
//
//				dataSetBGMDataFour := NewDataSetBGMDataAvg(deviceId, newDatumTimeFour, 24, highBloodGlucose)
//				dataSetBGMDataFourCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMDataFour), nil, nil)
//
//				dataSetBGMDataFive := NewDataSetBGMDataAvg(deviceId, newDatumTimeFive, 24, lowBloodGlucose)
//				dataSetBGMDataFiveCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMDataFive), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataOneCursor))
//				Expect(err).ToNot(HaveOccurred())
//
//				// first day, should have 24 buckets
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(24))
//				Expect(*userBGMSummary.Stats.Periods["1d"].TotalRecords).To(Equal(24 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["1d"].TotalRecords).To(Equal(0))
//				Expect(*userBGMSummary.Stats.Periods["7d"].TotalRecords).To(Equal(24 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["7d"].TotalRecords).To(Equal(0))
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataTwoCursor))
//				Expect(err).ToNot(HaveOccurred())
//
//				// 33 days elapsed, should have 33*24 (792) buckets
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(792))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(792))
//				Expect(*userBGMSummary.Stats.Periods["14d"].TotalRecords).To(Equal(24 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["14d"].TotalRecords).To(Equal(0))
//				Expect(*userBGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 6))
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataThreeCursor))
//				Expect(err).ToNot(HaveOccurred())
//
//				// 47 days elapsed, should have 47*24 (1128) buckets
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1128))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1128))
//				Expect(*userBGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 2 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 6))
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataFourCursor))
//				Expect(err).ToNot(HaveOccurred())
//
//				// 59 days elapsed, should have 59*24 (1416) buckets
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1416))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1416))
//				Expect(*userBGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 3 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 1 * 6))
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataFiveCursor))
//				Expect(err).ToNot(HaveOccurred())
//
//				// 60 days elapsed, should have 60*24 (1440) buckets
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1440))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1440))
//				Expect(*userBGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 4 * 6))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 1 * 6))
//
//				// check that the data matches the expectation described at the top of the test
//				Expect(*userBGMSummary.Stats.Periods["1d"].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))
//				Expect(*userBGMSummary.Stats.Periods["1d"].TotalRecords).To(Equal(144))
//
//				Expect(*userBGMSummary.Stats.OffsetPeriods["1d"].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["1d"].TotalRecords).To(Equal(144))
//
//				Expect(*userBGMSummary.Stats.Periods["7d"].AverageGlucoseMmol).To(BeNumerically("~", (highBloodGlucose+lowBloodGlucose)/2, 0.001))
//				Expect(*userBGMSummary.Stats.Periods["7d"].TotalRecords).To(Equal(144 * 2))
//
//				Expect(*userBGMSummary.Stats.OffsetPeriods["7d"].AverageGlucoseMmol).To(BeNumerically("~", veryLowBloodGlucose, 0.001))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["7d"].TotalRecords).To(Equal(144))
//
//				Expect(*userBGMSummary.Stats.Periods["14d"].AverageGlucoseMmol).To(BeNumerically("~", (highBloodGlucose+lowBloodGlucose+veryLowBloodGlucose)/3, 0.001))
//				Expect(*userBGMSummary.Stats.Periods["14d"].TotalRecords).To(Equal(144 * 3))
//
//				Expect(*userBGMSummary.Stats.OffsetPeriods["14d"].AverageGlucoseMmol).To(BeNumerically("~", veryHighBloodGlucose, 0.001))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["14d"].TotalRecords).To(Equal(144))
//
//				Expect(*userBGMSummary.Stats.Periods["30d"].AverageGlucoseMmol).To(BeNumerically("~", (veryHighBloodGlucose+highBloodGlucose+lowBloodGlucose+veryLowBloodGlucose)/4, 0.001))
//				Expect(*userBGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(144 * 4))
//
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].AverageGlucoseMmol).To(BeNumerically("~", inTargetBloodGlucose, 0.001))
//				Expect(*userBGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(144))
//			})
//		})
//
//		Context("ClearInvalidatedBuckets", func() {
//			var dataSetBGMDataCursor *mongo.Cursor
//
//			It("trims the correct buckets", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 10, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(10))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(-5 * time.Hour))
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(5))
//
//				// we didn't overshoot and nil something we shouldn't have
//				Expect(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
//
//				Expect(firstData).To(Equal(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1].LastRecordTime))
//			})
//
//			It("trims the all buckets with data beyond the beginning of the buckets", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 10, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(10))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(-15 * time.Hour))
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(0))
//
//				Expect(firstData.IsZero()).To(BeTrue())
//			})
//
//			It("doesnt trim if only modified in the future", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 10, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(10))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(datumTime.Add(time.Hour))
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//
//				// we didn't overshoot and nil something we shouldn't have
//				Expect(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
//
//				Expect(firstData).To(Equal(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1].LastRecordTime))
//			})
//
//			It("doesnt trim if only modified on the same hour, but after the bucket time", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				midDatumTime := datumTime.Add(30 * time.Minute)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, midDatumTime, 9, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(10))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(midDatumTime.Add(10 * time.Minute))
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//
//				// we didn't overshoot and nil something we shouldn't have
//				Expect(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
//
//				Expect(firstData).To(Equal(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1].LastRecordTime))
//			})
//
//			It("trims if modified on the same hour, and before the bucket time", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				midDatumTime := datumTime.Add(30 * time.Minute)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, midDatumTime, 9, inTargetBloodGlucose)
//				dataSetBGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetBGMData), nil, nil)
//
//				err = userBGMSummary.Stats.Update(ctx, BGMCursorFetcher(dataSetBGMDataCursor))
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(10))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(10))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(midDatumTime.Add(-10 * time.Minute))
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(9))
//
//				// we didn't overshoot and nil something we shouldn't have
//				Expect(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
//
//				Expect(firstData).To(Equal(userBGMSummary.Stats.Buckets[len(userBGMSummary.Stats.Buckets)-1].LastRecordTime))
//			})
//
//			It("successfully does nothing if there are no buckets", func() {
//				userBGMSummary = types.Create[*types.BGMStats](userId)
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(0))
//				Expect(userBGMSummary.Stats.TotalHours).To(Equal(0))
//
//				firstData := userBGMSummary.Stats.ClearInvalidatedBuckets(datumTime)
//
//				// we have the right length
//				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(0))
//
//				Expect(firstData.IsZero()).To(BeTrue())
//			})
//		})
//	})
//})
