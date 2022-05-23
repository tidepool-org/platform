package summary_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
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
	datum.Time = pointer.FromString(datumTime.Format(time.RFC3339Nano))

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
			Expect(duration).To(Equal(int64(15)))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewContinuous(pointer.FromString(units), &datumTime, &deviceID)

			duration := summary.GetDuration(otherDatum)
			Expect(duration).To(Equal(int64(5)))
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
				secondDatumTime := datumTime.AddDate(0, 0, 14)
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
					Expect(userSummary.DailyStats[6].TotalRecords).To(Equal(int64(12 * i)))
				}

				for i := 0; i < 6; i++ {
					f := fmt.Sprintf("day %d", i)
					By(f)
					Expect(userSummary.DailyStats[i].TotalRecords).To(Equal(int64(288)))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(int64(1440)))

					lastRecordTime := datumTime.Add(-((time.Hour * 24 * time.Duration(5-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userSummary.DailyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 288))
				}

				// check last day
				Expect(userSummary.DailyStats[6].TotalRecords).To(Equal(int64(288)))

				averageGlucose := userSummary.DailyStats[6].TotalGlucose / float64(userSummary.DailyStats[6].TotalRecords)
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
					Expect(userSummary.DailyStats[i].TotalRecords).To(Equal(int64(288)))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(int64(1440)))

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
					Expect(userSummary.DailyStats[i].TotalRecords).To(Equal(int64(285)))
					Expect(userSummary.DailyStats[i].TotalCGMMinutes).To(Equal(int64(1425)))

					lastRecordTime := datumTime.Add(-((time.Hour * 24 * time.Duration(4-i)) + time.Minute*5))
					Expect(userSummary.DailyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("very low minutes")
				Expect(userSummary.DailyStats[0].VeryBelowRangeMinutes).To(Equal(int64(1425)))
				Expect(userSummary.DailyStats[0].BelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].InRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].AboveRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].VeryAboveRangeMinutes).To(Equal(int64(0)))

				By("very low records")
				Expect(userSummary.DailyStats[0].VeryBelowRangeRecords).To(Equal(int64(285)))
				Expect(userSummary.DailyStats[0].BelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].InRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].AboveRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[0].VeryAboveRangeRecords).To(Equal(int64(0)))

				By("low minutes")
				Expect(userSummary.DailyStats[1].VeryBelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].BelowRangeMinutes).To(Equal(int64(1425)))
				Expect(userSummary.DailyStats[1].InRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].AboveRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].VeryAboveRangeMinutes).To(Equal(int64(0)))

				By("low records")
				Expect(userSummary.DailyStats[1].VeryBelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].BelowRangeRecords).To(Equal(int64(285)))
				Expect(userSummary.DailyStats[1].InRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].AboveRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[1].VeryAboveRangeRecords).To(Equal(int64(0)))

				By("in-range minutes")
				Expect(userSummary.DailyStats[2].VeryBelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].BelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].InRangeMinutes).To(Equal(int64(1425)))
				Expect(userSummary.DailyStats[2].AboveRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].VeryAboveRangeMinutes).To(Equal(int64(0)))

				By("in-range records")
				Expect(userSummary.DailyStats[2].VeryBelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].BelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].InRangeRecords).To(Equal(int64(285)))
				Expect(userSummary.DailyStats[2].AboveRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[2].VeryAboveRangeRecords).To(Equal(int64(0)))

				By("high minutes")
				Expect(userSummary.DailyStats[3].VeryBelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].BelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].InRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].AboveRangeMinutes).To(Equal(int64(1425)))
				Expect(userSummary.DailyStats[3].VeryAboveRangeMinutes).To(Equal(int64(0)))

				By("high records")
				Expect(userSummary.DailyStats[3].VeryBelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].BelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].InRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[3].AboveRangeRecords).To(Equal(int64(285)))
				Expect(userSummary.DailyStats[3].VeryAboveRangeRecords).To(Equal(int64(0)))

				By("very high minutes")
				Expect(userSummary.DailyStats[4].VeryBelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].BelowRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].InRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].AboveRangeMinutes).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].VeryAboveRangeMinutes).To(Equal(int64(1425)))

				By("very high records")
				Expect(userSummary.DailyStats[4].VeryBelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].BelowRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].InRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].AboveRangeRecords).To(Equal(int64(0)))
				Expect(userSummary.DailyStats[4].VeryAboveRangeRecords).To(Equal(int64(285)))
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

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.TimeInRange).To(Equal(0.200))
				Expect(*userSummary.TimeVeryBelowRange).To(Equal(0.200))
				Expect(*userSummary.TimeBelowRange).To(Equal(0.200))
				Expect(*userSummary.TimeAboveRange).To(Equal(0.200))
				Expect(*userSummary.TimeVeryAboveRange).To(Equal(0.200))

				// ranges calc only generates 98.9% of a day, count needs to be divisible by 5
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 0.989, 0.001))
			})

			It("Returns correct average glucose for stats", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				err = userSummary.CalculateSummary()
				Expect(err).ToNot(HaveOccurred())

				Expect(*userSummary.AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
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
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMI, 0.001))
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
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 0.5, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(userSummary.GlucoseMgmtIndicator).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 7, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMI, 0.001))
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
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMIFirst, 0.001))

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 7, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
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
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 0.07142, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(userSummary.GlucoseMgmtIndicator).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 14, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
			})
		})
	})
})
