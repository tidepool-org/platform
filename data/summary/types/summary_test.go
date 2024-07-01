package types_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

const (
	veryLowBloodGlucose     = 3.0
	lowBloodGlucose         = 3.9
	highBloodGlucose        = 10.0
	veryHighBloodGlucose    = 13.9
	extremeHighBloodGlucose = 19.4
	inTargetBloodGlucose    = 5.0
)

var units = "mmol/L"

type DataRanges struct {
	Min         float64
	Max         float64
	Padding     float64
	VeryLow     float64
	Low         float64
	High        float64
	VeryHigh    float64
	ExtremeHigh float64
}

func NewDataRanges() DataRanges {
	return DataRanges{
		Min:         1,
		Max:         25,
		Padding:     0.01,
		VeryLow:     veryLowBloodGlucose,
		Low:         lowBloodGlucose,
		High:        highBloodGlucose,
		VeryHigh:    veryHighBloodGlucose,
		ExtremeHigh: extremeHighBloodGlucose,
	}
}

func NewDataRangesSingle(value float64) DataRanges {
	return DataRanges{
		Min:         value,
		Max:         value,
		Padding:     0,
		VeryLow:     value,
		Low:         value,
		High:        value,
		VeryHigh:    value,
		ExtremeHigh: value,
	}
}

func NewGlucose(typ *string, units *string, datumTime *time.Time, deviceID *string, uploadId *string) *glucose.Glucose {
	timestamp := time.Now().UTC().Truncate(time.Millisecond)
	datum := glucose.New(*typ)
	datum.Units = units

	datum.Active = true
	datum.ArchivedDataSetID = nil
	datum.ArchivedTime = nil
	datum.CreatedTime = &timestamp
	datum.CreatedUserID = nil
	datum.DeletedTime = nil
	datum.DeletedUserID = nil
	datum.DeviceID = deviceID
	datum.ModifiedTime = &timestamp
	datum.ModifiedUserID = nil
	datum.Time = datumTime
	datum.UploadID = uploadId

	return &datum
}

func ExpectedAverage(windowSize int, hoursAdded int, newAvg float64, oldAvg float64) float64 {
	oldHoursRemaining := windowSize - hoursAdded
	oldAvgTotal := oldAvg * math.Max(float64(oldHoursRemaining), 0)
	newAvgTotal := newAvg * math.Min(float64(hoursAdded), float64(windowSize))

	return (oldAvgTotal + newAvgTotal) / float64(windowSize)
}

func ConvertToIntArray[T data.Datum](arr []T) []interface{} {
	s := make([]interface{}, len(arr))
	for i, v := range arr {
		s[i] = v
	}

	return s
}

var _ = Describe("Summary", func() {
	var datumTime time.Time
	var deviceID string
	var uploadId string

	BeforeEach(func() {
		deviceID = "SummaryTestDevice"
		uploadId = test.RandomSetID()
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("GetDuration", func() {
		var libreDatum *glucose.Glucose
		var otherDatum *glucose.Glucose
		typ := pointer.FromString("cbg")

		It("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
			libreDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID, &uploadId)
			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")

			duration := types.GetDuration(libreDatum)
			Expect(duration).To(Equal(15))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewGlucose(typ, pointer.FromString(units), &datumTime, &deviceID, &uploadId)

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
})
