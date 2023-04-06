package types_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo"
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
	var err error
	var dataSetCGMData []*glucose.Glucose

	BeforeEach(func() {
		userId = userTest.RandomID()
		deviceId = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("CreateCGMSummary", func() {
		It("Correctly initializes a cgm summary", func() {
			summary := types.Create[types.CGMStats, *types.CGMStats](userId)
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
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 336, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct hour count when given 1 week", func() {
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 168, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct hour count when given 3 weeks", func() {
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 504, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct record count when given overlapping records", func() {
				var doubledCGMData = make([]*glucose.Glucose, 288*2)

				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)

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
				userCGMSummary = types.Create[types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userCGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(12))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := requestedAvgGlucose - 4
				userCGMSummary = types.Create[types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 168, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userCGMSummary.Stats.Buckets)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", requestedAvgGlucose*12*5, 0.001))

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
				userCGMSummary = types.Create[types.CGMStats](userId)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 144, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(144))

				// TODO move to 0.5 hour to test more cases
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

					averageGlucose := userCGMSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes)
					Expect(averageGlucose).To(BeNumerically("~", i-143))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceId, datumTime, 24, requestedAvgGlucose+2)
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

					expectedTotalGlucose = (requestedAvgGlucose + float64(i/24)) * 12 * 5
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", expectedTotalGlucose, 0.001))
				}
			})

			It("Returns correct hourly stats for hours with different Time in Range", func() {
				var lastRecordTime time.Time
				userCGMSummary = types.Create[types.CGMStats](userId)
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
		})

		Context("CalculateSummary", func() {
			var newDatumTime time.Time

			It("Returns correct time in range for stats", func() {
				var expectedCGMUse float64
				userCGMSummary = types.Create[types.CGMStats](userId)
				ranges := NewDataRanges()
				dataSetCGMData = NewDataSetCGMDataRanges(deviceId, datumTime, 720, ranges)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				stopPoints := []int{1, 7, 14, 30}
				for _, v := range stopPoints {
					periodKey := strconv.Itoa(v) + "d"

					f := fmt.Sprintf("period %s", periodKey)
					By(f)

					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInTargetMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInTargetRecords).To(Equal(48 * v))
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInTargetPercent).To(Equal(0.200))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInTargetPercent).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowRecords).To(Equal(48 * v))
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryLowPercent).To(Equal(0.200))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowPercent).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInLowMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInLowRecords).To(Equal(48 * v))
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInLowPercent).To(Equal(0.200))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInLowPercent).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInHighMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInHighRecords).To(Equal(48 * v))
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInHighPercent).To(Equal(0.200))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInHighPercent).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighMinutes).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))
					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())

					// ranges calc only generates 83.3% of an hour, each hour needs to be divisible by 5
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeCGMUseMinutes).To(Equal(1200 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].TimeCGMUseRecords).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].HasTimeCGMUsePercent).To(BeTrue())

					Expect(userCGMSummary.Stats.Periods[periodKey].TotalRecords).To(Equal(240 * v))
					Expect(userCGMSummary.Stats.Periods[periodKey].AverageDailyRecords).To(BeNumerically("==", 240))

					// this value is a bit funny, its 83.3%, but the missing end of the final day gets compensated off
					// resulting in 83.6% only on the first day
					if v == 1 {
						expectedCGMUse = 0.836
					} else {
						expectedCGMUse = 0.833
					}

					Expect(*userCGMSummary.Stats.Periods[periodKey].TimeCGMUsePercent).To(BeNumerically("~", expectedCGMUse, 0.001))
				}
			})

			It("Returns correct average glucose for stats", func() {
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, requestedAvgGlucose)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}
			})

			It("Correctly removes GMI when CGM use drop below 0.7", func() {
				userCGMSummary = types.Create[types.CGMStats](userId)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, requestedAvgGlucose)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}

				// start the real test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime.AddDate(0, 0, 31), 16, requestedAvgGlucose)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720)) // hits 4 days over 30 day cap

				userCGMSummary.Stats.CalculateSummary()

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(30 * 24)) // 30 days currently capped
				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(
						BeNumerically("~", 960/(float64(periodInts[i])*1440), 0.005))
					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())

					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(192))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(960))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with no rolling", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, requestedAvgGlucose)
				userCGMSummary = types.Create[types.CGMStats](userId)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with rolling <100% cgm use", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 1, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose + 4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(1))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(
						BeNumerically("~", 60/(float64(periodInts[i])*1440), 0.006))

					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(12))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(60))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 720, requestedAvgGlucose+4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with rolling 100% cgm use", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userId)
				newDatumTime = datumTime.Add(time.Duration(23) * time.Hour)
				expectedGMIFirst := types.CalculateGMI(requestedAvgGlucose - 4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for _, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 23, requestedAvgGlucose+4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					expectedAverage := ExpectedAverage(periodInts[i]*24, 23, requestedAvgGlucose+4, requestedAvgGlucose-4)
					expectedGMI := types.CalculateGMI(expectedAverage)
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", expectedAverage, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}
			})

			It("Returns correctly non-rolling summary with two 30 day windows", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 24, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 31)
				expectedGMISecond := types.CalculateGMI(requestedAvgGlucose + 4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(24))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(24))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440/(1440*float64(periodInts[i])), 0.005))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(288))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(1440))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					if *userCGMSummary.Stats.Periods[period].TimeCGMUsePercent > 0.7 {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					} else {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
						Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					}
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 168, requestedAvgGlucose+4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720)) // 30 days

				for i, period := range periodKeys {
					if i == 0 || i == 1 {
						Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(288 * periodInts[i]))
						Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(1440 * periodInts[i]))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					} else {
						Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(7 * 288))
						Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(7 * 1440))
						Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440*7/(1440*float64(periodInts[i])), 0.005))
					}

					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					if *userCGMSummary.Stats.Periods[period].TimeCGMUsePercent > 0.7 {
						Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
					} else {
						Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
						Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					}
				}
			})

			It("Returns correctly calculated summary with rolling dropping cgm use", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, datumTime, 720, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose - 4)

				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(720))

				userCGMSummary.Stats.CalculateSummary()
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(*userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
				}

				// start the actual test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceId, newDatumTime, 1, requestedAvgGlucose+4)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720)) // 30 days

				for _, period := range periodKeys {
					Expect(*userCGMSummary.Stats.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 0.03, 0.03))
					Expect(userCGMSummary.Stats.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseRecords).To(Equal(12))
					Expect(userCGMSummary.Stats.Periods[period].TimeCGMUseMinutes).To(Equal(60))
					Expect(userCGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.05))
					Expect(userCGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
					Expect(userCGMSummary.Stats.Periods[period].GlucoseManagementIndicator).To(BeNil())
					Expect(userCGMSummary.Stats.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
				}
			})
		})
	})
})
