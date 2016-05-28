package normalizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/normalizer"
)

var _ = Describe("StandardBloodGlucose", func() {
	Context("when units are nil", func() {
		var standard *normalizer.StandardBloodGlucose

		BeforeEach(func() {
			standard = normalizer.NewStandardBloodGlucose(nil)
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("returns nil units", func() {
			Expect(standard.Units()).To(BeNil())
		})

		It("returns nil value", func() {
			Expect(standard.Value(nil)).To(BeNil())
		})

		It("returns the value unchanged", func() {
			value := 123.45
			Expect(*standard.Value(&value)).To(Equal(123.45))
		})

		It("returns nil units and nil value", func() {
			resultUnits, resultValue := standard.UnitsAndValue(nil)
			Expect(resultUnits).To(BeNil())
			Expect(resultValue).To(BeNil())
		})

		It("returns nil units and the value unchanged", func() {
			value := 123.45
			resultUnits, resultValue := standard.UnitsAndValue(&value)
			Expect(resultUnits).To(BeNil())
			Expect(*resultValue).To(Equal(123.45))
		})
	})

	Context("when units are unknown", func() {
		var standard *normalizer.StandardBloodGlucose

		BeforeEach(func() {
			units := "unknown"
			standard = normalizer.NewStandardBloodGlucose(&units)
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("returns nil units", func() {
			Expect(standard.Units()).To(BeNil())
		})

		It("returns nil value", func() {
			Expect(standard.Value(nil)).To(BeNil())
		})

		It("returns the value unchanged", func() {
			value := 123.45
			Expect(*standard.Value(&value)).To(Equal(123.45))
		})

		It("returns nil units and nil value", func() {
			resultUnits, resultValue := standard.UnitsAndValue(nil)
			Expect(resultUnits).To(BeNil())
			Expect(resultValue).To(BeNil())
		})

		It("returns nil units and the value unchanged", func() {
			value := 123.45
			resultUnits, resultValue := standard.UnitsAndValue(&value)
			Expect(resultUnits).To(BeNil())
			Expect(*resultValue).To(Equal(123.45))
		})
	})

	Context("when units are valid", func() {
		DescribeTable("and value is nil",
			func(units string, expectedUnits string) {
				standard := normalizer.NewStandardBloodGlucose(&units)
				Expect(standard).ToNot(BeNil())
				Expect(*standard.Units()).To(Equal(expectedUnits))
				Expect(standard.Value(nil)).To(BeNil())
				resultUnits, resultValue := standard.UnitsAndValue(nil)
				Expect(*resultUnits).To(Equal(expectedUnits))
				Expect(resultValue).To(BeNil())
			},
			Entry("has mmol/l units", "mmol/l", "mmol/L"),
			Entry("has mmol/L units", "mmol/L", "mmol/L"),
			Entry("has mg/dl units", "mg/dl", "mmol/L"),
			Entry("has mg/dL units", "mg/dL", "mmol/L"),
		)

		DescribeTable("and value is not nil",
			func(units string, value float64, expectedUnits string, expectedValue float64) {
				standard := normalizer.NewStandardBloodGlucose(&units)
				Expect(standard).ToNot(BeNil())
				Expect(*standard.Units()).To(Equal(expectedUnits))
				Expect(*standard.Value(&value)).To(Equal(expectedValue))
				resultUnits, resultValue := standard.UnitsAndValue(&value)
				Expect(*resultUnits).To(Equal(expectedUnits))
				Expect(*resultValue).To(Equal(expectedValue))
			},
			Entry("has mmol/l units", "mmol/l", 12.345, "mmol/L", 12.345),
			Entry("has mmol/L units", "mmol/L", 12.345, "mmol/L", 12.345),
			Entry("has mg/dl units", "mg/dl", 123.45, "mmol/L", 6.852398394945712),
			Entry("has mg/dL units", "mg/dL", 123.45, "mmol/L", 6.852398394945712),
		)
	})
})
