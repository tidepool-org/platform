package bloodglucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/bloodglucose"
)

var _ = Describe("Bloodglucose", func() {
	It("has MmolL", func() {
		Expect(bloodglucose.MmolL).To(Equal("mmol/L"))
	})

	It("has Mmoll", func() {
		Expect(bloodglucose.Mmoll).To(Equal("mmol/l"))
	})

	It("has MgdL", func() {
		Expect(bloodglucose.MgdL).To(Equal("mg/dL"))
	})

	It("has Mgdl", func() {
		Expect(bloodglucose.Mgdl).To(Equal("mg/dl"))
	})

	It("has MmolLToMgdLConversionFactor", func() {
		Expect(bloodglucose.MmolLToMgdLConversionFactor).To(Equal(18.01559))
	})

	It("has MmolLLowerLimit", func() {
		Expect(bloodglucose.MmolLLowerLimit).To(Equal(0.0))
	})

	It("has MmolLUpperLimit", func() {
		Expect(bloodglucose.MmolLUpperLimit).To(Equal(55.0))
	})

	It("has MgdLLowerLimit", func() {
		Expect(bloodglucose.MgdLLowerLimit).To(Equal(0.0))
	})

	It("has MgdLUpperLimit", func() {
		Expect(bloodglucose.MgdLUpperLimit).To(Equal(1000.0))
	})

	Context("ConvertValue", func() {
		DescribeTable("returns expected value when",
			func(value float64, fromUnits string, toUnits string, expected float64) {
				Expect(bloodglucose.ConvertValue(value, fromUnits, toUnits)).To(Equal(expected))
			},
			Entry("has unknown from units", 12.345, "unknown", "mg/dL", 12.345),
			Entry("has unknown to units", 12.345, "mg/dL", "unknown", 12.345),
			Entry("converts from mmol/L to mmol/L", 12.345, "mmol/L", "mmol/L", 12.345),
			Entry("converts from mmol/L to mmol/l", 12.345, "mmol/L", "mmol/l", 12.345),
			Entry("converts from mmol/L to mg/dL", 12.345, "mmol/L", "mg/dL", 222.40245855),
			Entry("converts from mmol/L to mg/dl", 12.345, "mmol/L", "mg/dl", 222.40245855),
			Entry("converts from mmol/l to mmol/L", 12.345, "mmol/l", "mmol/L", 12.345),
			Entry("converts from mmol/l to mmol/l", 12.345, "mmol/l", "mmol/l", 12.345),
			Entry("converts from mmol/l to mg/dL", 12.345, "mmol/l", "mg/dL", 222.40245855),
			Entry("converts from mmol/l to mg/dl", 12.345, "mmol/l", "mg/dl", 222.40245855),
			Entry("converts from mg/dL to mmol/L", 123.0, "mg/dL", "mmol/L", 6.8274200289860065),
			Entry("converts from mg/dL to mmol/l", 123.0, "mg/dL", "mmol/l", 6.8274200289860065),
			Entry("converts from mg/dL to mg/dL", 123.0, "mg/dL", "mg/dL", 123.0),
			Entry("converts from mg/dL to mg/dl", 123.0, "mg/dL", "mg/dl", 123.0),
			Entry("converts from mg/dl to mmol/L", 123.0, "mg/dl", "mmol/L", 6.8274200289860065),
			Entry("converts from mg/dl to mmol/l", 123.0, "mg/dl", "mmol/l", 6.8274200289860065),
			Entry("converts from mg/dl to mg/dL", 123.0, "mg/dl", "mg/dL", 123.0),
			Entry("converts from mg/dl to mg/dl", 123.0, "mg/dl", "mg/dl", 123.0),
		)
	})
})
