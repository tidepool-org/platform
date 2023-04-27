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

func NewDataSetBGMDataAvg(deviceId string, startTime time.Time, hours float64, reqAvg float64) []*glucose.Glucose {
	requiredRecords := int(hours * 6)
	typ := pointer.FromString("smbg")

	var dataSetData = make([]*glucose.Glucose, requiredRecords)

	// generate X hours of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 10)

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceId)
			datum.Value = pointer.FromFloat64(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

func NewDataSetBGMDataRanges(deviceId string, startTime time.Time, hours float64, ranges DataRanges) []*glucose.Glucose {
	requiredRecords := int(hours * 5)
	typ := pointer.FromString("smbg")

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
		for i, bracket := range glucoseBrackets {
			datumTime := startTime.Add(-time.Duration(count+i+1) * time.Minute * 12)

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceId)
			datum.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

var _ = Describe("BGM Summary", func() {
	var userId string
	var datumTime time.Time
	var deviceId string
	var err error
	var dataSetBGMData []*glucose.Glucose

	BeforeEach(func() {
		userId = userTest.RandomID()
		deviceId = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("CreateBGMSummary", func() {
		It("Correctly initializes a bgm summary", func() {
			summary := types.Create[types.BGMStats, *types.BGMStats](userId)
			Expect(summary).To(Not(BeNil()))
			Expect(summary.Type).To(Equal("bgm"))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userBGMSummary *types.Summary[types.BGMStats, *types.BGMStats]
		var periodKeys = []string{"1d", "7d", "14d", "30d"}
		var periodInts = []int{1, 7, 14, 30}

		Context("AddData Bucket Testing", func() {
			It("Returns correct hour count when given 2 weeks", func() {
				userBGMSummary = types.Create[types.BGMStats](userId)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 336, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct hour count when given 1 week", func() {
				userBGMSummary = types.Create[types.BGMStats](userId)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct hour count when given 3 weeks", func() {
				userBGMSummary = types.Create[types.BGMStats](userId)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 504, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct record count when given overlapping records", func() {
				// NOTE CGM would filter these, we are testing that they don't get filtered here
				var doubledBGMData = make([]*glucose.Glucose, 288*2)

				userBGMSummary = types.Create[types.BGMStats](userId)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)

				// interlace the lists
				for i := 0; i < len(dataSetBGMData); i += 1 {
					doubledBGMData[i*2] = dataSetBGMData[i]
					doubledBGMData[i*2+1] = dataSetBGMDataTwo[i]
				}
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userBGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(6))
			})

			It("Returns correct record count when given overlapping records across multiple calculations", func() {
				// NOTE CGM would filter these, we are testing that they don't get filtered here
				userBGMSummary = types.Create[types.BGMStats](userId)

				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime.Add(15*time.Second), 24, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))
				Expect(userBGMSummary.Stats.Buckets[0].Data.TotalRecords).To(Equal(12))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := lowBloodGlucose
				userBGMSummary = types.Create[types.BGMStats](userId)

				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 168, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userBGMSummary.Stats.Buckets)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", inTargetBloodGlucose*6, 0.001))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 10*time.Minute)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, secondDatumTime, 168, secondRequestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(528)) // 22 days

				By("check total glucose and dates for second batch")
				newHourlyStatsLen = len(userBGMSummary.Stats.Buckets)
				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetBGMData)/12 // 12 per day, need length without the gap
				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", secondRequestedAvgGlucose*6))

					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 10*time.Minute)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("check total glucose and dates for gap")
				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64(0)))
				}
			})

			It("Returns correct stats when given multiple batches in a day", func() {
				var incrementalDatumTime time.Time
				var lastRecordTime time.Time
				userBGMSummary = types.Create[types.BGMStats](userId)

				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 144, inTargetBloodGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(144))

				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetBGMData = NewDataSetBGMDataAvg(deviceId, incrementalDatumTime, 1, float64(i))

					err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(144 + i))
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))
				}

				for i := 144; i < len(userBGMSummary.Stats.Buckets); i++ {
					f := fmt.Sprintf("hour %d", i)
					By(f)
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))

					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*10)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64((i - 143) * 6)))

					averageGlucose := userBGMSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userBGMSummary.Stats.Buckets[i].Data.TotalRecords)
					Expect(averageGlucose).To(BeNumerically("~", i-143, 0.005))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userBGMSummary = types.Create[types.BGMStats](userId)

				// Datasets use +1 and +2 offset to allow for checking via iteration
				dataSetBGMDataOne := NewDataSetBGMDataAvg(deviceId, datumTime.AddDate(0, 0, -2), 24, inTargetBloodGlucose)
				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceId, datumTime.AddDate(0, 0, -1), 24, inTargetBloodGlucose+1)
				dataSetBGMDataThree := NewDataSetBGMDataAvg(deviceId, datumTime, 24, inTargetBloodGlucose+2)
				dataSetBGMData = append(dataSetBGMDataOne, dataSetBGMDataTwo...)
				dataSetBGMData = append(dataSetBGMData, dataSetBGMDataThree...)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(72))

				for i := len(userBGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(6))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userBGMSummary.Stats.Buckets)-i-1) - 10*time.Minute)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))

					expectedTotalGlucose = (inTargetBloodGlucose + float64(i/24)) * 6
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(BeNumerically("~", expectedTotalGlucose, 0.001))
				}
			})

			It("Returns correct hourly stats for hours with different Time in Range", func() {
				var lastRecordTime time.Time
				userBGMSummary = types.Create[types.BGMStats](userId)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetBGMDataOne := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-4*time.Hour), 1, veryLowRange)
				dataSetBGMDataTwo := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-3*time.Hour), 1, lowRange)
				dataSetBGMDataThree := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-2*time.Hour), 1, inRange)
				dataSetBGMDataFour := NewDataSetBGMDataRanges(deviceId, datumTime.Add(-1*time.Hour), 1, highRange)
				dataSetBGMDataFive := NewDataSetBGMDataRanges(deviceId, datumTime, 1, veryHighRange)

				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataOne)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataTwo)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataThree)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataFour)
				Expect(err).ToNot(HaveOccurred())
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMDataFive)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(5))

				By("check record counters for insurance")
				for i := len(userBGMSummary.Stats.Buckets) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalRecords).To(Equal(5))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userBGMSummary.Stats.Buckets)-i-1) - time.Minute*12)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("very low records")
				Expect(userBGMSummary.Stats.Buckets[0].Data.VeryLowRecords).To(Equal(5))
				Expect(userBGMSummary.Stats.Buckets[0].Data.LowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[0].Data.TargetRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[0].Data.HighRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[0].Data.VeryHighRecords).To(Equal(0))

				By("low records")
				Expect(userBGMSummary.Stats.Buckets[1].Data.VeryLowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[1].Data.LowRecords).To(Equal(5))
				Expect(userBGMSummary.Stats.Buckets[1].Data.TargetRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[1].Data.HighRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[1].Data.VeryHighRecords).To(Equal(0))

				By("in-range records")
				Expect(userBGMSummary.Stats.Buckets[2].Data.VeryLowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[2].Data.LowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[2].Data.TargetRecords).To(Equal(5))
				Expect(userBGMSummary.Stats.Buckets[2].Data.HighRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[2].Data.VeryHighRecords).To(Equal(0))

				By("high records")
				Expect(userBGMSummary.Stats.Buckets[3].Data.VeryLowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[3].Data.LowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[3].Data.TargetRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[3].Data.HighRecords).To(Equal(5))
				Expect(userBGMSummary.Stats.Buckets[3].Data.VeryHighRecords).To(Equal(0))

				By("very high records")
				Expect(userBGMSummary.Stats.Buckets[4].Data.VeryLowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[4].Data.LowRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[4].Data.TargetRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[4].Data.HighRecords).To(Equal(0))
				Expect(userBGMSummary.Stats.Buckets[4].Data.VeryHighRecords).To(Equal(5))
			})
		})

		Context("CalculateSummary", func() {
			var newDatumTime time.Time

			It("Returns correct time in range for stats", func() {
				userBGMSummary = types.Create[types.BGMStats](userId)
				ranges := NewDataRanges()
				dataSetBGMData = NewDataSetBGMDataRanges(deviceId, datumTime, 720, ranges)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				stopPoints := []int{1, 7, 14, 30}
				for _, v := range stopPoints {
					periodKey := strconv.Itoa(v) + "d"

					f := fmt.Sprintf("period %s", periodKey)
					By(f)

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInTargetRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInTargetRecords).To(Equal(24 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInTargetPercent).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInTargetPercent).To(Equal(0.200))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryLowRecords).To(Equal(24 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryLowPercent).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryLowPercent).To(Equal(0.200))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInLowRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInLowRecords).To(Equal(24 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInLowPercent).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInLowPercent).To(Equal(0.200))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInHighRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInHighRecords).To(Equal(24 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInHighPercent).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInHighPercent).To(Equal(0.200))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryHighRecords).To(Equal(24 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasTotalRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].TotalRecords).To(Equal(120 * v))

					Expect(userBGMSummary.Stats.Periods[periodKey].HasAverageDailyRecords).To(BeTrue())
					Expect(*userBGMSummary.Stats.Periods[periodKey].AverageDailyRecords).To(BeNumerically("==", 120))
				}
			})

			It("Returns correct average glucose for stats", func() {
				userBGMSummary = types.Create[types.BGMStats](userId)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()

				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(Equal(inTargetBloodGlucose))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with no rolling", func() {
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, inTargetBloodGlucose)
				userBGMSummary = types.Create[types.BGMStats](userId)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()

				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", inTargetBloodGlucose, 0.001))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with rolling low to high record counts", func() {
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 1, lowBloodGlucose)
				userBGMSummary = types.Create[types.BGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 30)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(1))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(1))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", lowBloodGlucose, 0.001))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}

				// start the actual test
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 720, highBloodGlucose)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", highBloodGlucose, 0.001))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})

			It("Returns correctly calculated summary with rolling high to low record counts", func() {
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 720, lowBloodGlucose)
				userBGMSummary = types.Create[types.BGMStats](userId)
				newDatumTime = datumTime.Add(time.Duration(23) * time.Hour)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", lowBloodGlucose, 0.005))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}

				// start the actual test
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 23, highBloodGlucose)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720))

				for i, period := range periodKeys {
					expectedAverage := ExpectedAverage(periodInts[i]*24, 23, highBloodGlucose, lowBloodGlucose)
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", expectedAverage, 0.005))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})

			It("Returns correctly non-rolling summary with two 30 day windows", func() {
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, datumTime, 24, lowBloodGlucose)
				userBGMSummary = types.Create[types.BGMStats](userId)
				newDatumTime = datumTime.AddDate(0, 0, 31)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(24))

				userBGMSummary.Stats.CalculateSummary()
				Expect(userBGMSummary.Stats.TotalHours).To(Equal(24))

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", lowBloodGlucose, 0.001))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}

				// start the actual test
				dataSetBGMData = NewDataSetBGMDataAvg(deviceId, newDatumTime, 168, highBloodGlucose)

				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(720))

				userBGMSummary.Stats.CalculateSummary()

				Expect(userBGMSummary.Stats.TotalHours).To(Equal(720)) // 30 days

				for _, period := range periodKeys {
					Expect(userBGMSummary.Stats.Periods[period].AverageGlucose.Value).To(BeNumerically("~", highBloodGlucose, 0.001))
					Expect(userBGMSummary.Stats.Periods[period].HasAverageGlucose).To(BeTrue())
				}
			})
		})
	})
})
