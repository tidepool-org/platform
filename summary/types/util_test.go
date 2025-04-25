package types_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
)

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
			libreDatum = NewGlucose(typ, pointer.FromString(Units), &datumTime, &deviceID, &uploadId)
			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")

			duration := GetDuration(libreDatum)
			Expect(duration).To(Equal(15))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewGlucose(typ, pointer.FromString(Units), &datumTime, &deviceID, &uploadId)

			duration := GetDuration(otherDatum)
			Expect(duration).To(Equal(5))
		})
	})

	Context("CalculateGMI", func() {
		// input and output examples sourced from https://diabetesjournals.org/care/article/41/11/2275/36593/
		It("Returns correct GMI for medical example 1", func() {
			gmi := CalculateGMI(5.55)
			Expect(gmi).To(Equal(5.7))
		})

		It("Returns correct GMI for medical example 2", func() {
			gmi := CalculateGMI(6.9375)
			Expect(gmi).To(Equal(6.3))
		})

		It("Returns correct GMI for medical example 3", func() {
			gmi := CalculateGMI(8.325)
			Expect(gmi).To(Equal(6.9))
		})

		It("Returns correct GMI for medical example 4", func() {
			gmi := CalculateGMI(9.722)
			Expect(gmi).To(Equal(7.5))
		})

		It("Returns correct GMI for medical example 5", func() {
			gmi := CalculateGMI(11.11)
			Expect(gmi).To(Equal(8.1))
		})

		It("Returns correct GMI for medical example 6", func() {
			gmi := CalculateGMI(12.4875)
			Expect(gmi).To(Equal(8.7))
		})

		It("Returns correct GMI for medical example 7", func() {
			gmi := CalculateGMI(13.875)
			Expect(gmi).To(Equal(9.3))
		})

		It("Returns correct GMI for medical example 8", func() {
			gmi := CalculateGMI(15.2625)
			Expect(gmi).To(Equal(9.9))
		})

		It("Returns correct GMI for medical example 9", func() {
			gmi := CalculateGMI(16.65)
			Expect(gmi).To(Equal(10.5))
		})

		It("Returns correct GMI for medical example 10", func() {
			gmi := CalculateGMI(19.425)
			Expect(gmi).To(Equal(11.7))
		})
	})

	Context("CalculateRealMinutes", func() {
		It("with a full hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 59, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440))
		})

		It("with a half hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 30, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-25))
		})

		It("with a start of hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 1, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-54))
		})

		It("with a near full hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 54, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-1))
		})

		It("with an on the hour endpoint", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 0, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 1440-55))
		})

		It("with 7d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(7, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 7*1440))
		})

		It("with 14d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(14, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 14*1440))
		})

		It("with 30d period", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 55, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(30, lastRecordTime, 5)
			Expect(realMinutes).To(BeNumerically("==", 30*1440))
		})

		It("with 15 minute duration", func() {
			lastRecordTime := time.Date(2016, time.Month(1), 1, 1, 40, 0, 0, time.UTC)
			realMinutes := CalculateWallMinutes(1, lastRecordTime, 15)
			Expect(realMinutes).To(BeNumerically("==", 1440-5))
		})
	})

	Context("BinDelta", func() {
		It("Delta Calc", func() {
			r1 := &Range{
				Minutes: 5,
				Records: 6,
				Percent: 7,
			}
			r2 := &Range{
				Minutes: 2,
				Records: 2,
				Percent: 2,
			}
			d1 := &Range{}
			BinDelta(r1, r2, d1)

			Expect(d1.Minutes).To(Equal(3))
			Expect(d1.Records).To(Equal(4))
			Expect(d1.Percent).To(Equal(5.0))
		})
	})

	Context("Delta", func() {
		It("Delta Calc", func() {
			a := 5
			b := 2
			c := 0
			Delta(&a, &b, &c)

			Expect(c).To(Equal(3))
		})
	})
})
