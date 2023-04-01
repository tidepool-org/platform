package summary_test

import (
	"fmt"
	userTest "github.com/tidepool-org/platform/user/test"
	"math/rand"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/summary/types"

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
	var userID string
	var datumTime time.Time
	var deviceID string
	var err error
	var dataSetCGMData []*glucose.Glucose
	var dataSetBGMData []*glucose.Glucose

	BeforeEach(func() {
		userID = userTest.RandomID()
		deviceID = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("CreateCGMSummary", func() {
		It("Correctly initializes a cgm summary", func() {
			summary := types.Create[types.CGMStats, *types.CGMStats]("1234")
			Expect(summary).To(Not(BeNil()))
			Expect(summary.Type).To(Equal("cgm"))
		})
	})

	Context("GetDuration", func() {
		var libreDatum *glucose.Glucose
		var otherDatum *glucose.Glucose
		typ := pointer.FromString("cbg")

		It("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
			libreDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)
			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")

			duration := types.GetDuration(libreDatum)
			Expect(duration).To(Equal(15))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID)

			duration := types.GetDuration(otherDatum)
			Expect(duration).To(Equal(5))
		})
	})

	Context("CalculateGMI", func() {
		// input and output examples sourced from https://diabetesjournals.org/care/article/41/11/2275/36593/
		It("Returns correct GMI for medical example 1", func() {
			gmi := types.CalculateGMI(5.55)
			Expect(gmi).To(Equal(5.7))
		})

		It("Returns correct GMI for medical example 2", func() {
			gmi := types.CalculateGMI(6.9375)
			Expect(gmi).To(Equal(6.3))
		})

		It("Returns correct GMI for medical example 3", func() {
			gmi := types.CalculateGMI(8.325)
			Expect(gmi).To(Equal(6.9))
		})

		It("Returns correct GMI for medical example 4", func() {
			gmi := types.CalculateGMI(9.722)
			Expect(gmi).To(Equal(7.5))
		})

		It("Returns correct GMI for medical example 5", func() {
			gmi := types.CalculateGMI(11.11)
			Expect(gmi).To(Equal(8.1))
		})

		It("Returns correct GMI for medical example 6", func() {
			gmi := types.CalculateGMI(12.4875)
			Expect(gmi).To(Equal(8.7))
		})

		It("Returns correct GMI for medical example 7", func() {
			gmi := types.CalculateGMI(13.875)
			Expect(gmi).To(Equal(9.3))
		})

		It("Returns correct GMI for medical example 8", func() {
			gmi := types.CalculateGMI(15.2625)
			Expect(gmi).To(Equal(9.9))
		})

		It("Returns correct GMI for medical example 9", func() {
			gmi := types.CalculateGMI(16.65)
			Expect(gmi).To(Equal(10.5))
		})

		It("Returns correct GMI for medical example 10", func() {
			gmi := types.CalculateGMI(19.425)
			Expect(gmi).To(Equal(11.7))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userCGMSummary *types.Summary[types.CGMStats, *types.CGMStats]
		var userBGMSummary *types.Summary[types.BGMStats, *types.BGMStats]
		var periodKeys = []string{"1d", "7d", "14d", "30d"}
		var periodInts = []int{1, 7, 14, 30}

		Context("CalculateCGMStats", func() {
			It("Returns correct day count when given 2 weeks", func() {
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct day count when given 1 week", func() {
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct day count when given 3 weeks", func() {
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 504, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct record count when given overlapping records", func() {
				var doubledCGMData = make([]*glucose.Glucose, 288*2)

				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)

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
				userCGMSummary = types.Create[types.CGMStats](userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)
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
				userCGMSummary = types.Create[types.CGMStats](userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
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

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, secondDatumTime, 168, secondRequestedAvgGlucose)
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
				userCGMSummary = types.Create[types.CGMStats](userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 144, requestedAvgGlucose)
				err = types.AddData(&userCGMSummary.Stats.Buckets, dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userCGMSummary.Stats.Buckets)).To(Equal(144))

				// TODO move to 0.5 hour to test more cases
				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetCGMData = NewDataSetCGMDataAvg(deviceID, incrementalDatumTime, 1, float64(i))

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
					Expect(userCGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(float64((i - 143) * 12 * 5)))

					averageGlucose := userCGMSummary.Stats.Buckets[i].Data.TotalGlucose / float64(userCGMSummary.Stats.Buckets[i].Data.TotalMinutes)
					Expect(averageGlucose).To(BeNumerically("~", i-143))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose+2)
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
				userCGMSummary = types.Create[types.CGMStats](userID)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetCGMDataOne := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-4*time.Hour), 1, veryLowRange)
				dataSetCGMDataTwo := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-3*time.Hour), 1, lowRange)
				dataSetCGMDataThree := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-2*time.Hour), 1, inRange)
				dataSetCGMDataFour := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-1*time.Hour), 1, highRange)
				dataSetCGMDataFive := NewDataSetCGMDataRanges(deviceID, datumTime, 1, veryHighRange)

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

		Context("CalculateBGMStats", func() {
			It("Returns correct day count when given 2 weeks", func() {
				userBGMSummary = types.Create[types.BGMStats](userID)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(336))
			})

			It("Returns correct day count when given 1 week", func() {
				userBGMSummary = types.Create[types.BGMStats](userID)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))
			})

			It("Returns correct day count when given 3 weeks", func() {
				userBGMSummary = types.Create[types.BGMStats](userID)
				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 504, requestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(504))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := requestedAvgGlucose - 4
				userBGMSummary = types.Create[types.BGMStats](userID)

				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userBGMSummary.Stats.Buckets)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(requestedAvgGlucose * 6))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 10*time.Minute)
					Expect(userBGMSummary.Stats.Buckets[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, secondDatumTime, 168, secondRequestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(528)) // 22 days

				By("check total glucose and dates for second batch")
				newHourlyStatsLen = len(userBGMSummary.Stats.Buckets)
				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetBGMData)/6 // 6 records per hour
				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(secondRequestedAvgGlucose * 6))

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
				userBGMSummary = types.Create[types.BGMStats](userID)

				dataSetBGMData = NewDataSetBGMDataAvg(deviceID, datumTime, 144, requestedAvgGlucose)
				err = types.AddData(&userBGMSummary.Stats.Buckets, dataSetBGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userBGMSummary.Stats.Buckets)).To(Equal(144))

				// TODO move to 0.5 hour to test more cases
				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetBGMData = NewDataSetBGMDataAvg(deviceID, incrementalDatumTime, 1, float64(i))

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
					Expect(averageGlucose).To(Equal(float64(i - 143)))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userBGMSummary = types.Create[types.BGMStats](userID)
				dataSetBGMDataOne := NewDataSetBGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
				dataSetBGMDataTwo := NewDataSetBGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
				dataSetBGMDataThree := NewDataSetBGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose+2)
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

					expectedTotalGlucose = (requestedAvgGlucose + float64(i/24)) * 6
					Expect(userBGMSummary.Stats.Buckets[i].Data.TotalGlucose).To(Equal(expectedTotalGlucose))
				}
			})

			It("Returns correct hourly stats for hours with different Time in Range", func() {
				var lastRecordTime time.Time
				userBGMSummary = types.Create[types.BGMStats](userID)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetBGMDataOne := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-4*time.Hour), 1, veryLowRange)
				dataSetBGMDataTwo := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-3*time.Hour), 1, lowRange)
				dataSetBGMDataThree := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-2*time.Hour), 1, inRange)
				dataSetBGMDataFour := NewDataSetBGMDataRanges(deviceID, datumTime.Add(-1*time.Hour), 1, highRange)
				dataSetBGMDataFive := NewDataSetBGMDataRanges(deviceID, datumTime, 1, veryHighRange)

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

		Context("CalculateCGMSummary", func() {
			It("Returns correct time in range for stats", func() {
				var expectedCGMUse float64
				userCGMSummary = types.Create[types.CGMStats](userID)
				ranges := NewDataRanges()
				dataSetCGMData = NewDataSetCGMDataRanges(deviceID, datumTime, 720, ranges)

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
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
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
				userCGMSummary = types.Create[types.CGMStats](userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, 31), 16, requestedAvgGlucose)

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

		})

		Context("UpdateCGM", func() {
			var newDatumTime time.Time

			It("Returns correctly calculated summary with no rolling", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose)
				userCGMSummary = types.Create[types.CGMStats](userID)
				//userCGMSummary.Dates.OutdatedSince = &datumTime
				expectedGMI := types.CalculateGMI(requestedAvgGlucose)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(720))
				//Expect(userCGMSummary.Dates.OutdatedSince).To(BeNil())

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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userID)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				//userCGMSummary.Dates.OutdatedSince = &datumTime
				expectedGMI := types.CalculateGMI(requestedAvgGlucose + 4)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Stats.TotalHours).To(Equal(1))
				//Expect(userCGMSummary.Dates.OutdatedSince).To(BeNil())

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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 720, requestedAvgGlucose+4)
				//userCGMSummary.Dates.OutdatedSince = &datumTime

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				//Expect(userCGMSummary.Dates.OutdatedSince).To(BeNil())
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

			//It("Returns correctly calculated summary with rolling 100% cgm use", func() {
			//	userData.CGM = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose-4)
			//	userSummary = summary.New(userID, false)
			//	newDatumTime = datumTime.AddDate(0, 0, 7)
			//	userSummary.CGM.OutdatedSince = &datumTime
			//	expectedGMIFirst := summary.CalculateGMI(requestedAvgGlucose - 4)
			//	expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose)
			//
			//	status = summary.UserLastUpdated{
			//		CGM: &summary.UserCGMLastUpdated{
			//			LastData:   datumTime,
			//			LastUpload: datumTime},
			//		BGM: &summary.UserBGMLastUpdated{
			//			LastData:   datumTime,
			//			LastUpload: datumTime},
			//	}
			//
			//	err = userSummary.Update(ctx, &status, &userData)
			//	Expect(err).ToNot(HaveOccurred())
			//	Expect(userSummary.CGM.TotalHours).To(Equal(336))
			//	Expect(userSummary.CGM.OutdatedSince).To(BeNil())
			//
			//	for _, period := range periodKeys {
			//		Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
			//		Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
			//		Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
			//		Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
			//		Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.001))
			//		Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
			//	}
			//
			//	// start the actual test
			//	userData.CGM = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)
			//	userSummary.CGM.OutdatedSince = &datumTime
			//
			//	status = summary.UserLastUpdated{
			//		CGM: &summary.UserCGMLastUpdated{
			//			LastData:   newDatumTime,
			//			LastUpload: newDatumTime},
			//		BGM: &summary.UserBGMLastUpdated{
			//			LastData:   newDatumTime,
			//			LastUpload: newDatumTime},
			//	}
			//
			//	err = userSummary.Update(ctx, &status, &userData)
			//	Expect(err).ToNot(HaveOccurred())
			//	Expect(userSummary.CGM.TotalHours).To(Equal(504))
			//	Expect(userSummary.CGM.OutdatedSince).To(BeNil())
			//
			//	for _, period := range periodKeys {
			//		// TODO make dynamic
			//		//Expect(*userSummary.CGM.Periods[period].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
			//		//Expect(userSummary.CGM.Periods[period].HasTimeCGMUsePercent).To(BeTrue())
			//		//Expect(userSummary.CGM.Periods[period].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
			//		Expect(userSummary.CGM.Periods[period].HasAverageGlucose).To(BeTrue())
			//		//Expect(*userSummary.CGM.Periods[period].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
			//		//Expect(userSummary.CGM.Periods[period].HasGlucoseManagementIndicator).To(BeTrue())
			//	}
			//})

			It("Returns correctly non-rolling summary with two 30 day windows", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userID)
				newDatumTime = datumTime.AddDate(0, 0, 31)
				expectedGMISecond := types.CalculateGMI(requestedAvgGlucose + 4)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 720, requestedAvgGlucose-4)
				userCGMSummary = types.Create[types.CGMStats](userID)
				newDatumTime = datumTime.AddDate(0, 0, 30)
				expectedGMI := types.CalculateGMI(requestedAvgGlucose - 4)

				err = userCGMSummary.Stats.Update(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 1, requestedAvgGlucose+4)

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
