package summary_test

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
)

const (
	veryLowBloodGlucose  = 3.0
	lowBloodGlucose      = 3.9
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	units                = "mmol/l"
	requestedAvgGlucose  = 5.0
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

func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, days int, reqAvg float64) []*continuous.Continuous {
	var dataSetData []*continuous.Continuous

	requiredRecords := days * 288

	// generate X days of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(glucoseValue)
			dataSetData = append(dataSetData, datum)
		}
	}

	return dataSetData
}

// creates a dataset with random values evenly divided between ranges
func NewDataSetCGMDataTimeInRange(deviceID string, startTime time.Time, days int, veryLow float64, low float64, high float64, veryHigh float64) []*continuous.Continuous {
	var dataSetData []*continuous.Continuous

	requiredRecords := days * 288

	glucoseBrackets := [5][2]float64{
		{1, veryLow - 0.01},
		{veryLow, low - 0.01},
		{low, high - 0.01},
		{high, veryHigh - 0.01},
		{veryHigh, 20},
	}

	// generate 2 weeks of data
	for count := 0; count < requiredRecords; count += 5 {
		for i, bracket := range glucoseBrackets {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData = append(dataSetData, datum)
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
	var highDeviceID string
	var lowDeviceID string
	var dataSetCGMData []*continuous.Continuous

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		userID = userTest.RandomID()
		deviceID = dataTest.NewDeviceID()
		highDeviceID = dataTest.NewDeviceID()
		lowDeviceID = dataTest.NewDeviceID()
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

	Context("CalculateWeight", func() {
		It("Returns correct weight for time range", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2015, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 8, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2018, time.Month(1), 14, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 1,
				NewPercentCGMUse: 1,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(*newWeight).To(BeNumerically("~", 0.334, 0.001))
		})

		It("Returns correct weight for time range with <100% cgm use", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 7, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2017, time.Month(1), 14, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 0.3,
				NewPercentCGMUse: 0.5,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(*newWeight).To(BeNumerically("~", 0.66, 0.001))
		})

		It("Returns correct weight for time range with >100% cgm use", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 7, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2017, time.Month(1), 14, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 0.5,
				NewPercentCGMUse: 1.0,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(*newWeight).To(BeNumerically("~", 0.7, 0.001))
		})

		It("Returns error on negative time range", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2018, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 1,
				NewPercentCGMUse: 1,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Invalid time period for calculation, endTime before lastData."))
			Expect(newWeight).To(BeNil())
		})

		It("Returns unchanged date and 1 weight when starttime is after lastdata", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2018, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 1,
				NewPercentCGMUse: 1,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())
			Expect(*newWeight).To(Equal(1.0))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		Context("CalculateStats", func() {
			It("Returns correct average glucose for records", func() {
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.AverageGlucose).To(BeNumerically("~", requestedAvgGlucose, 0.001))
			})

			It("Returns correct time in range value for records", func() {
				dataSetCGMData = NewDataSetCGMDataTimeInRange(deviceID, datumTime, 14, veryLowBloodGlucose, lowBloodGlucose, highBloodGlucose, veryHighBloodGlucose)

				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.TimeInRange).To(Equal(0.200))
				Expect(stats.TimeVeryBelowRange).To(Equal(0.200))
				Expect(stats.TimeBelowRange).To(Equal(0.200))
				Expect(stats.TimeAboveRange).To(Equal(0.200))
				Expect(stats.TimeVeryAboveRange).To(Equal(0.200))
				Expect(stats.TimeCGMUse).Should(BeNumerically("~", 1.000, 0.001))
			})

			It("Returns correct DeviceID competing for records", func() {
				dataSetCGMDataHigh := NewDataSetCGMDataAvg(highDeviceID, datumTime, 14, requestedAvgGlucose+2)

				// here we generate one less day of records to simulate a device with incomplete data
				dataSetCGMDataLow := NewDataSetCGMDataAvg(lowDeviceID, datumTime, 13, requestedAvgGlucose-2)

				dataSetCGMData = append(dataSetCGMDataHigh, dataSetCGMDataLow...)

				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.DeviceID).To(Equal(highDeviceID))
				Expect(stats.AverageGlucose).To(BeNumerically("~", requestedAvgGlucose+2, 0.001))
			})
		})

		Context("ReweightStats", func() {
			var weight float64
			var stats *summary.Stats
			var userSummary *summary.Summary

			BeforeEach(func() {
				stats = &summary.Stats{
					TimeInRange:    1,
					TimeBelowRange: 1,
					TimeAboveRange: 1,
					TimeCGMUse:     1,
					AverageGlucose: 1,
				}

				userSummary = summary.New(userID)
				userSummary.AverageGlucose = &summary.Glucose{
					Value: pointer.FromFloat64(0.0),
				}
				userSummary.TimeInRange = pointer.FromFloat64(0.0)
				userSummary.TimeBelowRange = pointer.FromFloat64(0.0)
				userSummary.TimeAboveRange = pointer.FromFloat64(0.0)
				userSummary.TimeCGMUse = pointer.FromFloat64(0.0)
			})

			It("Returns correctly reweighted stats for 0 weight", func() {
				weight = 0
				reweightedStats, err := summary.ReweightStats(stats, userSummary, weight)

				Expect(err).ToNot(HaveOccurred())
				Expect(reweightedStats.TimeInRange).To(Equal(*userSummary.TimeInRange))
				Expect(reweightedStats.TimeBelowRange).To(Equal(*userSummary.TimeBelowRange))
				Expect(reweightedStats.TimeAboveRange).To(Equal(*userSummary.TimeAboveRange))
				Expect(reweightedStats.TimeCGMUse).To(Equal(*userSummary.TimeCGMUse))
				Expect(reweightedStats.AverageGlucose).To(Equal(*userSummary.AverageGlucose.Value))
			})

			It("Returns correctly reweighted stats for 1 weight", func() {
				weight = 1
				reweightedStats, err := summary.ReweightStats(stats, userSummary, weight)

				Expect(err).ToNot(HaveOccurred())
				Expect(reweightedStats.TimeInRange).To(Equal(stats.TimeInRange))
				Expect(reweightedStats.TimeBelowRange).To(Equal(stats.TimeBelowRange))
				Expect(reweightedStats.TimeAboveRange).To(Equal(stats.TimeAboveRange))
				Expect(reweightedStats.TimeCGMUse).To(Equal(stats.TimeCGMUse))
				Expect(reweightedStats.AverageGlucose).To(Equal(stats.AverageGlucose))
			})

			It("Returns correctly reweighted stats for 0.5 weight", func() {
				weight = 0.5
				midstats := &summary.Stats{
					TimeInRange:    0.5,
					TimeBelowRange: 0.5,
					TimeAboveRange: 0.5,
					TimeCGMUse:     0.5,
					AverageGlucose: 0.5,
				}

				reweightedStats, err := summary.ReweightStats(stats, userSummary, weight)
				Expect(err).ToNot(HaveOccurred())
				Expect(reweightedStats.TimeInRange).To(Equal(midstats.TimeInRange))
				Expect(reweightedStats.TimeBelowRange).To(Equal(midstats.TimeBelowRange))
				Expect(reweightedStats.TimeAboveRange).To(Equal(midstats.TimeAboveRange))
				Expect(reweightedStats.TimeCGMUse).To(Equal(midstats.TimeCGMUse))
				Expect(reweightedStats.AverageGlucose).To(Equal(midstats.AverageGlucose))
			})

			It("Returns error on negative weight", func() {
				weight = -1
				reweightedStats, err := summary.ReweightStats(stats, userSummary, weight)

				Expect(reweightedStats).To(Equal(stats))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid weight (<0||>1) for stats"))
			})

			It("Returns error on greater than 1 weight", func() {
				weight = 2
				reweightedStats, err := summary.ReweightStats(stats, userSummary, weight)

				Expect(reweightedStats).To(Equal(stats))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid weight (<0||>1) for stats"))
			})
		})

		Context("Update", func() {
			var userData []*continuous.Continuous
			var userSummary *summary.Summary
			var status *summary.UserLastUpdated
			var err error
			var newDatumTime time.Time

			//BeforeEach(func() {
			//	datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
			//})

			It("Returns correctly calculated summary with no rolling", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 14, requestedAvgGlucose)
				userSummary = summary.New(userID)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
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
				expectedAverageGlucose := (requestedAvgGlucose-4)*0.333333333 + (requestedAvgGlucose+4)*0.66666666
				expectedCGMUse := (0.5)*0.333333333 + (1)*0.66666666
				expectedGMI := summary.CalculateGMI(expectedAverageGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
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

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", expectedAverageGlucose, 0.001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", expectedCGMUse, 0.001))
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

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
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

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
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

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.00001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 0.07142, 0.00001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(userSummary.GlucoseMgmtIndicator).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 14, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				userSummary, err = summary.Update(ctx, userSummary, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(*userSummary.AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.00001))
				Expect(*userSummary.TimeCGMUse).To(BeNumerically("~", 1.0, 0.00001))
				Expect(userSummary.OutdatedSince).To(BeNil())
				Expect(*userSummary.GlucoseMgmtIndicator).To(BeNumerically("~", expectedGMISecond, 0.00001))
			})
		})
	})
})
