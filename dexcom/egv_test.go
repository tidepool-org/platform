package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("EGV", func() {
	It("EGVUnitMgdL is expected", func() {
		Expect(dexcom.EGVUnitMgdL).To(Equal("mg/dL"))
	})

	It("EGVUnitMgdLMinute is expected", func() {
		Expect(dexcom.EGVUnitMgdLMinute).To(Equal("mg/dL/min"))
	})

	It("EGVValueMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValueMgdLMaximum).To(Equal(1000.0))
	})

	It("EGVValueMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValueMgdLMinimum).To(Equal(0.0))
	})

	It("EGVValuePinnedMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMaximum).To(Equal(400.0))
	})

	It("EGVValuePinnedMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMinimum).To(Equal(40.0))
	})

	It("EGVStatusHigh is expected", func() {
		Expect(dexcom.EGVStatusHigh).To(Equal("high"))
	})

	It("EGVStatusLow is expected", func() {
		Expect(dexcom.EGVStatusLow).To(Equal("low"))
	})

	It("EGVStatusOK is expected", func() {
		Expect(dexcom.EGVStatusOK).To(Equal("ok"))
	})

	It("EGVStatusOutOfCalibration is expected", func() {
		Expect(dexcom.EGVStatusOutOfCalibration).To(Equal("outOfCalibration"))
	})

	It("EGVStatusSensorNoise is expected", func() {
		Expect(dexcom.EGVStatusSensorNoise).To(Equal("sensorNoise"))
	})

	It("EGVTrendDoubleUp is expected", func() {
		Expect(dexcom.EGVTrendDoubleUp).To(Equal("doubleUp"))
	})

	It("EGVTrendSingleUp is expected", func() {
		Expect(dexcom.EGVTrendSingleUp).To(Equal("singleUp"))
	})

	It("EGVTrendFortyFiveUp is expected", func() {
		Expect(dexcom.EGVTrendFortyFiveUp).To(Equal("fortyFiveUp"))
	})

	It("EGVTrendFlat is expected", func() {
		Expect(dexcom.EGVTrendFlat).To(Equal("flat"))
	})

	It("EGVTrendFortyFiveDown is expected", func() {
		Expect(dexcom.EGVTrendFortyFiveDown).To(Equal("fortyFiveDown"))
	})

	It("EGVTrendSingleDown is expected", func() {
		Expect(dexcom.EGVTrendSingleDown).To(Equal("singleDown"))
	})

	It("EGVTrendDoubleDown is expected", func() {
		Expect(dexcom.EGVTrendDoubleDown).To(Equal("doubleDown"))
	})

	It("EGVTrendNone is expected", func() {
		Expect(dexcom.EGVTrendNone).To(Equal("none"))
	})

	It("EGVTrendNotComputable is expected", func() {
		Expect(dexcom.EGVTrendNotComputable).To(Equal("notComputable"))
	})

	It("EGVTrendRateOutOfRange is expected", func() {
		Expect(dexcom.EGVTrendRateOutOfRange).To(Equal("rateOutOfRange"))
	})

	It("EGVTransmitterTickMinimum is expected", func() {
		Expect(dexcom.EGVTransmitterTickMinimum).To(Equal(0))
	})

	It("EGVsResponseRateUnits returns expected", func() {
		Expect(dexcom.EGVsResponseRateUnits()).To(Equal([]string{"mg/dL/min"}))
	})

	It("EGVsResponseUnits returns expected", func() {
		Expect(dexcom.EGVsResponseUnits()).To(Equal([]string{"mg/dL"}))
	})

	It("EGVStatuses returns expected", func() {
		Expect(dexcom.EGVStatuses()).To(Equal([]string{"high", "low", "ok", "outOfCalibration", "sensorNoise"}))
	})

	It("EGVTrends returns expected", func() {
		Expect(dexcom.EGVTrends()).To(Equal([]string{"doubleUp", "singleUp", "fortyFiveUp", "flat", "fortyFiveDown", "singleDown", "doubleDown", "none", "notComputable", "rateOutOfRange"}))
	})
})
