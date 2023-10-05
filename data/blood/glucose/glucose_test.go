package glucose_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	It("has MgdL", func() {
		Expect(glucose.MgdL).To(Equal("mg/dL"))
	})

	It("has Mgdl", func() {
		Expect(glucose.Mgdl).To(Equal("mg/dl"))
	})

	It("has MmolLMinimum", func() {
		Expect(glucose.MmolLMinimum).To(Equal(0.0))
	})

	It("has MmolLMaximum", func() {
		Expect(glucose.MmolLMaximum).To(Equal(55.0))
	})

	It("has MgdLMinimum", func() {
		Expect(glucose.MgdLMinimum).To(Equal(0.0))
	})

	It("has MgdLMaximum", func() {
		Expect(glucose.MgdLMaximum).To(Equal(1000.0))
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

	Context("Convert", func() {

		It("panics on unhandled units", func() {
			Expect(func() { glucose.Convert(0, "foo", glucose.MgdL) }).Should(Panic())
			Expect(func() { glucose.Convert(0, glucose.MgdL, "foo") }).Should(Panic())
		})

		It("returns equivalent units unchanged", func() {
			var got float64
			expected := float64(9.99135198374)
			input := expected

			got = glucose.Convert(input, glucose.MgdL, glucose.MgdL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mgdl, glucose.MgdL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.MgdL, glucose.Mgdl)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mgdl, glucose.Mgdl)
			Expect(got).To(Equal(expected))

			got = glucose.Convert(input, glucose.MmolL, glucose.MmolL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mmoll, glucose.MmolL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.MmolL, glucose.Mmoll)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mmoll, glucose.Mmoll)
			Expect(got).To(Equal(expected))
		})

		It("converts from mg/dL => mmol/L", func() {
			var got float64
			expected := float64(9.99135)
			input := float64(180.0)

			got = glucose.Convert(input, glucose.MgdL, glucose.MmolL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mgdl, glucose.MmolL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.MgdL, glucose.Mmoll)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mgdl, glucose.Mmoll)
			Expect(got).To(Equal(expected))
		})

		It("converts from mmol/L => mg/dL", func() {
			var got float64
			expected := float64(180.00007)
			input := float64(9.99135)

			got = glucose.Convert(input, glucose.MmolL, glucose.MgdL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.MmolL, glucose.Mgdl)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mmoll, glucose.MgdL)
			Expect(got).To(Equal(expected))
			got = glucose.Convert(input, glucose.Mmoll, glucose.Mgdl)
			Expect(got).To(Equal(expected))
		})

	})
})
