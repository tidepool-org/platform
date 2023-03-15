package summary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data/summary/types"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/data/types/blood/glucose"

	"github.com/tidepool-org/platform/pointer"
)

const (
	veryLowBloodGlucose  = 3.0
	lowBloodGlucose      = 3.9
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	units                = "mmol/L"
	requestedAvgGlucose  = 7.0
)

func NewGlucose(typ *string, units *string, datumTime *time.Time, deviceID *string) *glucose.Glucose {
	datum := glucose.New(*typ)
	datum.Units = units

	datum.Active = true
	datum.ArchivedDataSetID = nil
	datum.ArchivedTime = nil
	datum.CreatedTime = nil
	datum.CreatedUserID = nil
	datum.DeletedTime = nil
	datum.DeletedUserID = nil
	datum.DeviceID = deviceID
	datum.ModifiedTime = nil
	datum.ModifiedUserID = nil
	datum.Time = datumTime

	return &datum
}

func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, hours float64, reqAvg float64) []*glucose.Glucose {
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

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
			datum.Value = pointer.FromFloat64(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

func NewDataSetBGMDataAvg(deviceID string, startTime time.Time, hours float64, reqAvg float64) []*glucose.Glucose {
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

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
			datum.Value = pointer.FromFloat64(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

type DataRanges struct {
	Min      float64
	Max      float64
	Padding  float64
	VeryLow  float64
	Low      float64
	High     float64
	VeryHigh float64
}

func NewDataRanges() DataRanges {
	return DataRanges{
		Min:      1,
		Max:      20,
		Padding:  0.01,
		VeryLow:  veryLowBloodGlucose,
		Low:      lowBloodGlucose,
		High:     highBloodGlucose,
		VeryHigh: veryHighBloodGlucose,
	}
}

func NewDataRangesSingle(value float64) DataRanges {
	return DataRanges{
		Min:      value,
		Max:      value,
		Padding:  0,
		VeryLow:  value,
		Low:      value,
		High:     value,
		VeryHigh: value,
	}
}

// creates a dataset with random values evenly divided between ranges
// NOTE: only generates 98.9% CGMUse, due to needing to be divisible by 5
func NewDataSetCGMDataRanges(deviceID string, startTime time.Time, hours float64, ranges DataRanges) []*glucose.Glucose {
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

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
			datum.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

func NewDataSetBGMDataRanges(deviceID string, startTime time.Time, hours float64, ranges DataRanges) []*glucose.Glucose {
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

			datum := NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
			datum.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

var _ = Describe("Summary", func() {
	//	var ctx context.Context
	//	var logger *logTest.Logger
	//	var userID string
	//	var datumTime time.Time
	//	var deviceID string
	//	var err error
	//	var dataSetCGMData []*glucose.Glucose
	//	var dataSetBGMData []*glucose.Glucose
	//
	//	BeforeEach(func() {
	//		logger = logTest.NewLogger()
	//		ctx = log.NewContextWithLogger(context.Background(), logger)
	//		userID = userTest.RandomID()
	//		deviceID = "SummaryTestDevice"
	//		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	//	})

	Context("CreateCGMSummary", func() {
		It("Correctly initializes a cgm summary", func() {
			summary := types.Create[types.CGMStats, *types.CGMStats]("1234")
			Expect(summary).To(Not(BeNil()))
			Expect(summary.Type).To(Equal("cgm"))
		})
	})
})

//
//	Context("GetDuration", func() {
//		var libreDatum *glucose.Glucose
//		var otherDatum *glucose.Glucose
//		typ := pointer.FromString("cbg")
//
//		It("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
//			libreDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
//			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")
//
//			duration := summary.GetDuration(libreDatum)
//			Expect(duration).To(Equal(15))
//		})
//
//		It("Returns correct duration for other devices", func() {
//			otherDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
//
//			duration := summary.GetDuration(otherDatum)
//			Expect(duration).To(Equal(5))
//		})
//	})
//
//	Context("CalculateGMI", func() {
//		// input and output examples sourced from https://diabetesjournals.org/care/article/41/11/2275/36593/
//		It("Returns correct GMI for medical example 1", func() {
//			gmi := summary.CalculateGMI(5.55)
//			Expect(gmi).To(Equal(5.7))
//		})
//
//		It("Returns correct GMI for medical example 2", func() {
//			gmi := summary.CalculateGMI(6.9375)
//			Expect(gmi).To(Equal(6.3))
//		})
//
//		It("Returns correct GMI for medical example 3", func() {
//			gmi := summary.CalculateGMI(8.325)
//			Expect(gmi).To(Equal(6.9))
//		})
//
//		It("Returns correct GMI for medical example 4", func() {
//			gmi := summary.CalculateGMI(9.722)
//			Expect(gmi).To(Equal(7.5))
//		})
//
//		It("Returns correct GMI for medical example 5", func() {
//			gmi := summary.CalculateGMI(11.11)
//			Expect(gmi).To(Equal(8.1))
//		})
//
//		It("Returns correct GMI for medical example 6", func() {
//			gmi := summary.CalculateGMI(12.4875)
//			Expect(gmi).To(Equal(8.7))
//		})
//
//		It("Returns correct GMI for medical example 7", func() {
//			gmi := summary.CalculateGMI(13.875)
//			Expect(gmi).To(Equal(9.3))
//		})
//
//		It("Returns correct GMI for medical example 8", func() {
//			gmi := summary.CalculateGMI(15.2625)
//			Expect(gmi).To(Equal(9.9))
//		})
//
//		It("Returns correct GMI for medical example 9", func() {
//			gmi := summary.CalculateGMI(16.65)
//			Expect(gmi).To(Equal(10.5))
//		})
//
//		It("Returns correct GMI for medical example 10", func() {
//			gmi := summary.CalculateGMI(19.425)
//			Expect(gmi).To(Equal(11.7))
//		})
//	})
//
//	Context("Summary calculations requiring datasets", func() {
//		var userSummary *summary.Summary
//		var periodKeys = []string{"1d", "7d", "14d", "30d"}
//		var periodInts = []int{1, 7, 14, 30}
//
//		Context("CalculateCGMStats", func() {
//			It("Returns correct day count when given 2 weeks", func() {
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(336))
//			})
//
//			It("Returns correct day count when given 1 week", func() {
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(168))
//			})
//
//			It("Returns correct day count when given 3 weeks", func() {
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 504, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(504))
//			})
//
//			It("Returns correct record count when given overlapping records", func() {
//				var doubledCGMData = make([]*glucose.Glucose, 288*2)
//
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
//				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)
//
//				// interlace the lists
//				for i := 0; i < len(dataSetCGMData); i += 1 {
//					doubledCGMData[i*2] = dataSetCGMData[i]
//					doubledCGMData[i*2+1] = dataSetCGMDataTwo[i]
//				}
//				err = userSummary.CalculateCGMStats(doubledCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(24))
//				Expect(userSummary.CGM.HourlyStats[0].TotalRecords).To(Equal(12))
//			})
//
//			It("Returns correct record count when given overlapping records across multiple calculations", func() {
//				userSummary = summary.New(userID, false)
//
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(24))
//				Expect(userSummary.CGM.HourlyStats[0].TotalRecords).To(Equal(12))
//			})
//
//			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
//				var lastRecordTime time.Time
//				var hourlyStatsLen int
//				var newHourlyStatsLen int
//				secondDatumTime := datumTime.AddDate(0, 0, 15)
//				secondRequestedAvgGlucose := requestedAvgGlucose - 4
//				userSummary = summary.New(userID, false)
//
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(168))
//
//				By("check total glucose and dates for first batch")
//				hourlyStatsLen = len(userSummary.CGM.HourlyStats)
//				for i := hourlyStatsLen - 1; i >= 0; i-- {
//					Expect(userSummary.CGM.HourlyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 12))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 5*time.Minute)
//					Expect(userSummary.CGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, secondDatumTime, 168, secondRequestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(528)) // 22 days
//
//				By("check total glucose and dates for second batch")
//				newHourlyStatsLen = len(userSummary.CGM.HourlyStats)
//				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetCGMData)/12 // 12 per day, need length without the gap
//				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
//					Expect(userSummary.CGM.HourlyStats[i].TotalGlucose).To(Equal(secondRequestedAvgGlucose * 12))
//
//					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 5*time.Minute)
//					Expect(userSummary.CGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("check total glucose and dates for gap")
//				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
//				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
//					Expect(userSummary.CGM.HourlyStats[i].TotalGlucose).To(Equal(float64(0)))
//				}
//			})
//
//			It("Returns correct stats when given multiple batches in a day", func() {
//				var incrementalDatumTime time.Time
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 144, requestedAvgGlucose)
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(144))
//
//				// TODO move to 0.5 hour to test more cases
//				for i := 1; i <= 24; i++ {
//					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
//					dataSetCGMData = NewDataSetCGMDataAvg(deviceID, incrementalDatumTime, 1, float64(i))
//
//					err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//					Expect(err).ToNot(HaveOccurred())
//					Expect(len(userSummary.CGM.HourlyStats)).To(Equal(144 + i))
//					Expect(userSummary.CGM.HourlyStats[i].TotalRecords).To(Equal(12))
//				}
//
//				for i := 144; i < len(userSummary.CGM.HourlyStats); i++ {
//					f := fmt.Sprintf("hour %d", i)
//					By(f)
//					Expect(userSummary.CGM.HourlyStats[i].TotalRecords).To(Equal(12))
//					Expect(userSummary.CGM.HourlyStats[i].TotalMinutes).To(Equal(60))
//
//					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*5)
//					Expect(userSummary.CGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//					Expect(userSummary.CGM.HourlyStats[i].TotalGlucose).To(Equal(float64((i - 143) * 12)))
//
//					averageGlucose := userSummary.CGM.HourlyStats[i].TotalGlucose / float64(userSummary.CGM.HourlyStats[i].TotalRecords)
//					Expect(averageGlucose).To(Equal(float64(i - 143)))
//				}
//			})
//
//			It("Returns correct daily stats for days with different averages", func() {
//				var expectedTotalGlucose float64
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
//				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
//				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose+2)
//				dataSetCGMData = append(dataSetCGMDataOne, dataSetCGMDataTwo...)
//				dataSetCGMData = append(dataSetCGMData, dataSetCGMDataThree...)
//
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(72))
//
//				for i := len(userSummary.CGM.HourlyStats) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userSummary.CGM.HourlyStats[i].TotalRecords).To(Equal(12))
//					Expect(userSummary.CGM.HourlyStats[i].TotalMinutes).To(Equal(60))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.CGM.HourlyStats)-i-1) - 5*time.Minute)
//					Expect(userSummary.CGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//
//					expectedTotalGlucose = (requestedAvgGlucose + float64(i/24)) * 12
//					Expect(userSummary.CGM.HourlyStats[i].TotalGlucose).To(Equal(expectedTotalGlucose))
//				}
//			})
//
//			It("Returns correct hourly stats for hours with different Time in Range", func() {
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
//				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
//				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
//				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
//				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)
//
//				dataSetCGMDataOne := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-4*time.Hour), 1, veryLowRange)
//				dataSetCGMDataTwo := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-3*time.Hour), 1, lowRange)
//				dataSetCGMDataThree := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-2*time.Hour), 1, inRange)
//				dataSetCGMDataFour := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-1*time.Hour), 1, highRange)
//				dataSetCGMDataFive := NewDataSetCGMDataRanges(deviceID, datumTime, 1, veryHighRange)
//
//				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
//				err = userSummary.CalculateCGMStats(dataSetCGMDataOne)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateCGMStats(dataSetCGMDataTwo)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateCGMStats(dataSetCGMDataThree)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateCGMStats(dataSetCGMDataFour)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateCGMStats(dataSetCGMDataFive)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(5))
//
//				By("check record counters for insurance")
//				for i := len(userSummary.CGM.HourlyStats) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userSummary.CGM.HourlyStats[i].TotalRecords).To(Equal(10))
//					Expect(userSummary.CGM.HourlyStats[i].TotalMinutes).To(Equal(50))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.CGM.HourlyStats)-i-1) - time.Minute*5)
//					Expect(userSummary.CGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("very low minutes")
//				Expect(userSummary.CGM.HourlyStats[0].VeryLowMinutes).To(Equal(50))
//				Expect(userSummary.CGM.HourlyStats[0].LowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].TargetMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].HighMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].VeryHighMinutes).To(Equal(0))
//
//				By("very low records")
//				Expect(userSummary.CGM.HourlyStats[0].VeryLowRecords).To(Equal(10))
//				Expect(userSummary.CGM.HourlyStats[0].LowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].TargetRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].HighRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[0].VeryHighRecords).To(Equal(0))
//
//				By("low minutes")
//				Expect(userSummary.CGM.HourlyStats[1].VeryLowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].LowMinutes).To(Equal(50))
//				Expect(userSummary.CGM.HourlyStats[1].TargetMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].HighMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].VeryHighMinutes).To(Equal(0))
//
//				By("low records")
//				Expect(userSummary.CGM.HourlyStats[1].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].LowRecords).To(Equal(10))
//				Expect(userSummary.CGM.HourlyStats[1].TargetRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].HighRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[1].VeryHighRecords).To(Equal(0))
//
//				By("in-range minutes")
//				Expect(userSummary.CGM.HourlyStats[2].VeryLowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].LowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].TargetMinutes).To(Equal(50))
//				Expect(userSummary.CGM.HourlyStats[2].HighMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].VeryHighMinutes).To(Equal(0))
//
//				By("in-range records")
//				Expect(userSummary.CGM.HourlyStats[2].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].LowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].TargetRecords).To(Equal(10))
//				Expect(userSummary.CGM.HourlyStats[2].HighRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[2].VeryHighRecords).To(Equal(0))
//
//				By("high minutes")
//				Expect(userSummary.CGM.HourlyStats[3].VeryLowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].LowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].TargetMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].HighMinutes).To(Equal(50))
//				Expect(userSummary.CGM.HourlyStats[3].VeryHighMinutes).To(Equal(0))
//
//				By("high records")
//				Expect(userSummary.CGM.HourlyStats[3].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].LowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].TargetRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[3].HighRecords).To(Equal(10))
//				Expect(userSummary.CGM.HourlyStats[3].VeryHighRecords).To(Equal(0))
//
//				By("very high minutes")
//				Expect(userSummary.CGM.HourlyStats[4].VeryLowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].LowMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].TargetMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].HighMinutes).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].VeryHighMinutes).To(Equal(50))
//
//				By("very high records")
//				Expect(userSummary.CGM.HourlyStats[4].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].LowRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].TargetRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].HighRecords).To(Equal(0))
//				Expect(userSummary.CGM.HourlyStats[4].VeryHighRecords).To(Equal(10))
//			})
//		})
//
//		Context("CalculateBGMStats", func() {
//			It("Returns correct day count when given 2 weeks", func() {
//				userSummary = summary.New(userID, false)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(336))
//			})
//
//			It("Returns correct day count when given 1 week", func() {
//				userSummary = summary.New(userID, false)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(168))
//			})
//
//			It("Returns correct day count when given 3 weeks", func() {
//				userSummary = summary.New(userID, false)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 504, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(504))
//			})
//
//			It("Returns correct record count when given overlapping records", func() {
//				var doubledBGMData = make([]*glucose.Glucose, 6*2)
//
//				userSummary = summary.New(userID, false)
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose)
//				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceID, datumTime.Add(15*time.Second), 1, requestedAvgGlucose)
//
//				// interlace the lists
//				for i := 0; i < len(dataSetBGMData); i += 1 {
//					doubledBGMData[i*2] = dataSetBGMData[i]
//					doubledBGMData[i*2+1] = dataSetBGMDataTwo[i]
//				}
//				err = userSummary.CalculateBGMStats(doubledBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(1))
//				Expect(userSummary.BGM.HourlyStats[0].TotalRecords).To(Equal(12))
//			})
//
//			It("Returns correct record count when given overlapping records across multiple calculations", func() {
//				userSummary = summary.New(userID, false)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime.Add(15*time.Second), 1, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(1))
//				Expect(userSummary.BGM.HourlyStats[0].TotalRecords).To(Equal(12))
//			})
//
//			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
//				var lastRecordTime time.Time
//				var hourlyStatsLen int
//				var newHourlyStatsLen int
//				secondDatumTime := datumTime.AddDate(0, 0, 15)
//				secondRequestedAvgGlucose := requestedAvgGlucose - 4
//				userSummary = summary.New(userID, false)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(168))
//
//				By("check total glucose and dates for first batch")
//				hourlyStatsLen = len(userSummary.BGM.HourlyStats)
//				for i := hourlyStatsLen - 1; i >= 0; i-- {
//					Expect(userSummary.BGM.HourlyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 6))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 10*time.Minute)
//					Expect(userSummary.BGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, secondDatumTime, 168, secondRequestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(528)) // 22 days
//
//				By("check total glucose and dates for second batch")
//				newHourlyStatsLen = len(userSummary.BGM.HourlyStats)
//				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetBGMData)/6 // 6 records per hour
//				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
//					Expect(userSummary.BGM.HourlyStats[i].TotalGlucose).To(Equal(secondRequestedAvgGlucose * 6))
//
//					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 10*time.Minute)
//					Expect(userSummary.BGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("check total glucose and dates for gap")
//				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
//				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
//					Expect(userSummary.BGM.HourlyStats[i].TotalGlucose).To(Equal(float64(0)))
//				}
//			})
//
//			It("Returns correct stats when given multiple batches in a day", func() {
//				var incrementalDatumTime time.Time
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//
//				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 144, requestedAvgGlucose)
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(144))
//
//				// TODO move to 0.5 hour to test more cases
//				for i := 1; i <= 24; i++ {
//					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
//					dataSetBGMData = NewDataSetBGMDataAvg(deviceID, incrementalDatumTime, 1, float64(i))
//
//					err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//					Expect(err).ToNot(HaveOccurred())
//					Expect(len(userSummary.BGM.HourlyStats)).To(Equal(144 + i))
//					Expect(userSummary.BGM.HourlyStats[i].TotalRecords).To(Equal(6))
//				}
//
//				for i := 144; i < len(userSummary.BGM.HourlyStats); i++ {
//					f := fmt.Sprintf("hour %d", i)
//					By(f)
//					Expect(userSummary.BGM.HourlyStats[i].TotalRecords).To(Equal(6))
//
//					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*10)
//					Expect(userSummary.BGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//					Expect(userSummary.BGM.HourlyStats[i].TotalGlucose).To(Equal(float64((i - 143) * 6)))
//
//					averageGlucose := userSummary.BGM.HourlyStats[i].TotalGlucose / float64(userSummary.BGM.HourlyStats[i].TotalRecords)
//					Expect(averageGlucose).To(Equal(float64(i - 143)))
//				}
//			})
//
//			It("Returns correct daily stats for days with different averages", func() {
//				var expectedTotalGlucose float64
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//				dataSetBGMDataOne := NewDataSetBGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
//				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
//				dataSetBGMDataThree := NewDataSetBGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose+2)
//				dataSetBGMData = append(dataSetBGMDataOne, dataSetBGMDataTwo...)
//				dataSetBGMData = append(dataSetBGMData, dataSetBGMDataThree...)
//
//				err = userSummary.CalculateBGMStats(dataSetBGMData)
//
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(72))
//
//				for i := len(userSummary.BGM.HourlyStats) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userSummary.BGM.HourlyStats[i].TotalRecords).To(Equal(6))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.BGM.HourlyStats)-i-1) - 10*time.Minute)
//					Expect(userSummary.BGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//
//					expectedTotalGlucose = (requestedAvgGlucose + float64(i/24)) * 6
//					Expect(userSummary.BGM.HourlyStats[i].TotalGlucose).To(Equal(expectedTotalGlucose))
//				}
//			})
//
//			It("Returns correct hourly stats for hours with different Time in Range", func() {
//				var lastRecordTime time.Time
//				userSummary = summary.New(userID, false)
//				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
//				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
//				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
//				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
//				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)
//
//				dataSetBGMDataOne := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-4*time.Hour), 1, veryLowRange)
//				dataSetBGMDataTwo := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-3*time.Hour), 1, lowRange)
//				dataSetBGMDataThree := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-2*time.Hour), 1, inRange)
//				dataSetBGMDataFour := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-1*time.Hour), 1, highRange)
//				dataSetBGMDataFive := NewDataSetBGMDataRanges(deviceID, datumTime, 1, veryHighRange)
//
//				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
//				err = userSummary.CalculateBGMStats(dataSetBGMDataOne)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateBGMStats(dataSetBGMDataTwo)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateBGMStats(dataSetBGMDataThree)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateBGMStats(dataSetBGMDataFour)
//				Expect(err).ToNot(HaveOccurred())
//				err = userSummary.CalculateBGMStats(dataSetBGMDataFive)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(len(userSummary.BGM.HourlyStats)).To(Equal(5))
//
//				By("check record counters for insurance")
//				for i := len(userSummary.BGM.HourlyStats) - 1; i >= 0; i-- {
//					f := fmt.Sprintf("hour %d", i+1)
//					By(f)
//					Expect(userSummary.BGM.HourlyStats[i].TotalRecords).To(Equal(5))
//
//					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.BGM.HourlyStats)-i-1) - time.Minute*12)
//					Expect(userSummary.BGM.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
//				}
//
//				By("very low records")
//				Expect(userSummary.BGM.HourlyStats[0].VeryLowRecords).To(Equal(5))
//				Expect(userSummary.BGM.HourlyStats[0].LowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[0].TargetRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[0].HighRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[0].VeryHighRecords).To(Equal(0))
//
//				By("low records")
//				Expect(userSummary.BGM.HourlyStats[1].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[1].LowRecords).To(Equal(5))
//				Expect(userSummary.BGM.HourlyStats[1].TargetRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[1].HighRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[1].VeryHighRecords).To(Equal(0))
//
//				By("in-range records")
//				Expect(userSummary.BGM.HourlyStats[2].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[2].LowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[2].TargetRecords).To(Equal(5))
//				Expect(userSummary.BGM.HourlyStats[2].HighRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[2].VeryHighRecords).To(Equal(0))
//
//				By("high records")
//				Expect(userSummary.BGM.HourlyStats[3].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[3].LowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[3].TargetRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[3].HighRecords).To(Equal(5))
//				Expect(userSummary.BGM.HourlyStats[3].VeryHighRecords).To(Equal(0))
//
//				By("very high records")
//				Expect(userSummary.BGM.HourlyStats[4].VeryLowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[4].LowRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[4].TargetRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[4].HighRecords).To(Equal(0))
//				Expect(userSummary.BGM.HourlyStats[4].VeryHighRecords).To(Equal(5))
//			})
//		})
//
//		Context("CalculateSummary", func() {
//			It("Returns correct time in range for stats", func() {
//				var expectedCGMUse float64
//				userSummary = summary.New(userID, false)
//				ranges := NewDataRanges()
//				dataSetCGMData = NewDataSetCGMDataRanges(deviceID, datumTime, 720, ranges)
//
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(720))
//
//				userSummary.CalculateCGMSummary()
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//
//				stopPoints := []int{1, 7, 14, 30}
//				for _, v := range stopPoints {
//					periodKey := strconv.Itoa(v) + "d"
//
//					f := fmt.Sprintf("period %s", periodKey)
//					By(f)
//
//					Expect(userSummary.CGM.Periods[periodKey].TimeInTargetMinutes).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeInTargetRecords).To(Equal(48 * v))
//					Expect(*userSummary.CGM.Periods[periodKey].TimeInTargetPercent).To(Equal(0.200))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeInTargetPercent).To(BeTrue())
//
//					Expect(userSummary.CGM.Periods[periodKey].TimeInVeryLowMinutes).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeInVeryLowRecords).To(Equal(48 * v))
//					Expect(*userSummary.CGM.Periods[periodKey].TimeInVeryLowPercent).To(Equal(0.200))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeInVeryLowPercent).To(BeTrue())
//
//					Expect(userSummary.CGM.Periods[periodKey].TimeInLowMinutes).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeInLowRecords).To(Equal(48 * v))
//					Expect(*userSummary.CGM.Periods[periodKey].TimeInLowPercent).To(Equal(0.200))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeInLowPercent).To(BeTrue())
//
//					Expect(userSummary.CGM.Periods[periodKey].TimeInHighMinutes).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeInHighRecords).To(Equal(48 * v))
//					Expect(*userSummary.CGM.Periods[periodKey].TimeInHighPercent).To(Equal(0.200))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeInHighPercent).To(BeTrue())
//
//					Expect(userSummary.CGM.Periods[periodKey].TimeInVeryHighMinutes).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))
//					Expect(*userSummary.CGM.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeInVeryHighPercent).To(BeTrue())
//
//					// ranges calc only generates 83.3% of an hour, each hour needs to be divisible by 5
//					Expect(userSummary.CGM.Periods[periodKey].TimeCGMUseMinutes).To(Equal(1200 * v))
//					Expect(userSummary.CGM.Periods[periodKey].TimeCGMUseRecords).To(Equal(240 * v))
//					Expect(userSummary.CGM.Periods[periodKey].HasTimeCGMUsePercent).To(BeTrue())
//
//					// this value is a bit funny, its 83.3%, but the missing end of the final day gets compensated off
//					// resulting in 83.6% only on the first day
//					if v == 1 {
//						expectedCGMUse = 0.836
//					} else {
//						expectedCGMUse = 0.833
//					}
//
//					Expect(*userSummary.CGM.Periods[periodKey].TimeCGMUsePercent).To(BeNumerically("~", expectedCGMUse, 0.001))
//				}
//			})
//
//			It("Returns correct average glucose for stats", func() {
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
//				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)
//
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(720))
//
//				userSummary.CalculateCGMSummary()
//
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//				}
//			})
//
//			It("Correctly removes GMI when CGM use drop below 0.7", func() {
//				userSummary = summary.New(userID, false)
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
//				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)
//
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(720))
//
//				userSummary.CalculateCGMSummary()
//
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(Equal(expectedGMI))
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//				}
//
//				// start the real test
//				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, 31), 16, requestedAvgGlucose)
//
//				err = userSummary.CalculateCGMStats(dataSetCGMData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(len(userSummary.CGM.HourlyStats)).To(Equal(720)) // hits 4 days over 30 day cap
//
//				userSummary.CalculateCGMSummary()
//
//				Expect(userSummary.CGM.TotalHours).To(Equal(30 * 24)) // 30 days currently capped
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(
//						BeNumerically("~", 960/(float64(periodInts[i])*1440), 0.005))
//					Expect(userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNil())
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
//
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(192))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(960))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//				}
//			})
//
//		})
//
//		Context("Update", func() {
//			var userData summary.UserData
//			var status summary.UserLastUpdated
//			var newDatumTime time.Time
//
//			It("Returns correctly calculated summary with no rolling", func() {
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
//				userSummary = summary.New(userID, false)
//				userSummary.CGM.OutdatedSince = &datumTime
//				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//				}
//			})
//
//			It("Returns correctly calculated summary with rolling <100% cgm use", func() {
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose-4)
//				userSummary = summary.New(userID, false)
//				newDatumTime = datumTime.AddDate(0, 0, 30)
//				userSummary.CGM.OutdatedSince = &datumTime
//				expectedGMI := summary.CalculateGMI(requestedAvgGlucose + 4)
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.TotalHours).To(Equal(1))
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(
//						BeNumerically("~", 60/(float64(periodInts[i])*1440), 0.006))
//
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(12))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(60))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNil())
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
//				}
//
//				// start the actual test
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, newDatumTime, 720, requestedAvgGlucose+4)
//				userSummary.CGM.OutdatedSince = &datumTime
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//				}
//			})
//
//			//It("Returns correctly calculated summary with rolling 100% cgm use", func() {
//			//	userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose-4)
//			//	userSummary = summary.New(userID, false)
//			//	newDatumTime = datumTime.AddDate(0, 0, 7)
//			//	userSummary.CGM.OutdatedSince = &datumTime
//			//	expectedGMIFirst := summary.CalculateGMI(requestedAvgGlucose - 4)
//			//	expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose)
//			//
//			//	status = summary.UserLastUpdated{
//			//		CGM: &summary.UserCGMLastUpdated{
//			//			LastData:   datumTime,
//			//			LastUpload: datumTime},
//			//		BGM: &summary.UserBGMLastUpdated{
//			//			LastData:   datumTime,
//			//			LastUpload: datumTime},
//			//	}
//			//
//			//	err = userSummary.Update(ctx, &status, &userData)
//			//	Expect(err).ToNot(HaveOccurred())
//			//	Expect(userSummary.CGM.TotalHours).To(Equal(336))
//			//	Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//			//
//			//	for _, period := range periodKeys {
//			//		Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
//			//		Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//			//		Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
//			//		Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//			//		Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.001))
//			//		Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//			//	}
//			//
//			//	// start the actual test
//			//	userData.CGM = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)
//			//	userSummary.CGM.OutdatedSince = &datumTime
//			//
//			//	status = summary.UserLastUpdated{
//			//		CGM: &summary.UserCGMLastUpdated{
//			//			LastData:   newDatumTime,
//			//			LastUpload: newDatumTime},
//			//		BGM: &summary.UserBGMLastUpdated{
//			//			LastData:   newDatumTime,
//			//			LastUpload: newDatumTime},
//			//	}
//			//
//			//	err = userSummary.Update(ctx, &status, &userData)
//			//	Expect(err).ToNot(HaveOccurred())
//			//	Expect(userSummary.CGM.TotalHours).To(Equal(504))
//			//	Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//			//
//			//	for _, period := range periodKeys {
//			//		// TODO make dynamic
//			//		//Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
//			//		//Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//			//		//Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
//			//		Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//			//		//Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
//			//		//Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//			//	}
//			//})
//
//			It("Returns correctly non-rolling summary with two 30 day windows", func() {
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose-4)
//				userSummary = summary.New(userID, false)
//				newDatumTime = datumTime.AddDate(0, 0, 31)
//				userSummary.CGM.OutdatedSince = &datumTime
//				expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose + 4)
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.TotalHours).To(Equal(24))
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440/(1440*float64(periodInts[i])), 0.005))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(288))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(1440))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					if *userSummary.CGM.Periods[period].TimeCGMUsePercent > 0.7 {
//						Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//					} else {
//						Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
//						Expect(userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNil())
//					}
//				}
//
//				// start the actual test
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)
//				userSummary.CGM.OutdatedSince = &datumTime
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(userSummary.CGM.TotalHours).To(Equal(720)) // 30 days
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for i, period := range periodKeys {
//					if i == 0 || i == 1 {
//						Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(288 * periodInts[i]))
//						Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(1440 * periodInts[i]))
//						Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					} else {
//						Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(7 * 288))
//						Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(7 * 1440))
//						Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1440*7/(1440*float64(periodInts[i])), 0.005))
//					}
//
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					if *userSummary.CGM.Periods[period].TimeCGMUsePercent > 0.7 {
//						Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
//						Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//					} else {
//						Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
//						Expect(userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNil())
//					}
//				}
//			})
//
//			It("Returns correctly calculated summary with rolling dropping cgm use", func() {
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose-4)
//				userSummary = summary.New(userID, false)
//				newDatumTime = datumTime.AddDate(0, 0, 30)
//				userSummary.CGM.OutdatedSince = &datumTime
//				expectedGMI := summary.CalculateGMI(requestedAvgGlucose - 4)
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.TotalHours).To(Equal(720))
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for i, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.005))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(periodInts[i] * 288))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(periodInts[i] * 1440))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
//				}
//
//				// start the actual test
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, newDatumTime, 1, requestedAvgGlucose+4)
//				userSummary.CGM.OutdatedSince = &datumTime
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   newDatumTime,
//						LastUpload: newDatumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//
//				Expect(userSummary.CGM.TotalHours).To(Equal(720)) // 30 days
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//
//				for _, period := range periodKeys {
//					Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 0.03, 0.03))
//					Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseRecords).To(Equal(12))
//					Expect(userSummary.CGM.Periods[period].TimeCGMUseMinutes).To(Equal(60))
//					Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.05))
//					Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
//					Expect(userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNil())
//					Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeFalse())
//				}
//			})
//
//			It("Returns correctly calculated summary with userData records before summary LastData", func() {
//				summaryLastData := datumTime.AddDate(0, 0, -7)
//				userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
//				userSummary = summary.New(userID, false)
//				userSummary.CGM.OutdatedSince = &datumTime
//				userSummary.CGM.LastData = &summaryLastData
//
//				status = summary.UserLastUpdated{
//					CGM: &summary.UserCGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//					BGM: &summary.UserBGMLastUpdated{
//						LastData:   datumTime,
//						LastUpload: datumTime},
//				}
//
//				err = userSummary.Update(ctx, &status, &userData)
//				Expect(err).ToNot(HaveOccurred())
//				Expect(userSummary.CGM.TotalHours).To(Equal(168))
//				Expect(userSummary.CGM.OutdatedSince).To(BeNil())
//			})
//		})
//	})
//})
