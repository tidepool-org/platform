package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("EGV", func() {

	It("EGVValueMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValueMgdLMaximum).To(Equal(1000.0))
	})

	It("EGVValueMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValueMgdLMinimum).To(Equal(0.0))
	})

	It("EGVValuePinnedMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMinimum).To(Equal(40.0))
	})

	It("EGVValuePinnedMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMaximum).To(Equal(400.0))
	})

	It("EGVValuePinnedMmolLMinimum is expected", func() {
		Expect(dexcom.EGVValuePinnedMmolLMinimum).To(Equal(2.22030))
	})

	It("EGVValuePinnedMmolLMaximum is expected", func() {
		Expect(dexcom.EGVValuePinnedMmolLMaximum).To(Equal(22.20299))
	})

	It("EGVTransmitterTickMinimum is expected", func() {
		Expect(dexcom.EGVTransmitterTickMinimum).To(Equal(0))
	})

	It("EGVsResponseRateUnits returns expected", func() {
		Expect(dexcom.EGVsResponseRateUnits()).To(Equal([]string{"unknown", "mg/dL/min", "mmol/L/min"}))
		Expect(dexcom.EGVsResponseRateUnits()).To(Equal([]string{dexcom.EGVUnitUnknown, dexcom.EGVUnitMgdLMinute, dexcom.EGVUnitMmolLMinute}))
	})

	It("EGVsResponseUnits returns expected", func() {
		Expect(dexcom.EGVsResponseUnits()).To(Equal([]string{"unknown", "mg/dL", "mmol/L"}))
		Expect(dexcom.EGVsResponseUnits()).To(Equal([]string{dexcom.EGVUnitUnknown, dexcom.EGVUnitMgdL, dexcom.EGVUnitMmolL}))
	})

	It("EGVStatuses returns expected", func() {
		Expect(dexcom.EGVStatuses()).To(Equal([]string{"unknown", "high", "low", "ok"}))
		Expect(dexcom.EGVStatuses()).To(Equal([]string{
			dexcom.EGVStatusUnknown,
			dexcom.EGVStatusHigh,
			dexcom.EGVStatusLow,
			dexcom.EGVStatusOK,
		}))
	})

	It("EGVTrends returns expected", func() {
		Expect(dexcom.EGVTrends()).To(Equal([]string{"doubleUp", "singleUp", "fortyFiveUp", "flat", "fortyFiveDown", "singleDown", "doubleDown", "none", "notComputable", "rateOutOfRange"}))
		Expect(dexcom.EGVTrends()).To(Equal([]string{
			dexcom.EGVTrendDoubleUp,
			dexcom.EGVTrendSingleUp,
			dexcom.EGVTrendFortyFiveUp,
			dexcom.EGVTrendFlat,
			dexcom.EGVTrendFortyFiveDown,
			dexcom.EGVTrendSingleDown,
			dexcom.EGVTrendDoubleDown,
			dexcom.EGVTrendNone,
			dexcom.EGVTrendNotComputable,
			dexcom.EGVTrendRateOutOfRange,
		}))
	})
})
