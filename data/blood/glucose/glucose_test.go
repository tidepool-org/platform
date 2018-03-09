package glucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

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

	It("has MmolLPrecisionFactor", func() {
		Expect(glucose.MmolLPrecisionFactor).To(Equal(100000.0))
	})

	It("has MgdLPrecisionFactor", func() {
		Expect(glucose.MgdLPrecisionFactor).To(Equal(1.0))
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
		Entry("returns no range for unknown units", pointer.String("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.String("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", pointer.String("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", pointer.String("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", pointer.String("mg/dl"), 0.0, 1000.0),
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
		Entry("returns unchanged units for unknown units", pointer.String("unknown"), pointer.String("unknown")),
		Entry("returns mmol/L for mmol/L", pointer.String("mmol/L"), pointer.String("mmol/L")),
		Entry("returns mmol/L for mmol/l", pointer.String("mmol/l"), pointer.String("mmol/L")),
		Entry("returns mmol/L for mg/dL", pointer.String("mg/dL"), pointer.String("mmol/L")),
		Entry("returns mmol/L for mg/dl", pointer.String("mg/dl"), pointer.String("mmol/L")),
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
			Entry("returns nil for nil value", nil, pointer.String("mmol/L"), nil),
			Entry("returns unchanged value for nil units", pointer.Float64(10.123456789012345), nil, pointer.Float64(10.123456789012345)),
			Entry("returns unchanged value for unknown units", pointer.Float64(10.123456789012345), pointer.String("unknown"), pointer.Float64(10.123456789012345)),
			Entry("returns unchanged value for mmol/L units", pointer.Float64(10.123456789012345), pointer.String("mmol/L"), pointer.Float64(10.123456789012345)),
			Entry("returns unchanged value for mmol/l units", pointer.Float64(10.123456789012345), pointer.String("mmol/l"), pointer.Float64(10.123456789012345)),
			Entry("returns converted value with limited precision for mg/dL units", pointer.Float64(180.123456789012345), pointer.String("mg/dL"), pointer.Float64(9.99135)),
			Entry("returns converted value with limited precision for mg/dl units", pointer.Float64(180.123456789012345), pointer.String("mg/dl"), pointer.Float64(9.99135)),
		)

		It("properly normalizes a range of mmol/L values", func() {
			for value := glucose.MmolLMinimum; value <= glucose.MmolLMaximum; value += 0.1 {
				normalizedValue := glucose.NormalizeValueForUnits(pointer.Float64(float64(value)), pointer.String("mmol/L"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(*normalizedValue).To(Equal(value))
			}
		})

		It("properly normalizes a range of mg/dL values", func() {
			for value := int(glucose.MgdLMinimum); value <= int(glucose.MgdLMaximum); value++ {
				normalizedValue := glucose.NormalizeValueForUnits(pointer.Float64(float64(value)), pointer.String("mg/dL"))
				Expect(normalizedValue).ToNot(BeNil())
				Expect(int(*normalizedValue*18.01559 + 0.5)).To(Equal(value))
			}
		})
	})

	Context("NormalizePrecisionForUnits", func() {
		DescribeTable("given value and units",
			func(value *float64, units *string, expectedValue *float64) {
				actualValue := glucose.NormalizePrecisionForUnits(value, units)
				if expectedValue == nil {
					Expect(actualValue).To(BeNil())
				} else {
					Expect(actualValue).ToNot(BeNil())
					Expect(*actualValue).To(Equal(*expectedValue))
				}
			},
			Entry("returns nil for nil value", nil, pointer.String("mmol/L"), nil),
			Entry("returns unchanged value for nil units", pointer.Float64(10.123456789012345), nil, pointer.Float64(10.123456789012345)),
			Entry("returns unchanged value for unknown units", pointer.Float64(10.123456789012345), pointer.String("unknown"), pointer.Float64(10.123456789012345)),
			Entry("returns value with limited precision for mmol/L units", pointer.Float64(10.123456789012345), pointer.String("mmol/L"), pointer.Float64(10.12346)),
			Entry("returns value with limited precision for mmol/l units", pointer.Float64(10.123456789012345), pointer.String("mmol/l"), pointer.Float64(10.12346)),
			Entry("returns value with limited precision for mg/dL units", pointer.Float64(180.123456789012345), pointer.String("mg/dL"), pointer.Float64(180)),
			Entry("returns value with limited precision for mg/dl units", pointer.Float64(180.123456789012345), pointer.String("mg/dl"), pointer.Float64(180)),
		)
	})
})
