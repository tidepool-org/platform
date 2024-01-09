package types_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

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

func NewDataSetCGMDataAvg(deviceId string, startTime time.Time, hours float64, reqAvg float64) []*glucose.Glucose {
	requiredRecords := int(hours * 12)
	typ := pointer.FromString("cbg")

	var dataSetData = make([]*glucose.Glucose, requiredRecords)

	// generate X hours of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceId)
			datum.Value = pointer.FromFloat64(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

// creates a dataset with random values evenly divided between ranges
// NOTE: only generates 98.9% CGMUse, due to needing to be divisible by 5
func NewDataSetCGMDataRanges(deviceId string, startTime time.Time, hours float64, ranges DataRanges) []*glucose.Glucose {
	requiredRecords := int(hours * 10)
	typ := pointer.FromString("cbg")
	var gapCompensation time.Duration

	var dataSetData = make([]*glucose.Glucose, requiredRecords)

	glucoseBrackets := [5][2]float64{
		{ranges.Min, ranges.VeryLow - ranges.Padding},
		{ranges.VeryLow, ranges.Low - ranges.Padding},
		{ranges.Low, ranges.High - ranges.Padding},
		{ranges.High, ranges.VeryHigh - ranges.Padding},
		{ranges.VeryHigh, ranges.Max},
	}

	// generate requiredRecords of data
	for count := 0; count < requiredRecords; count += 5 {
		gapCompensation = 10 * time.Minute * time.Duration(int(float64(count+1)/10))
		for i, bracket := range glucoseBrackets {
			datumTime := startTime.Add(time.Duration(-(count+i+1))*time.Minute*5 - gapCompensation)

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceId)
			datum.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

var _ = Describe("CGM Summary", func() {
	var userId string
	var datumTime time.Time
	var deviceId string
	var logger log.Logger
	var ctx context.Context
	var err error
	var dataSetCGMData []*glucose.Glucose

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		userId = userTest.RandomID()
		deviceId = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("CreateCGMSummary", func() {
		It("Correctly initializes a cgm summary", func() {
			summary := types.Create[*types.CGMStats](userId)
			Expect(summary).To(Not(BeNil()))
			Expect(summary.Type).To(Equal("cgm"))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userCGMSummary *types.Summary[types.CGMStats, *types.CGMStats]
		var periodKeys = []string{"1d", "7d", "14d", "30d"}
		var periodInts = []int{1, 7, 14, 30}

		Context("AddData Bucket Testing", func() {
			It("Returns correct hour count when given 2 weeks", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 336, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct hour count when given 1 week", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct hour count when given 3 weeks", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 504, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct record count when given overlapping records", func() {
				var doubledCGMData = make([]*glucose.Glucose, 288*2)

				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)

				// interlace the lists
				for i := 0; i < len(dataSetCGMData); i += 1 {
					doubledCGMData[i*2] = dataSetCGMData[i]
					doubledCGMData[i*2+1] = dataSetCGMDataTwo[i]
				}
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userCGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(12))
			})

			It("Returns correct record count when given overlapping records across multiple calculations", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userCGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(12))
			})

			It("Returns correct records when given >60d of data", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime, 5, NewDataRangesSingle(lowBloodGlucose-0.5))
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(10))

				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime.Add(1*time.Hour), 1, NewDataRangesSingle(highBloodGlucose+0.5))
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(10))

				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime.Add(24*60*time.Hour), 1, NewDataRangesSingle(inTargetBloodGlucose-0.5))
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Stats.Buckets[0].Data.HighRecords).To(Equal(10))

				for i := 0; i < len(userCGMSummary.Stats.Buckets); i++ {
					Expect(userCGMSummary.Stats.Buckets[i]).ToNot(BeNil())
				}
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1440))
			})

			It("Returns correct records when given data a full 60d ahead of previous data", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime, 1, NewDataRangesSingle(lowBloodGlucose-0.5))
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime.Add(24*62*time.Hour), 1, NewDataRangesSingle(inTargetBloodGlucose-0.5))
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < len(userCGMSummary.Stats.Buckets); i++ {
					Expect(userCGMSummary.Stats.Buckets[i]).ToNot(BeNil())
				}
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := lowBloodGlucose
				userCGMSummary = types.Create[*types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userCGMSummary.Stats.Buckets)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", inTargetBloodGlucose*12*5, 0.001))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, secondDatumTime, 168, secondRequestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(528)) // 22 days

				By("check total glucose and dates for second batch")
				newHourlyStatsLen = len(userCGMSummary.Stats.Buckets)
				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetCGMData)/12 // 12 per day, need length without the gap
				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", secondRequestedAvgGlucose*12*5))

					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("check total glucose and dates for gap")
				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64(0)))
				}
			})

			It("Returns correct stats when given multiple batches in a day", func() {
				var incrementalDatumTime time.Time
				var lastRecordTime time.Time
				userCGMSummary = types.Create[*types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 144, inTargetBloodGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(144))

				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetCGMData = NewDataSetCGMDataAvg(deviceId, incrementalDatumTime, 1, float64(i))

					err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(144 + i))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
				}

				for i := 144; i < len(userCGMSummary.Stats.Buckets); i++ {
					f := fmt.Sprintf("hour %d", i)
					By(f)
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*5)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", float64((i-143)*12*5), 0.001))

					averageGlucoseMmol := userCGMSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes)
					Expect(averageGlucoseMmol).To(BeNumerically("~", i-143))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userCGMSummary = types.Create[*types.CGMStats](userId)

				// Datasets use +1 and +2 offset to allow for checking via iteration
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -2), 24, inTargetBloodGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -1), 24, inTargetBloodGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose+2)
				dataSetCGMData = append(dataSetCGMDataOne, dataSetCGMDataTwo...)
				dataSetCGMData = append(dataSetCGMData, dataSetCGMDataThree...)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(72))

				for i := len(userCGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userCGMSummary.Stats.Buckets)-i-1) - 5*time.Minute)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))

					expectedTotalGlucose = (inTargetBloodGlucose + float64(i/24)) * 12 * 5
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", expectedTotalGlucose, 0.001))
				}
			})

			It("Returns correct hourly stats for hours with different Time in Range", func() {
				var lastRecordTime time.Time
				userCGMSummary = types.Create[*types.CGMStats](userId)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetCGMDataOne := NewDataSetCGMDataRanges(deviceId, datumTime.Add(-4*time.Hour), 1, veryLowRange)
				dataSetCGMDataTwo := NewDataSetCGMDataRanges(deviceId, datumTime.Add(-3*time.Hour), 1, lowRange)
				dataSetCGMDataThree := NewDataSetCGMDataRanges(deviceId, datumTime.Add(-2*time.Hour), 1, inRange)
				dataSetCGMDataFour := NewDataSetCGMDataRanges(deviceId, datumTime.Add(-1*time.Hour), 1, highRange)
				dataSetCGMDataFive := NewDataSetCGMDataRanges(deviceId, datumTime, 1, veryHighRange)

				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataOne)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataTwo)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataThree)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataFour)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataFive)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(5))

				By("check record counters for insurance")
				for i := len(userCGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(10))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(50))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userCGMSummary.Stats.Buckets)-i-1) - time.Minute*5)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("very low minutes")
				Expect(userCGMSummary.Stats.Buckets[0].Data.VeryLowMinutes).To(Equal(50))
				Expect(userCGMSummary.Stats.Buckets[0].Data.LowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.TargetMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.HighMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.VeryHighMinutes).To(Equal(0))

				By("very low records")
				Expect(userCGMSummary.Stats.Buckets[0].Data.VeryLowRecords).To(Equal(10))
				Expect(userCGMSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.TargetRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.HighRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[0].Data.VeryHighRecords).To(Equal(0))

				By("low minutes")
				Expect(userCGMSummary.Stats.Buckets[1].Data.VeryLowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.LowMinutes).To(Equal(50))
				Expect(userCGMSummary.Stats.Buckets[1].Data.TargetMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.HighMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.VeryHighMinutes).To(Equal(0))

				By("low records")
				Expect(userCGMSummary.Stats.Buckets[1].Data.VeryLowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.LowRecords).To(Equal(10))
				Expect(userCGMSummary.Stats.Buckets[1].Data.TargetRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.HighRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[1].Data.VeryHighRecords).To(Equal(0))

				By("in-range minutes")
				Expect(userCGMSummary.Stats.Buckets[2].Data.VeryLowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.LowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.TargetMinutes).To(Equal(50))
				Expect(userCGMSummary.Stats.Buckets[2].Data.HighMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.VeryHighMinutes).To(Equal(0))

				By("in-range records")
				Expect(userCGMSummary.Stats.Buckets[2].Data.VeryLowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.LowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.TargetRecords).To(Equal(10))
				Expect(userCGMSummary.Stats.Buckets[2].Data.HighRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[2].Data.VeryHighRecords).To(Equal(0))

				By("high minutes")
				Expect(userCGMSummary.Stats.Buckets[3].Data.VeryLowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.LowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.TargetMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.HighMinutes).To(Equal(50))
				Expect(userCGMSummary.Stats.Buckets[3].Data.VeryHighMinutes).To(Equal(0))

				By("high records")
				Expect(userCGMSummary.Stats.Buckets[3].Data.VeryLowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.LowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.TargetRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[3].Data.HighRecords).To(Equal(10))
				Expect(userCGMSummary.Stats.Buckets[3].Data.VeryHighRecords).To(Equal(0))

				By("very high minutes")
				Expect(userCGMSummary.Stats.Buckets[4].Data.VeryLowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.LowMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.TargetMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.HighMinutes).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.VeryHighMinutes).To(Equal(50))

				By("very high records")
				Expect(userCGMSummary.Stats.Buckets[4].Data.VeryLowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.LowRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.TargetRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.HighRecords).To(Equal(0))
				Expect(userCGMSummary.Stats.Buckets[4].Data.VeryHighRecords).To(Equal(10))
			})

			It("Returns correct hourly stats for days uploaded in reverse", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userCGMSummary = types.Create[*types.CGMStats](userId)

				// Datasets use +1 and +2 offset to allow for checking via iteration
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -2), 24, inTargetBloodGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -1), 24, inTargetBloodGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose+2)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataThree)
				Expect(err).ToNot(HaveOccurred())

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataTwo)
				Expect(err).ToNot(HaveOccurred())

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMDataOne)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(72))

				for i := len(userCGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
					By(fmt.Sprintf("hour %d", i+1))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(12))
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userCGMSummary.Stats.Buckets)-i-1) - 5*time.Minute)
					Expect(userCGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))

					expectedTotalGlucose = (inTargetBloodGlucose + float64(i/24)) * 12 * 5
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", expectedTotalGlucose, 0.001))
				}
			})
		})

		Context("CalculateDelta", func() {
			It("Returns correct deltas for periods", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				for i, period := range periodKeys {
					vFloat := float64(i) * 7.5
					vFloatTwo := vFloat * 2
					vInt := i * 7
					vIntTwo := vInt * 2

					userCGMSummary.Stats.Periods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          pointer.FromAny(vFloat),
						TimeCGMUseMinutes:          pointer.FromAny(vInt),
						TimeCGMUseRecords:          pointer.FromAny(vInt),
						AverageGlucoseMmol:         pointer.FromAny(vFloat),
						GlucoseManagementIndicator: pointer.FromAny(vFloat),
						TotalRecords:               pointer.FromAny(vInt),
						AverageDailyRecords:        pointer.FromAny(vFloat),
						TimeInTargetPercent:        pointer.FromAny(vFloat),
						TimeInTargetMinutes:        pointer.FromAny(vInt),
						TimeInTargetRecords:        pointer.FromAny(vInt),
						TimeInLowPercent:           pointer.FromAny(vFloat),
						TimeInLowMinutes:           pointer.FromAny(vInt),
						TimeInLowRecords:           pointer.FromAny(vInt),
						TimeInVeryLowPercent:       pointer.FromAny(vFloat),
						TimeInVeryLowMinutes:       pointer.FromAny(vInt),
						TimeInVeryLowRecords:       pointer.FromAny(vInt),
						TimeInHighPercent:          pointer.FromAny(vFloat),
						TimeInHighMinutes:          pointer.FromAny(vInt),
						TimeInHighRecords:          pointer.FromAny(vInt),
						TimeInVeryHighPercent:      pointer.FromAny(vFloat),
						TimeInVeryHighMinutes:      pointer.FromAny(vInt),
						TimeInVeryHighRecords:      pointer.FromAny(vInt),
					}

					userCGMSummary.Stats.OffsetPeriods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          pointer.FromAny(vFloatTwo),
						TimeCGMUseMinutes:          pointer.FromAny(vIntTwo),
						TimeCGMUseRecords:          pointer.FromAny(vIntTwo),
						AverageGlucoseMmol:         pointer.FromAny(vFloatTwo),
						GlucoseManagementIndicator: pointer.FromAny(vFloatTwo),
						TotalRecords:               pointer.FromAny(vIntTwo),
						AverageDailyRecords:        pointer.FromAny(vFloatTwo),
						TimeInTargetPercent:        pointer.FromAny(vFloatTwo),
						TimeInTargetMinutes:        pointer.FromAny(vIntTwo),
						TimeInTargetRecords:        pointer.FromAny(vIntTwo),
						TimeInLowPercent:           pointer.FromAny(vFloatTwo),
						TimeInLowMinutes:           pointer.FromAny(vIntTwo),
						TimeInLowRecords:           pointer.FromAny(vIntTwo),
						TimeInVeryLowPercent:       pointer.FromAny(vFloatTwo),
						TimeInVeryLowMinutes:       pointer.FromAny(vIntTwo),
						TimeInVeryLowRecords:       pointer.FromAny(vIntTwo),
						TimeInHighPercent:          pointer.FromAny(vFloatTwo),
						TimeInHighMinutes:          pointer.FromAny(vIntTwo),
						TimeInHighRecords:          pointer.FromAny(vIntTwo),
						TimeInVeryHighPercent:      pointer.FromAny(vFloatTwo),
						TimeInVeryHighMinutes:      pointer.FromAny(vIntTwo),
						TimeInVeryHighRecords:      pointer.FromAny(vIntTwo),
					}
				}

				userCGMSummary.Stats.CalculateDelta()

				for i, period := range periodKeys {
					floatDiff := float64(i)*7.5 - float64(i)*7.5*2
					intDiff := i*7 - i*7*2

					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUsePercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(Equal(-floatDiff))

					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicatorDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].GlucoseManagementIndicatorDelta).To(Equal(-floatDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TotalRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(Equal(-floatDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInTargetMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInLowMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInLowMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryLowMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInHighMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInHighMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(Equal(-intDiff))

					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(Equal(floatDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(Equal(-floatDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryHighMinutesDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighMinutesDelta).To(Equal(-intDiff))
					Expect(*userCGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(Equal(intDiff))
					Expect(*userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(Equal(-intDiff))
				}
			})

			It("Returns correct nil deltas with nil latest stats", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				for _, period := range periodKeys {
					userCGMSummary.Stats.Periods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          nil,
						TimeCGMUseMinutes:          nil,
						TimeCGMUseRecords:          nil,
						AverageGlucoseMmol:         nil,
						GlucoseManagementIndicator: nil,
						TotalRecords:               nil,
						AverageDailyRecords:        nil,
						TimeInTargetPercent:        nil,
						TimeInTargetMinutes:        nil,
						TimeInTargetRecords:        nil,
						TimeInLowPercent:           nil,
						TimeInLowMinutes:           nil,
						TimeInLowRecords:           nil,
						TimeInVeryLowPercent:       nil,
						TimeInVeryLowMinutes:       nil,
						TimeInVeryLowRecords:       nil,
						TimeInHighPercent:          nil,
						TimeInHighMinutes:          nil,
						TimeInHighRecords:          nil,
						TimeInVeryHighPercent:      nil,
						TimeInVeryHighMinutes:      nil,
						TimeInVeryHighRecords:      nil,
					}

					userCGMSummary.Stats.OffsetPeriods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          pointer.FromAny(1.0),
						TimeCGMUseMinutes:          pointer.FromAny(1),
						TimeCGMUseRecords:          pointer.FromAny(1),
						AverageGlucoseMmol:         pointer.FromAny(1.0),
						GlucoseManagementIndicator: pointer.FromAny(1.0),
						TotalRecords:               pointer.FromAny(1),
						AverageDailyRecords:        pointer.FromAny(1.0),
						TimeInTargetPercent:        pointer.FromAny(1.0),
						TimeInTargetMinutes:        pointer.FromAny(1),
						TimeInTargetRecords:        pointer.FromAny(1),
						TimeInLowPercent:           pointer.FromAny(1.0),
						TimeInLowMinutes:           pointer.FromAny(1),
						TimeInLowRecords:           pointer.FromAny(1),
						TimeInVeryLowPercent:       pointer.FromAny(1.0),
						TimeInVeryLowMinutes:       pointer.FromAny(1),
						TimeInVeryLowRecords:       pointer.FromAny(1),
						TimeInHighPercent:          pointer.FromAny(1.0),
						TimeInHighMinutes:          pointer.FromAny(1),
						TimeInHighRecords:          pointer.FromAny(1),
						TimeInVeryHighPercent:      pointer.FromAny(1.0),
						TimeInVeryHighMinutes:      pointer.FromAny(1),
						TimeInVeryHighRecords:      pointer.FromAny(1),
					}
				}

				userCGMSummary.Stats.CalculateDelta()

				for _, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUsePercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUsePercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicatorDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].GlucoseManagementIndicatorDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TotalRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(BeNil())
				}
			})

			It("Returns correct nil deltas with nil offset stats", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)

				for _, period := range periodKeys {
					userCGMSummary.Stats.Periods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          pointer.FromAny(1.0),
						TimeCGMUseMinutes:          pointer.FromAny(1),
						TimeCGMUseRecords:          pointer.FromAny(1),
						AverageGlucoseMmol:         pointer.FromAny(1.0),
						GlucoseManagementIndicator: pointer.FromAny(1.0),
						TotalRecords:               pointer.FromAny(1),
						AverageDailyRecords:        pointer.FromAny(1.0),
						TimeInTargetPercent:        pointer.FromAny(1.0),
						TimeInTargetMinutes:        pointer.FromAny(1),
						TimeInTargetRecords:        pointer.FromAny(1),
						TimeInLowPercent:           pointer.FromAny(1.0),
						TimeInLowMinutes:           pointer.FromAny(1),
						TimeInLowRecords:           pointer.FromAny(1),
						TimeInVeryLowPercent:       pointer.FromAny(1.0),
						TimeInVeryLowMinutes:       pointer.FromAny(1),
						TimeInVeryLowRecords:       pointer.FromAny(1),
						TimeInHighPercent:          pointer.FromAny(1.0),
						TimeInHighMinutes:          pointer.FromAny(1),
						TimeInHighRecords:          pointer.FromAny(1),
						TimeInVeryHighPercent:      pointer.FromAny(1.0),
						TimeInVeryHighMinutes:      pointer.FromAny(1),
						TimeInVeryHighRecords:      pointer.FromAny(1),
					}

					userCGMSummary.Stats.OffsetPeriods[period] = &types.CGMPeriod{
						TimeCGMUsePercent:          nil,
						TimeCGMUseMinutes:          nil,
						TimeCGMUseRecords:          nil,
						AverageGlucoseMmol:         nil,
						GlucoseManagementIndicator: nil,
						TotalRecords:               nil,
						AverageDailyRecords:        nil,
						TimeInTargetPercent:        nil,
						TimeInTargetMinutes:        nil,
						TimeInTargetRecords:        nil,
						TimeInLowPercent:           nil,
						TimeInLowMinutes:           nil,
						TimeInLowRecords:           nil,
						TimeInVeryLowPercent:       nil,
						TimeInVeryLowMinutes:       nil,
						TimeInVeryLowRecords:       nil,
						TimeInHighPercent:          nil,
						TimeInHighMinutes:          nil,
						TimeInHighRecords:          nil,
						TimeInVeryHighPercent:      nil,
						TimeInVeryHighMinutes:      nil,
						TimeInVeryHighRecords:      nil,
					}
				}

				userCGMSummary.Stats.CalculateDelta()

				for _, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUsePercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUsePercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeCGMUseRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].AverageGlucoseMmolDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].AverageGlucoseMmolDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicatorDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].GlucoseManagementIndicatorDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TotalRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TotalRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].AverageDailyRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].AverageDailyRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInTargetRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInTargetRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInLowRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInLowRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryLowRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryLowRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInHighRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInHighRecordsDelta).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighPercentDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighMinutesDelta).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].TimeInVeryHighRecordsDelta).To(BeNil())
					Expect(userCGMSummary.Stats.OffsetPeriods[period].TimeInVeryHighRecordsDelta).To(BeNil())
				}
			})
		})

		Context("CalculateSummary/Update", func() {
			var newDatumTime time.Time
			var dataSetCGMDataCursor *mongo.Cursor

			It("Returns correct time in range for stats", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				ranges := NewDataRanges()
				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime, 720, ranges)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				stopPoints := []int{1, 7, 14, 30}
				for _, v := range stopPoints {
					periodKey := strconv.Itoa(v) + "d"

					f := fmt.Sprintf("period %s", periodKey)
					By(f)

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInTargetMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInTargetMinutes).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInTargetRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInTargetRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInTargetPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInTargetPercent).To(Equal(0.200))

					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowMinutes).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowPercent).To(Equal(0.200))

					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInAnyLowMinutes).To(Equal(240 * 2 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInAnyLowMinutes).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInAnyLowRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInAnyLowRecords).To(Equal(48 * 2 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInAnyLowPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInAnyLowPercent).To(Equal(0.400))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInLowMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInLowMinutes).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInLowRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInLowRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInLowPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInLowPercent).To(Equal(0.200))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInHighMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInHighMinutes).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInHighRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInHighRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInHighPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInHighPercent).To(Equal(0.200))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighMinutes).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighMinutes).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTotalRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TotalRecords).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasAverageDailyRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].AverageDailyRecords).To(BeNumerically("==", 240))

					// ranges calc only generates 83.3% of an hour, each hour needs to be divisible by 5
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeCGMUseMinutes).To(Equal(1200 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeCGMUseRecords).To(Equal(240 * v))

					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeCGMUsePercent).To(BeNumerically("~", 0.833, 0.001))
				}
			})

			It("Returns correct average glucose for stats", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				expectedGMI := types.CalculateGMI(inTargetBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(Equal(inTargetBloodGlucose))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
				}
			})

			It("Correctly removes GMI when CGM use drop below 0.7", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				expectedGMI := types.CalculateGMI(inTargetBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(Equal(inTargetBloodGlucose))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
				}

				// start the real test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, 31), 16, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1440))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(60 * 24)) // 60 days currently capped

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(
						BeNumerically("~", 960/(float64(periodInts[i])*1440), 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(192))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(960))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(Equal(inTargetBloodGlucose))
				}
			})

			It("Returns correctly calculated summary with no rolling", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				expectedGMI := types.CalculateGMI(inTargetBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", inTargetBloodGlucose, 0.001))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				}
			})

			It("Returns correctly calculated summary with rolling <100% cgm use", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				expectedGMI := types.CalculateGMI(highBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 1, lowBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(
						BeNumerically("~", 60/(float64(periodInts[i])*1440), 0.006))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(12))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(60))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))

					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 720, highBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(721))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(721))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				}
			})

			It("Returns correctly calculated summary with rolling 100% cgm use", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				newDatumTime = datumTime.Add(time.Duration(23) * time.Hour)
				expectedGMIFirst := types.CalculateGMI(lowBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, lowBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.005))

					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 23, highBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(743))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(743))

				for i, period := range periodKeys {
					expectedAverage := ExpectedAverage(periodInts[i]*24, 23, highBloodGlucose, lowBloodGlucose)
					expectedGMI := types.CalculateGMI(expectedAverage)
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", expectedAverage, 0.005))

					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}
			})

			It("Returns correctly non-rolling summary with two 30 day windows", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 31)
				expectedGMISecond := types.CalculateGMI(highBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, highBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(24))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440/(1440*float64(periodInts[i])), 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))

					if *userCGMSummary.Stats.Periods[period].TimeCGMUsePercent > 0.7 {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					} else {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
						Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					}
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 168, highBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(768))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(768)) // 30 days

				for i, period := range periodKeys {
					if i == 0 || i == 1 {
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(288 * periodInts[i]))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(1440 * periodInts[i]))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					} else {
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(7 * 288))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(7 * 1440))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440*7/(1440*float64(periodInts[i])), 0.005))
					}

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))

					if *userCGMSummary.Stats.Periods[period].TimeCGMUsePercent > 0.7 {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
						Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
					} else {
						Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
					}
				}
			})

			It("Returns correctly calculated summary with rolling dropping cgm use", func() {
				userCGMSummary = types.Create[*types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				expectedGMI := types.CalculateGMI(lowBloodGlucose)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, lowBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 1, highBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1440)) // 60 days

				for _, period := range periodKeys {
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 0.03, 0.03))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseRecords).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(12))

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUseMinutes).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(60))

					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucoseMmol).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.05))

					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
				}
			})

			//It("Returns correct record count when given single buckets in strange places", func() {
			//	userCGMSummary = types.Create[*types.CGMStats](userId)
			//
			//	// initial single bucket
			//	dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceId, datumTime, 1, inTargetBloodGlucose)
			//
			//	// add another single bucket forward to check off-by-one
			//	dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.Add(1*time.Hour), 1, inTargetBloodGlucose)
			//
			//	// 1 bucket gap
			//	dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceId, datumTime.Add(3*time.Hour), 1, inTargetBloodGlucose)
			//
			//	// 0 bucket gap, with overlap with previous
			//	dataSetCGMDataFour := NewDataSetCGMDataAvg(deviceId, datumTime.Add(3.5*60*time.Minute), 1, inTargetBloodGlucose)
			//
			//	// same bucket as before
			//	dataSetCGMDataFive := NewDataSetCGMDataAvg(deviceId, datumTime.Add(4*60*time.Minute), 1, inTargetBloodGlucose)
			//
			//	// lots of buckets ahead
			//	dataSetCGMDataSix := NewDataSetCGMDataAvg(deviceId, datumTime.Add(14*24*time.Hour), 1, inTargetBloodGlucose)
			//
			//	allDataSet := make([]*glucose.Glucose, 0, len(dataSetCGMDataOne)+len(dataSetCGMDataTwo)+len(dataSetCGMDataThree)+len(dataSetCGMDataFour)+len(dataSetCGMDataFive)+len(dataSetCGMDataSix))
			//
			//	err = userCGMSummary.Stats.Update(allDataSet)
			//	Expect(err).ToNot(HaveOccurred())
			//})

			It("Returns correctly moving offset periods", func() {
				// Here we generate 5 1d datasets, and add them in a specific order:
				// -59d -- target glucose
				// -27d -- veryHigh glucose
				// -13d -- veryLow glucose
				//  -1d -- high glucose
				//   0d -- low glucose

				// This should result in:
				//  1d regular -- low, 288 readings (from 0d)
				//  1d offset  -- high, 288 readings (from 1d)
				//  7d regular -- (high+low)/2, 288*2 (576) readings (from 0d + 1d)
				//  7d offset  -- veryLow, 288 readings (from 14d)
				// 14d regular -- (high+low+veryLow)/3, 288*3 (864) readings (from 1d + 2d + 14d)
				// 14d offset  -- veryHigh, 288 readings (from 28d)
				// 30d regular -- (high+veryHigh+low+veryLow)/4, 288*4 (1152) readings (from 1d + 2d + 14d + 28d)
				// 30d offset  -- target, 288 readings (from 60d)

				userCGMSummary = types.Create[*types.CGMStats](userId)

				newDatumTimeOne := datumTime.AddDate(0, 0, -59)
				newDatumTimeTwo := datumTime.AddDate(0, 0, -27)
				newDatumTimeThree := datumTime.AddDate(0, 0, -13)
				newDatumTimeFour := datumTime.AddDate(0, 0, -1)
				newDatumTimeFive := datumTime

				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceId, newDatumTimeOne, 24, inTargetBloodGlucose)
				dataSetCGMDataOneCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMDataOne), nil, nil)

				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, newDatumTimeTwo, 24, veryHighBloodGlucose)
				dataSetCGMDataTwoCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMDataTwo), nil, nil)

				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceId, newDatumTimeThree, 24, veryLowBloodGlucose)
				dataSetCGMDataThreeCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMDataThree), nil, nil)

				dataSetCGMDataFour := NewDataSetCGMDataAvg(deviceId, newDatumTimeFour, 24, highBloodGlucose)
				dataSetCGMDataFourCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMDataFour), nil, nil)

				dataSetCGMDataFive := NewDataSetCGMDataAvg(deviceId, newDatumTimeFive, 24, lowBloodGlucose)
				dataSetCGMDataFiveCursor, err := mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMDataFive), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataOneCursor)
				Expect(err).ToNot(HaveOccurred())

				// first day, should have 24 buckets
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(24))
				Expect(*userCGMSummary.Stats.Periods["1d"].TotalRecords).To(Equal(24 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["1d"].TotalRecords).To(Equal(0))
				Expect(*userCGMSummary.Stats.Periods["7d"].TotalRecords).To(Equal(24 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["7d"].TotalRecords).To(Equal(0))

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataTwoCursor)
				Expect(err).ToNot(HaveOccurred())

				// 33 days elapsed, should have 33*24 (792) buckets
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(792))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(792))
				Expect(*userCGMSummary.Stats.Periods["14d"].TotalRecords).To(Equal(24 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["14d"].TotalRecords).To(Equal(0))
				Expect(*userCGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 12))

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataThreeCursor)
				Expect(err).ToNot(HaveOccurred())

				// 47 days elapsed, should have 47*24 (1128) buckets
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1128))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1128))
				Expect(*userCGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 2 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 12))

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataFourCursor)
				Expect(err).ToNot(HaveOccurred())

				// 59 days elapsed, should have 59*24 (1416) buckets
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1416))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1416))
				Expect(*userCGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 3 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 1 * 12))

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataFiveCursor)
				Expect(err).ToNot(HaveOccurred())

				// 60 days elapsed, should have 60*24 (1440) buckets
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1440))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1440))
				Expect(*userCGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(24 * 4 * 12))
				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(24 * 1 * 12))

				// check that the data matches the expectation described at the top of the test
				Expect(*userCGMSummary.Stats.Periods["1d"].AverageGlucoseMmol).To(BeNumerically("~", lowBloodGlucose, 0.001))
				Expect(*userCGMSummary.Stats.Periods["1d"].TotalRecords).To(Equal(288))

				Expect(*userCGMSummary.Stats.OffsetPeriods["1d"].AverageGlucoseMmol).To(BeNumerically("~", highBloodGlucose, 0.001))
				Expect(*userCGMSummary.Stats.OffsetPeriods["1d"].TotalRecords).To(Equal(288))

				Expect(*userCGMSummary.Stats.Periods["7d"].AverageGlucoseMmol).To(BeNumerically("~", (highBloodGlucose+lowBloodGlucose)/2, 0.001))
				Expect(*userCGMSummary.Stats.Periods["7d"].TotalRecords).To(Equal(288 * 2))

				Expect(*userCGMSummary.Stats.OffsetPeriods["7d"].AverageGlucoseMmol).To(BeNumerically("~", veryLowBloodGlucose, 0.001))
				Expect(*userCGMSummary.Stats.OffsetPeriods["7d"].TotalRecords).To(Equal(288))

				Expect(*userCGMSummary.Stats.Periods["14d"].AverageGlucoseMmol).To(BeNumerically("~", (highBloodGlucose+lowBloodGlucose+veryLowBloodGlucose)/3, 0.001))
				Expect(*userCGMSummary.Stats.Periods["14d"].TotalRecords).To(Equal(288 * 3))

				Expect(*userCGMSummary.Stats.OffsetPeriods["14d"].AverageGlucoseMmol).To(BeNumerically("~", veryHighBloodGlucose, 0.001))
				Expect(*userCGMSummary.Stats.OffsetPeriods["14d"].TotalRecords).To(Equal(288))

				Expect(*userCGMSummary.Stats.Periods["30d"].AverageGlucoseMmol).To(BeNumerically("~", (veryHighBloodGlucose+highBloodGlucose+lowBloodGlucose+veryLowBloodGlucose)/4, 0.001))
				Expect(*userCGMSummary.Stats.Periods["30d"].TotalRecords).To(Equal(288 * 4))

				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].AverageGlucoseMmol).To(BeNumerically("~", inTargetBloodGlucose, 0.001))
				Expect(*userCGMSummary.Stats.OffsetPeriods["30d"].TotalRecords).To(Equal(288))
			})
		})

		Context("ClearInvalidatedBuckets", func() {
			It("trims the correct buckets", func() {
				var dataSetCGMDataCursor *mongo.Cursor
				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 10, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(10))

				status := &types.UserLastUpdated{EarliestModified: datumTime.Add(-5 * time.Hour)}

				userCGMSummary.Stats.ClearInvalidatedBuckets(status)

				// we have the right length
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(5))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userCGMSummary.Stats.Buckets[len(userCGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
			})

			It("doesnt trim if only modified in the future", func() {
				var dataSetCGMDataCursor *mongo.Cursor
				userCGMSummary = types.Create[*types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 10, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(10))

				status := &types.UserLastUpdated{EarliestModified: datumTime.Add(time.Hour)}

				userCGMSummary.Stats.ClearInvalidatedBuckets(status)

				// we have the right length
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userCGMSummary.Stats.Buckets[len(userCGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
			})

			It("doesnt trim if only modified on the same hour, but after the bucket time", func() {
				var dataSetCGMDataCursor *mongo.Cursor
				userCGMSummary = types.Create[*types.CGMStats](userId)
				midDatumTime := datumTime.Add(30 * time.Minute)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, midDatumTime, 9, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(10))

				status := &types.UserLastUpdated{EarliestModified: midDatumTime.Add(10 * time.Minute)}

				userCGMSummary.Stats.ClearInvalidatedBuckets(status)

				// we have the right length
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userCGMSummary.Stats.Buckets[len(userCGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
			})

			It("trims if modified on the same hour, and before the bucket time", func() {
				var dataSetCGMDataCursor *mongo.Cursor
				userCGMSummary = types.Create[*types.CGMStats](userId)
				midDatumTime := datumTime.Add(30 * time.Minute)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, midDatumTime, 9, inTargetBloodGlucose)
				dataSetCGMDataCursor, err = mongo.NewCursorFromDocuments(ConvertToIntArray(dataSetCGMData), nil, nil)

				err = userCGMSummary.Stats.Update(ctx, dataSetCGMDataCursor)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(10))
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(10))

				status := &types.UserLastUpdated{EarliestModified: midDatumTime.Add(-10 * time.Minute)}

				userCGMSummary.Stats.ClearInvalidatedBuckets(status)

				// we have the right length
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(9))

				// we didn't overshoot and nil something we shouldn't have
				Expect(userCGMSummary.Stats.Buckets[len(userCGMSummary.Stats.Buckets)-1]).ToNot(BeNil())
			})
		})
	})
})
