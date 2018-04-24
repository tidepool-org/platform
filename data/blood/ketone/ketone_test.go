package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/data/blood/ketone"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Ketone", func() {
	It("has MmolL", func() {
		Expect(ketone.MmolL).To(Equal("mmol/L"))
	})

	It("has Mmoll", func() {
		Expect(ketone.Mmoll).To(Equal("mmol/l"))
	})

	It("has MmolLMinimum", func() {
		Expect(ketone.MmolLMinimum).To(Equal(0.0))
	})

	It("has MmolLMaximum", func() {
		Expect(ketone.MmolLMaximum).To(Equal(10.0))
	})

	Context("Units", func() {
		It("returns the expected units", func() {
			Expect(ketone.Units()).To(ConsistOf("mmol/L", "mmol/l"))
		})
	})

	DescribeTable("ValueRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := ketone.ValueRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", pointer.String("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.String("mmol/L"), 0.0, 10.0),
		Entry("returns expected range for mmol/l units", pointer.String("mmol/l"), 0.0, 10.0),
		Entry("returns no range for mg/dL units", pointer.String("mg/dL"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for mg/dl units", pointer.String("mg/dl"), -math.MaxFloat64, math.MaxFloat64),
	)

	DescribeTable("NormalizeUnits",
		func(units *string, expectedUnits *string) {
			actualUnits := ketone.NormalizeUnits(units)
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
		Entry("returns unchanged units for mg/dL", pointer.String("mg/dL"), pointer.String("mg/dL")),
		Entry("returns unchanged units for mg/dl", pointer.String("mg/dl"), pointer.String("mg/dl")),
	)

	DescribeTable("NormalizeValueForUnits",
		func(value *float64, units *string, expectedValue *float64) {
			actualValue := ketone.NormalizeValueForUnits(value, units)
			if expectedValue == nil {
				Expect(actualValue).To(BeNil())
			} else {
				Expect(actualValue).ToNot(BeNil())
				Expect(*actualValue).To(Equal(*expectedValue))
			}
		},
		Entry("returns nil for nil value", nil, pointer.String("mmol/L"), nil),
		Entry("returns unchanged value for nil units", pointer.Float64(10.0), nil, pointer.Float64(10.0)),
		Entry("returns unchanged value for unknown units", pointer.Float64(10.0), pointer.String("unknown"), pointer.Float64(10.0)),
		Entry("returns unchanged value for mmol/L units", pointer.Float64(10.0), pointer.String("mmol/L"), pointer.Float64(10.0)),
		Entry("returns unchanged value for mmol/l units", pointer.Float64(10.0), pointer.String("mmol/l"), pointer.Float64(10.0)),
		Entry("returns unchanged value for mg/dL units", pointer.Float64(180.0), pointer.String("mg/dL"), pointer.Float64(180.0)),
		Entry("returns unchanged value for mg/dl units", pointer.Float64(180.0), pointer.String("mg/dl"), pointer.Float64(180.0)),
	)
})
