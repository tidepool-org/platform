package types_test

import (
	"math"
	"time"

	userTest "github.com/tidepool-org/platform/user/test"

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
	inTargetBloodGlucose = 5.0
	units                = "mmol/L"
)

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

func ExpectedAverage(windowSize int, hoursAdded int, newAvg float64, oldAvg float64) float64 {
	oldHoursRemaining := windowSize - hoursAdded
	oldAvgTotal := oldAvg * math.Max(float64(oldHoursRemaining), 0)
	newAvgTotal := newAvg * math.Min(float64(hoursAdded), float64(windowSize))

	return (oldAvgTotal + newAvgTotal) / float64(windowSize)
}

var _ = Describe("Summary", func() {
	var datumTime time.Time
	var deviceID string

	BeforeEach(func() {
		deviceID = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
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

	Context("CalculateRealMinutes", func() {
		It("with a full hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 59, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440))
		})

		It("with a half hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 30, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-25))
		})

		It("with a start of hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 1, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-54))
		})

		It("with a near full hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 54, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-1))
		})

		It("with an on the hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 0, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-55))
		})

		It("with 7d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(7, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 7*1440))
		})

		It("with 14d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(14, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 14*1440))
		})

		It("with 30d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(30, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 30*1440))
		})

		It("with 15 minute duration", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 40, 0, 0, time.UTC)
			realMinutes := types.CalculateRealMinutes(1, lastRecordTime, 15)
			Expect(realMinutes).To(BeNumerically("==", 1440-5))
		})
	})

	Context("GetStartTime", func() {
		var userId string
		var userSummary *types.Summary[types.CGMStats, *types.CGMStats]

		BeforeEach(func() {
			userId = userTest.RandomID()
			userSummary = types.Create[types.CGMStats, *types.CGMStats](userId)
		})

		// NOTE we use CGM types here, but it doesn't matter for the test

		It("Returns correct start time for summary >60d behind", func() {
			timestamp := time.Date(2020, time.Month(1), 1, 1, 1, 0, 0, time.UTC)

			userSummary.Dates.LastData = &timestamp
			status := types.UserLastUpdated{
				LastData: timestamp.AddDate(0, 0, types.HoursAgoToKeep/24+1),
			}
			startTime := types.GetStartTime(userSummary, &status)
			Expect(startTime).To(Equal(status.LastData.AddDate(0, 0, -types.HoursAgoToKeep/24)))
		})

		It("Returns correct start time for summary <60d behind", func() {
			timestamp := time.Date(2020, time.Month(1), 1, 1, 1, 0, 0, time.UTC)

			userSummary.Dates.LastData = &timestamp
			status := types.UserLastUpdated{
				LastData: timestamp.AddDate(0, 0, types.HoursAgoToKeep/24/2),
			}
			startTime := types.GetStartTime(userSummary, &status)
			Expect(startTime).To(Equal(status.LastData.AddDate(0, 0, -types.HoursAgoToKeep/24/2)))
		})
	})
})
