package summary_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/data/summary"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
)

const (
	veryLowBloodGlucose  = 3.0
	lowBloodGlucose      = 3.9
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	units                = "mmol/L"
	requestedAvgGlucose  = 7.0
)

func NewContinuous(units *string, datumTime *time.Time, deviceID *string) *continuous.Continuous {
	datum := continuous.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "cbg"

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

	return datum
}

func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, days float64, reqAvg float64) []*continuous.Continuous {
	requiredRecords := int(days * 288)

	var dataSetData = make([]*continuous.Continuous, requiredRecords)

	// generate X days of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(glucoseValue)

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
func NewDataSetCGMDataRanges(deviceID string, startTime time.Time, days float64, ranges DataRanges) []*continuous.Continuous {
	requiredRecords := int(days * 285)

	var dataSetData = make([]*continuous.Continuous, requiredRecords)

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
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

var _ = Describe("Summary", func() {
	var ctx context.Context
	var logger *logTest.Logger
	var userID string
	var datumTime time.Time
	var deviceID string
	var err error
	var dataSetCGMData []*continuous.Continuous

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		userID = userTest.RandomID()
		deviceID = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("GetDuration", func() {
		var libreDatum *continuous.Continuous
		var otherDatum *continuous.Continuous

		It("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
			libreDatum = NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")

			duration := summary.GetDuration(libreDatum)
			Expect(duration).To(Equal(15))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewContinuous(pointer.FromString(units), &datumTime, &deviceID)

			duration := summary.GetDuration(otherDatum)
			Expect(duration).To(Equal(5))
		})
	})

	Context("CalculateGMI", func() {
		// input and output examples sourced from https://diabetesjournals.org/care/article/41/11/2275/36593/
		It("Returns correct GMI for medical example 1", func() {
			gmi := summary.CalculateGMI(5.55)
			Expect(gmi).To(Equal(5.7))
		})

		It("Returns correct GMI for medical example 2", func() {
			gmi := summary.CalculateGMI(6.9375)
			Expect(gmi).To(Equal(6.3))
		})

		It("Returns correct GMI for medical example 3", func() {
			gmi := summary.CalculateGMI(8.325)
			Expect(gmi).To(Equal(6.9))
		})

		It("Returns correct GMI for medical example 4", func() {
			gmi := summary.CalculateGMI(9.722)
			Expect(gmi).To(Equal(7.5))
		})

		It("Returns correct GMI for medical example 5", func() {
			gmi := summary.CalculateGMI(11.11)
			Expect(gmi).To(Equal(8.1))
		})

		It("Returns correct GMI for medical example 6", func() {
			gmi := summary.CalculateGMI(12.4875)
			Expect(gmi).To(Equal(8.7))
		})

		It("Returns correct GMI for medical example 7", func() {
			gmi := summary.CalculateGMI(13.875)
			Expect(gmi).To(Equal(9.3))
		})

		It("Returns correct GMI for medical example 8", func() {
			gmi := summary.CalculateGMI(15.2625)
			Expect(gmi).To(Equal(9.9))
		})

		It("Returns correct GMI for medical example 9", func() {
			gmi := summary.CalculateGMI(16.65)
			Expect(gmi).To(Equal(10.5))
		})

		It("Returns correct GMI for medical example 10", func() {
			gmi := summary.CalculateGMI(19.425)
			Expect(gmi).To(Equal(11.7))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userSummary *summary.Summary
		Context("CalculateStats", func() {
			It("Returns correct day count when given 2 weeks", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(14))
			})

			It("Returns correct day count when given 1 week", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 7, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(7))
			})

			It("Returns correct day count when given 3 weeks", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 21, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(14))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := requestedAvgGlucose - 4
				userSummary = summary.New(userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 7, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(7))

				By("check total glucose and dates for first batch")
				for i := 0; i < 7; i++ {
					Expect(userSummary.DailyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 288))

					lastRecordTime = datumTime.Add(-((time.Hour * 24 * time.Duration(6-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, secondDatumTime, 7, secondRequestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(7))

				By("check total glucose and dates for second batch")
				for i := 0; i < 7; i++ {
					Expect(userSummary.DailyStats[i].TotalGlucose).To(Equal(secondRequestedAvgGlucose * 288))

					lastRecordTime = secondDatumTime.Add(-((time.Hour * 24 * time.Duration(6-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}
			})

			It("Returns correct stats when given multiple batches in a day", func() {
				var incrementalDatumTime time.Time
				userSummary = summary.New(userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 6, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(6))

				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetCGMData = NewDataSetCGMDataAvg(deviceID, incrementalDatumTime, float64(12)/float64(288), float64(i))

					err = userSummary.CalculateStats(dataSetCGMData)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(userSummary.DailyStats)).To(Equal(7))
					Expect(userSummary.DailyStats[6].TotalCGMRecords).To(Equal(12 * i))
				}

				for i := 0; i < 6; i++ {
					f := fmt.Sprintf("day %d", i)
					By(f)
					Expect(userSummary.DailyStats[i].TotalCGMRecords).To(Equal(288))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(1440))

					lastRecordTime := datumTime.Add(-((time.Hour * 24 * time.Duration(5-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userSummary.DailyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 288))
				}

				// check last day
				Expect(userSummary.DailyStats[6].TotalCGMRecords).To(Equal(288))

				averageGlucose := userSummary.DailyStats[6].TotalGlucose / float64(userSummary.DailyStats[6].TotalCGMRecords)
				Expect(averageGlucose).To(Equal(12.5)) // (1+24)/2
			})

			It("Returns correct daily stats for days with different averages", func() {
				userSummary = summary.New(userID)
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 1, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 1, requestedAvgGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose+2)
				dataSetCGMData = append(dataSetCGMDataOne, dataSetCGMDataTwo...)
				dataSetCGMData = append(dataSetCGMData, dataSetCGMDataThree...)

				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())

				Expect(len(userSummary.DailyStats)).To(Equal(3))

				for i := 0; i < 3; i++ {
					f := fmt.Sprintf("day %d", i)
					By(f)
					Expect(userSummary.DailyStats[i].TotalCGMRecords).To(Equal(288))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(1440))

					lastRecordTime := datumTime.Add(-((time.Hour * 24 * time.Duration(2-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))

					Expect(userSummary.DailyStats[i].TotalGlucose).To(Equal((requestedAvgGlucose + float64(i)) * 288))
				}
			})

			It("Returns correct daily stats for days with different Time in Range", func() {
				userSummary = summary.New(userID)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetCGMDataOne := NewDataSetCGMDataRanges(deviceID, datumTime.AddDate(0, 0, -4), 1, veryLowRange)
				dataSetCGMDataTwo := NewDataSetCGMDataRanges(deviceID, datumTime.AddDate(0, 0, -3), 1, lowRange)
				dataSetCGMDataThree := NewDataSetCGMDataRanges(deviceID, datumTime.AddDate(0, 0, -2), 1, inRange)
				dataSetCGMDataFour := NewDataSetCGMDataRanges(deviceID, datumTime.AddDate(0, 0, -1), 1, highRange)
				dataSetCGMDataFive := NewDataSetCGMDataRanges(deviceID, datumTime.AddDate(0, 0, 0), 1, veryHighRange)

				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
				err = userSummary.CalculateStats(dataSetCGMDataOne)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataTwo)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataThree)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataFour)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataFive)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userSummary.DailyStats)).To(Equal(5))

				By("check record counters for insurance")
				for i := 0; i < 5; i++ {
					f := fmt.Sprintf("day %d", i)
					By(f)
					Expect(userSummary.DailyStats[i].TotalCGMRecords).To(Equal(285))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(1425))

					lastRecordTime := datumTime.Add(-((time.Hour * 24 * time.Duration(4-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("very low minutes")
				Expect(userSummary.DailyStats[0].VeryLowMinutes).To(Equal(1425))
				Expect(userSummary.DailyStats[0].LowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[0].TargetMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[0].HighMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[0].VeryHighMinutes).To(Equal(0))

				By("very low records")
				Expect(userSummary.DailyStats[0].VeryLowRecords).To(Equal(285))
				Expect(userSummary.DailyStats[0].LowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[0].TargetRecords).To(Equal(0))
				Expect(userSummary.DailyStats[0].HighRecords).To(Equal(0))
				Expect(userSummary.DailyStats[0].VeryHighRecords).To(Equal(0))

				By("low minutes")
				Expect(userSummary.DailyStats[1].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[1].LowMinutes).To(Equal(1425))
				Expect(userSummary.DailyStats[1].TargetMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[1].HighMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[1].VeryHighMinutes).To(Equal(0))

				By("low records")
				Expect(userSummary.DailyStats[1].VeryLowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[1].LowRecords).To(Equal(285))
				Expect(userSummary.DailyStats[1].TargetRecords).To(Equal(0))
				Expect(userSummary.DailyStats[1].HighRecords).To(Equal(0))
				Expect(userSummary.DailyStats[1].VeryHighRecords).To(Equal(0))

				By("in-range minutes")
				Expect(userSummary.DailyStats[2].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[2].LowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[2].TargetMinutes).To(Equal(1425))
				Expect(userSummary.DailyStats[2].HighMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[2].VeryHighMinutes).To(Equal(0))

				By("in-range records")
				Expect(userSummary.DailyStats[2].VeryLowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[2].LowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[2].TargetRecords).To(Equal(285))
				Expect(userSummary.DailyStats[2].HighRecords).To(Equal(0))
				Expect(userSummary.DailyStats[2].VeryHighRecords).To(Equal(0))

				By("high minutes")
				Expect(userSummary.DailyStats[3].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[3].LowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[3].TargetMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[3].HighMinutes).To(Equal(1425))
				Expect(userSummary.DailyStats[3].VeryHighMinutes).To(Equal(0))

				By("high records")
				Expect(userSummary.DailyStats[3].VeryLowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[3].LowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[3].TargetRecords).To(Equal(0))
				Expect(userSummary.DailyStats[3].HighRecords).To(Equal(285))
				Expect(userSummary.DailyStats[3].VeryHighRecords).To(Equal(0))

				By("very high minutes")
				Expect(userSummary.DailyStats[4].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[4].LowMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[4].TargetMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[4].HighMinutes).To(Equal(0))
				Expect(userSummary.DailyStats[4].VeryHighMinutes).To(Equal(1425))

				By("very high records")
				Expect(userSummary.DailyStats[4].VeryLowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[4].LowRecords).To(Equal(0))
				Expect(userSummary.DailyStats[4].TargetRecords).To(Equal(0))
				Expect(userSummary.DailyStats[4].HighRecords).To(Equal(0))
				Expect(userSummary.DailyStats[4].VeryHighRecords).To(Equal(285))
			})
		})

		Context("StoreWinningStats", func() {
			It("Stores the right stats with competing devices", func() {
				stats := make(map[string]*summary.Stats)
				userSummary = summary.New(userID)

				stats["worst"] = &summary.Stats{
					DeviceID:        "worse",
					TotalCGMMinutes: 100,
				}
				stats["best"] = &summary.Stats{
					DeviceID:        "best",
					TotalCGMMinutes: 1000,
				}
				stats["worst"] = &summary.Stats{
					DeviceID:        "worst",
					TotalCGMMinutes: 10,
				}

				err = userSummary.StoreWinningStats(stats)

				Expect(userSummary.DailyStats[0].DeviceID).To(Equal("best"))
			})

			// stateful cases are checked as part of the CalculateSummary, as there is already
			// heavy state created for those, and the process is relatively heavy.
		})

		Context("CalculateSummary", func() {
			It("Returns correct time in range for stats", func() {
				userSummary = summary.New(userID)
				ranges := NewDataRanges()
				dataSetCGMData = NewDataSetCGMDataRanges(deviceID, datumTime, 14, ranges)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(14))

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.TotalDays).To(Equal(14))

				Expect(*userSummary.Periods["14d"].TimeInTargetPercent).To(Equal(0.200))
				Expect(*userSummary.Periods["14d"].TimeInTargetMinutes).To(Equal(3990))
				Expect(*userSummary.Periods["14d"].TimeInTargetRecords).To(Equal(798))

				Expect(*userSummary.Periods["14d"].TimeInVeryLowPercent).To(Equal(0.200))
				Expect(*userSummary.Periods["14d"].TimeInVeryLowMinutes).To(Equal(3990))
				Expect(*userSummary.Periods["14d"].TimeInVeryLowRecords).To(Equal(798))

				Expect(*userSummary.Periods["14d"].TimeInLowPercent).To(Equal(0.200))
				Expect(*userSummary.Periods["14d"].TimeInLowMinutes).To(Equal(3990))
				Expect(*userSummary.Periods["14d"].TimeInLowRecords).To(Equal(798))

				Expect(*userSummary.Periods["14d"].TimeInHighPercent).To(Equal(0.200))
				Expect(*userSummary.Periods["14d"].TimeInHighMinutes).To(Equal(3990))
				Expect(*userSummary.Periods["14d"].TimeInHighRecords).To(Equal(798))

				Expect(*userSummary.Periods["14d"].TimeInVeryHighPercent).To(Equal(0.200))
				Expect(*userSummary.Periods["14d"].TimeInVeryHighMinutes).To(Equal(3990))
				Expect(*userSummary.Periods["14d"].TimeInVeryHighRecords).To(Equal(798))

				// ranges calc only generates 98.9% of a day, count needs to be divisible by 5
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.989, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(19950))
				Expect(*userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(3990))
			})

			It("Returns correct average glucose for stats", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(14))

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(Equal(expectedGMI))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(20160))
				Expect(*userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(4032))
			})

			It("Correctly removes GMI when CGM use drop below 0.7", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(14))

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(Equal(expectedGMI))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(20160))
				Expect(*userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(4032))

				// start the real test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, 20), 1, requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.DailyStats)).To(Equal(1))

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.TotalDays).To(Equal(1))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.0714, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(1440))
				Expect(*userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(288))
			})

		})

		Context("Update", func() {
			var userData []*continuous.Continuous
			var status *summary.UserLastUpdated
			var newDatumTime time.Time

			It("Returns correctly calculated summary with no rolling", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				userSummary = summary.New(userID)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling <100% cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 7, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 7)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(7))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.5, 0.001))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 7, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling 100% cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 7)
				userSummary.OutdatedSince = &datumTime
				expectedGMIFirst := summary.CalculateGMI(requestedAvgGlucose - 4)
				expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 7, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly non-rolling summary with two 2 week sets", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 1, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 14)
				userSummary.OutdatedSince = &datumTime
				expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose + 4)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(1))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.07142, 0.001))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 14, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling dropping cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 15)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose - 4)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(14))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 7, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(7))
				Expect(*userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
				Expect(*userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.5, 0.001))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with userData records before summary LastData", func() {
				summaryLastData := datumTime.AddDate(0, 0, -7)
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				userSummary = summary.New(userID)
				userSummary.OutdatedSince = &datumTime
				userSummary.LastData = &summaryLastData

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.TotalDays).To(Equal(7))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})
		})
	})
})
