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
		Entry("returns no range for unknown units", pointer.FromString("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.FromString("mmol/L"), 0.0, 10.0),
		Entry("returns expected range for mmol/l units", pointer.FromString("mmol/l"), 0.0, 10.0),
		Entry("returns no range for mg/dL units", pointer.FromString("mg/dL"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for mg/dl units", pointer.FromString("mg/dl"), -math.MaxFloat64, math.MaxFloat64),
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
		Entry("returns unchanged units for unknown units", pointer.FromString("unknown"), pointer.FromString("unknown")),
		Entry("returns mmol/L for mmol/L", pointer.FromString("mmol/L"), pointer.FromString("mmol/L")),
		Entry("returns mmol/L for mmol/l", pointer.FromString("mmol/l"), pointer.FromString("mmol/L")),
		Entry("returns unchanged units for mg/dL", pointer.FromString("mg/dL"), pointer.FromString("mg/dL")),
		Entry("returns unchanged units for mg/dl", pointer.FromString("mg/dl"), pointer.FromString("mg/dl")),
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
		Entry("returns nil for nil value", nil, pointer.FromString("mmol/L"), nil),
		Entry("returns unchanged value for nil units", pointer.FromFloat64(10.0), nil, pointer.FromFloat64(10.0)),
		Entry("returns unchanged value for unknown units", pointer.FromFloat64(10.0), pointer.FromString("unknown"), pointer.FromFloat64(10.0)),
		Entry("returns unchanged value for mmol/L units", pointer.FromFloat64(10.0), pointer.FromString("mmol/L"), pointer.FromFloat64(10.0)),
		Entry("returns unchanged value for mmol/l units", pointer.FromFloat64(10.0), pointer.FromString("mmol/l"), pointer.FromFloat64(10.0)),
		Entry("returns unchanged value for mg/dL units", pointer.FromFloat64(180.0), pointer.FromString("mg/dL"), pointer.FromFloat64(180.0)),
		Entry("returns unchanged value for mg/dl units", pointer.FromFloat64(180.0), pointer.FromString("mg/dl"), pointer.FromFloat64(180.0)),
	)
})
