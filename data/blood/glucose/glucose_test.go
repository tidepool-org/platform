package glucose_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Glucose", func() {
	It("has MmolL", func() {
		Expect(glucose.MmolL).To(Equal("mmol/L"))
	})

	It("has Mmoll", func() {
		Expect(glucose.Mmoll).To(Equal("mmol/l"))
	})

	It("has MmolLMinute", func() {
		Expect(glucose.MmolLMinute).To(Equal("mmol/L/minute"))
	})

	It("has MgdL", func() {
		Expect(glucose.MgdL).To(Equal("mg/dL"))
	})

	It("has Mgdl", func() {
		Expect(glucose.Mgdl).To(Equal("mg/dl"))
	})

	It("has MgdLMinute", func() {
		Expect(glucose.MgdLMinute).To(Equal("mg/dL/minute"))
	})

	It("has MmolLMinimum", func() {
		Expect(glucose.MmolLMinimum).To(Equal(0.0))
	})

	It("has MmolLMaximum", func() {
		Expect(glucose.MmolLMaximum).To(Equal(55.0))
	})

	It("has MmolLMinuteMinimum", func() {
		Expect(glucose.MmolLMinuteMinimum).To(Equal(-5.5))
	})

	It("has MmolLMinuteMaximum", func() {
		Expect(glucose.MmolLMinuteMaximum).To(Equal(5.5))
	})

	It("has MgdLMinimum", func() {
		Expect(glucose.MgdLMinimum).To(Equal(0.0))
	})

	It("has MgdLMaximum", func() {
		Expect(glucose.MgdLMaximum).To(Equal(1000.0))
	})

	It("has MgdLMinuteMinimum", func() {
		Expect(glucose.MgdLMinuteMinimum).To(Equal(-100.0))
	})

	It("has MgdLMinuteMaximum", func() {
		Expect(glucose.MgdLMinuteMaximum).To(Equal(100.0))
	})

	It("has MmolLToMgdLConversionFactor", func() {
		Expect(glucose.MmolLToMgdLConversionFactor).To(Equal(18.01559))
	})

	It("has MmolLToMgdLPrecisionFactor", func() {
		Expect(glucose.MmolLToMgdLPrecisionFactor).To(Equal(100000.0))
	})

	Context("Units", func() {
		It("returns the expected units", func() {
			Expect(glucose.Units()).To(ConsistOf("mmol/L", "mmol/l", "mg/dL", "mg/dl"))
		})
	})

	Context("RateUnits", func() {
		It("returns the expected units", func() {
			Expect(glucose.RateUnits()).To(ConsistOf("mmol/L/minute", "mg/dL/minute"))
		})
	})

	DescribeTable("ValueRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.ValueRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", pointer.FromString("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.FromString("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", pointer.FromString("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", pointer.FromString("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", pointer.FromString("mg/dl"), 0.0, 1000.0),
	)

	DescribeTable("NormalizeUnits",
		func(units *string, expectedUnits *string) {
			actualUnits := glucose.NormalizeUnits(units)
			if expectedUnits == nil {
				Expect(actualUnits).To(BeNil())
			} else {
				Expect(actualUnits).ToNot(BeNil())
				Expect(*actualUnits).To(Equal(*expectedUnits))
			}
		},
		Entry("returns nil for nil", nil, nil),
		Entry("returns unchanged units for unknown units", pointer.FromString("unknown"), pointer.FromString("unknown")),
		Entry("returns mmol/L for mmol/L", pointer.FromString("mmol/L"), pointer.FromString("mmol/L")),
		Entry("returns mmol/L for mmol/l", pointer.FromString("mmol/l"), pointer.FromString("mmol/L")),
		Entry("returns mmol/L for mg/dL", pointer.FromString("mg/dL"), pointer.FromString("mmol/L")),
		Entry("returns mmol/L for mg/dl", pointer.FromString("mg/dl"), pointer.FromString("mmol/L")),
	)

	Context("NormalizeValueForUnits", func() {
		DescribeTable("given value and units",
			func(value *float64, units *string, expectedValue *float64) {
				actualValue := glucose.NormalizeValueForUnits(value, units)
				if expectedValue == nil {
					Expect(actualValue).To(BeNil())
				} else {
					Expect(actualValue).ToNot(BeNil())
					Expect(*actualValue).To(Equal(*expectedValue))
				}
			},
			Entry("returns nil for nil value", nil, pointer.FromString("mmol/L"), nil),
			Entry("returns unchanged value for nil units", pointer.FromFloat64(10.0), nil, pointer.FromFloat64(10.0)),
			Entry("returns unchanged value for unknown units", pointer.FromFloat64(10.0), pointer.FromString("unknown"), pointer.FromFloat64(10.0)),
			Entry("returns unchanged value for mmol/L units", pointer.FromFloat64(10.0), pointer.FromString("mmol/L"), pointer.FromFloat64(10.0)),
			Entry("returns unchanged value for mmol/l units", pointer.FromFloat64(10.0), pointer.FromString("mmol/l"), pointer.FromFloat64(10.0)),
			Entry("returns converted value for mg/dL units", pointer.FromFloat64(180.0), pointer.FromString("mg/dL"), pointer.FromFloat64(9.99135)),
			Entry("returns converted value for mg/dl units", pointer.FromFloat64(180.0), pointer.FromString("mg/dl"), pointer.FromFloat64(9.99135)),
		)

		It("properly normalizes a range of mmol/L values", func() {
			for value := glucose.MmolLMinimum; value <= glucose.MmolLMaximum; value += 0.1 {
				normalizedValue := glucose.NormalizeValueForUnits(pointer.FromFloat64(float64(value)), pointer.FromString("mmol/L"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(*normalizedValue).To(Equal(value))
			}
		})

		It("properly normalizes a range of mg/dL values", func() {
			for value := int(glucose.MgdLMinimum); value <= int(glucose.MgdLMaximum); value++ {
				normalizedValue := glucose.NormalizeValueForUnits(pointer.FromFloat64(float64(value)), pointer.FromString("mg/dL"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(int(*normalizedValue*18.01559 + 0.5)).To(Equal(value))
			}
		})
	})

	DescribeTable("ValueRangeForRateUnits",
		func(rateUnits *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.ValueRangeForRateUnits(rateUnits)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", pointer.FromString("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L/minute units", pointer.FromString("mmol/L/minute"), -5.5, 5.5),
		Entry("returns expected range for mg/dL/minute units", pointer.FromString("mg/dL/minute"), -100.0, 100.0),
	)

	DescribeTable("NormalizeRateUnits",
		func(rateUnits *string, expectedRateUnits *string) {
			actualRateUnits := glucose.NormalizeRateUnits(rateUnits)
			if expectedRateUnits == nil {
				Expect(actualRateUnits).To(BeNil())
			} else {
				Expect(actualRateUnits).ToNot(BeNil())
				Expect(*actualRateUnits).To(Equal(*expectedRateUnits))
			}
		},
		Entry("returns nil for nil", nil, nil),
		Entry("returns unchanged units for unknown units", pointer.FromString("unknown"), pointer.FromString("unknown")),
		Entry("returns mmol/L/minute for mmol/L/minute", pointer.FromString("mmol/L/minute"), pointer.FromString("mmol/L/minute")),
		Entry("returns mmol/L/minute for mg/dL/minute", pointer.FromString("mg/dL/minute"), pointer.FromString("mmol/L/minute")),
	)

	Context("NormalizeValueForRateUnits", func() {
		DescribeTable("given value and units",
			func(value *float64, rateUnits *string, expectedValue *float64) {
				actualValue := glucose.NormalizeValueForRateUnits(value, rateUnits)
				if expectedValue == nil {
					Expect(actualValue).To(BeNil())
				} else {
					Expect(actualValue).ToNot(BeNil())
					Expect(*actualValue).To(Equal(*expectedValue))
				}
			},
			Entry("returns nil for nil value", nil, pointer.FromString("mmol/L/minute"), nil),
			Entry("returns unchanged value for nil units", pointer.FromFloat64(1.0), nil, pointer.FromFloat64(1.0)),
			Entry("returns unchanged value for unknown units", pointer.FromFloat64(1.0), pointer.FromString("unknown"), pointer.FromFloat64(1.0)),
			Entry("returns unchanged value for mmol/L/minute units", pointer.FromFloat64(1.0), pointer.FromString("mmol/L/minute"), pointer.FromFloat64(1.0)),
			Entry("returns converted value for mg/dL/minute units", pointer.FromFloat64(18.0), pointer.FromString("mg/dL/minute"), pointer.FromFloat64(0.99913)),
		)

		It("properly normalizes a range of mmol/L/minute values", func() {
			for value := glucose.MmolLMinuteMinimum; value <= glucose.MmolLMinuteMaximum; value += 0.1 {
				normalizedValue := glucose.NormalizeValueForRateUnits(pointer.FromFloat64(float64(value)), pointer.FromString("mmol/L/minute"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(*normalizedValue).To(Equal(value))
			}
		})

		It("properly normalizes a range of mg/dL/minute values", func() {
			for value := int(glucose.MgdLMinuteMinimum); value <= int(glucose.MgdLMinuteMaximum); value++ {
				normalizedValue := glucose.NormalizeValueForRateUnits(pointer.FromFloat64(float64(value)), pointer.FromString("mg/dL/minute"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(int(*normalizedValue*18.01559 + math.Copysign(0.5, float64(value)))).To(Equal(value))
			}
		})
	})
})
