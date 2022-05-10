package summary_test

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dataTest "github.com/tidepool-org/platform/data/test"
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

func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, reqAvg float64) []*continuous.Continuous {
	dataSetData := []*continuous.Continuous{}

	// generate 2 weeks of data
	for count := 0; count < 4032; count += 2 {
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
func NewDataSetCGMDataTimeInRange(deviceID string, startTime time.Time, veryLow float64, low float64, high float64, veryHigh float64) []*continuous.Continuous {
	dataSetData := []*continuous.Continuous{}

	glucoseBrackets := [5][2]float64{
		{1, veryLow - 0.01},
		{veryLow, low - 0.01},
		{low, high - 0.01},
		{high, veryHigh - 0.01},
		{veryHigh, 20},
	}

	// generate 2 weeks of data
	for count := 0; count < 4032; count += 5 {
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
	var userID string
	var datumTime time.Time
	var deviceID string
	var highDeviceID string
	var lowDeviceID string
	var dataSetCGMData []*continuous.Continuous

	BeforeEach(func() {
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

	Context("CalculateWeight", func() {
		It("Returns correct weight for time range", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2015, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2018, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 1,
				NewPercentCGMUse: 1,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(*newWeight).To(BeNumerically("~", 0.333, 0.001))
		})

		It("Returns correct weight for time range with <100% cgm use", func() {
			input := summary.WeightingInput{
				StartTime:        time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				EndTime:          time.Date(2018, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				LastData:         time.Date(2017, time.Month(1), 1, 0, 0, 0, 0, time.UTC),
				OldPercentCGMUse: 0.3,
				NewPercentCGMUse: 0.5,
			}

			newWeight, err := summary.CalculateWeight(&input)
			Expect(err).ToNot(HaveOccurred())

			Expect(*newWeight).To(BeNumerically("~", 0.625, 0.001))
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
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, requestedAvgGlucose)
				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.AverageGlucose).To(BeNumerically("~", requestedAvgGlucose, 0.001))
			})

			It("Returns correct time in range value for records", func() {
				dataSetCGMData = NewDataSetCGMDataTimeInRange(deviceID, datumTime, veryLowBloodGlucose, lowBloodGlucose, highBloodGlucose, veryHighBloodGlucose)

				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.TimeInRange).To(Equal(0.200))
				Expect(stats.TimeVeryBelowRange).To(Equal(0.200))
				Expect(stats.TimeBelowRange).To(Equal(0.200))
				Expect(stats.TimeAboveRange).To(Equal(0.200))
				Expect(stats.TimeVeryAboveRange).To(Equal(0.200))
				Expect(stats.TimeCGMUse).Should(BeNumerically("~", 1.000, 0.001))
			})

			It("Returns correct DeviceID competing for records", func() {
				dataSetCGMDataHigh := NewDataSetCGMDataAvg(highDeviceID, datumTime, requestedAvgGlucose+2)

				// here we chop off the last ~1000 records to simulate 2 devices with incomplete data
				dataSetCGMDataLow := NewDataSetCGMDataAvg(lowDeviceID, datumTime, requestedAvgGlucose-2)[:3000]

				dataSetCGMData = append(dataSetCGMDataHigh, dataSetCGMDataLow...)

				stats := summary.CalculateStats(dataSetCGMData, 20160)

				Expect(stats.DeviceID).To(Equal(highDeviceID))
				Expect(stats.AverageGlucose).To(Equal(requestedAvgGlucose + 2))
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
	})
})
