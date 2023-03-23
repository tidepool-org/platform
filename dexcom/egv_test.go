package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure/validator"
	platform_test "github.com/tidepool-org/platform/test"
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
	Describe("Validate", func() {
		var getTestEVG = func() *dexcom.EGV {
			return test.RandomEGV(pointer.FromString(platform_test.RandomStringFromArray(dexcom.EGVsResponseUnits())))
		}
		DescribeTable("requires",
			func(setupEGVFunc func() *dexcom.EGV) {
				testEGV := setupEGVFunc()
				validator := validator.New()
				testEGV.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("systemTime to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.SystemTime = nil
				return egv
			}),
			Entry("displayTime to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.DisplayTime = nil
				return egv
			}),
			Entry("id to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.ID = nil
				return egv
			}),
			Entry("transmitterTicks to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.TransmitterTicks = nil
				return egv
			}),
			Entry("transmitterGeneration to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.TransmitterGeneration = nil
				return egv
			}),
			Entry("unit to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.Unit = nil
				return egv
			}),
			Entry("displayDevice to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.DisplayDevice = nil
				return egv
			}),
		)
		DescribeTable("does not require",
			func(setupEGVFunc func() *dexcom.EGV) {
				testEGV := setupEGVFunc()
				validator := validator.New()
				testEGV.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			},
			Entry("transmitterID to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.TransmitterID = nil
				return egv
			}),
			Entry("value to be set if unknown units", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.Unit = pointer.FromString(dexcom.EGVUnitUnknown)
				egv.Value = nil
				return egv
			}),
			Entry("trendRate to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.TrendRate = nil
				return egv
			}),
			Entry("trend to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.Trend = nil
				return egv
			}),
			Entry("status to be set", func() *dexcom.EGV {
				egv := getTestEVG()
				egv.Status = nil
				return egv
			}),
		)
	})
})
