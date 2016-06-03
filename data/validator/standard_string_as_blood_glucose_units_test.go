package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("StandardStringAsBloodGlucoseUnits", func() {
	It("New returns nil if context is nil", func() {
		value := "mg/dL"
		Expect(validator.NewStandardStringAsBloodGlucoseUnits(nil, "necromancer", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(standardContext).ToNot(BeNil())
			Expect(err).To(Succeed())
		})

		Context("new validator with nil reference and nil units", func() {
			var standard *validator.StandardStringAsBloodGlucoseUnits
			var result data.BloodGlucoseUnits

			BeforeEach(func() {
				standard = validator.NewStandardStringAsBloodGlucoseUnits(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-does-not-exist"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value does not exist"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value does not exist"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		Context("new validator with valid reference and an unknown units", func() {
			var standard *validator.StandardStringAsBloodGlucoseUnits
			var result data.BloodGlucoseUnits

			BeforeEach(func() {
				value := "unknown"
				standard = validator.NewStandardStringAsBloodGlucoseUnits(standardContext, "necromancer", &value)
			})

			It("exists", func() {
				Expect(standard).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standard.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"unknown\" is not one of [\"mmol/l\", \"mmol/L\", \"mg/dl\", \"mg/dL\"]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/necromancer"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standard))
				})
			})
		})

		DescribeTable("new validator with valid reference does not add an error when",
			func(units string) {
				standard := validator.NewStandardStringAsBloodGlucoseUnits(standardContext, "necromancer", &units)
				Expect(standard).ToNot(BeNil())
				Expect(standard.Exists()).To(BeIdenticalTo(standard))
				Expect(standardContext.Errors()).To(BeEmpty())
			},
			Entry("has mmol/l units", "mmol/l"),
			Entry("has mmol/L units", "mmol/L"),
			Entry("has mg/dl units", "mg/dl"),
			Entry("has mg/dL units", "mg/dL"),
		)
	})
})
