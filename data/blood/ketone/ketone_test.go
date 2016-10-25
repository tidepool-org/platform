package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/blood/ketone"
)

var _ = Describe("Glucose", func() {
	It("has MmolL", func() {
		Expect(ketone.MmolL).To(Equal("mmol/L"))
	})

	It("has Mmoll", func() {
		Expect(ketone.Mmoll).To(Equal("mmol/l"))
	})

	It("has MmolLLowerLimit", func() {
		Expect(ketone.MmolLLowerLimit).To(Equal(0.0))
	})

	It("has MmolLUpperLimit", func() {
		Expect(ketone.MmolLUpperLimit).To(Equal(10.0))
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
		Entry("returns no range for unknown units", app.StringAsPointer("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", app.StringAsPointer("mmol/L"), 0.0, 10.0),
		Entry("returns expected range for mmol/l units", app.StringAsPointer("mmol/l"), 0.0, 10.0),
		Entry("returns no range for mg/dL units", app.StringAsPointer("mg/dL"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for mg/dl units", app.StringAsPointer("mg/dl"), -math.MaxFloat64, math.MaxFloat64),
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
		Entry("returns unchanged units for unknown units", app.StringAsPointer("unknown"), app.StringAsPointer("unknown")),
		Entry("returns mmol/L for mmol/L", app.StringAsPointer("mmol/L"), app.StringAsPointer("mmol/L")),
		Entry("returns mmol/L for mmol/l", app.StringAsPointer("mmol/l"), app.StringAsPointer("mmol/L")),
		Entry("returns unchanged units for mg/dL", app.StringAsPointer("mg/dL"), app.StringAsPointer("mg/dL")),
		Entry("returns unchanged units for mg/dl", app.StringAsPointer("mg/dl"), app.StringAsPointer("mg/dl")),
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
		Entry("returns nil for nil value", nil, app.StringAsPointer("mmol/L"), nil),
		Entry("returns unchanged value for nil units", app.FloatAsPointer(10.0), nil, app.FloatAsPointer(10.0)),
		Entry("returns unchanged value for unknown units", app.FloatAsPointer(10.0), app.StringAsPointer("unknown"), app.FloatAsPointer(10.0)),
		Entry("returns unchanged value for mmol/L units", app.FloatAsPointer(10.0), app.StringAsPointer("mmol/L"), app.FloatAsPointer(10.0)),
		Entry("returns unchanged value for mmol/l units", app.FloatAsPointer(10.0), app.StringAsPointer("mmol/l"), app.FloatAsPointer(10.0)),
		Entry("returns unchanged value for mg/dL units", app.FloatAsPointer(180.0), app.StringAsPointer("mg/dL"), app.FloatAsPointer(180.0)),
		Entry("returns unchanged value for mg/dl units", app.FloatAsPointer(180.0), app.StringAsPointer("mg/dl"), app.FloatAsPointer(180.0)),
	)
})
