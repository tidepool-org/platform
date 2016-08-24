package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("StandardStringAsBloodGlucoseUnits", func() {
	It("NewStandardStringAsBloodGlucoseUnits returns nil if context is nil", func() {
		value := "mg/dL"
		Expect(validator.NewStandardStringAsBloodGlucoseUnits(nil, "necromancer", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil units", func() {
			var standardStringAsBloodGlucoseUnits *validator.StandardStringAsBloodGlucoseUnits
			var result data.BloodGlucoseUnits

			BeforeEach(func() {
				standardStringAsBloodGlucoseUnits = validator.NewStandardStringAsBloodGlucoseUnits(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardStringAsBloodGlucoseUnits).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringAsBloodGlucoseUnits.Exists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value does not exist"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value does not exist"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringAsBloodGlucoseUnits))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringAsBloodGlucoseUnits.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringAsBloodGlucoseUnits))
				})
			})
		})

		Context("new validator with valid reference and an unknown units", func() {
			var standardStringAsBloodGlucoseUnits *validator.StandardStringAsBloodGlucoseUnits
			var result data.BloodGlucoseUnits

			BeforeEach(func() {
				value := "unknown"
				standardStringAsBloodGlucoseUnits = validator.NewStandardStringAsBloodGlucoseUnits(standardContext, "necromancer", &value)
			})

			It("exists", func() {
				Expect(standardStringAsBloodGlucoseUnits).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringAsBloodGlucoseUnits.Exists()
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
					Expect(result).To(BeIdenticalTo(standardStringAsBloodGlucoseUnits))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringAsBloodGlucoseUnits.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(2))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"unknown\" is not one of [\"mmol/l\", \"mmol/L\", \"mg/dl\", \"mg/dL\"]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/necromancer"))
					Expect(standardContext.Errors()[1]).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[1].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[1].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[1].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Source.Pointer).To(Equal("/necromancer"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringAsBloodGlucoseUnits))
				})
			})
		})

		DescribeTable("new validator with valid reference does not add an error when",
			func(units string) {
				standardStringAsBloodGlucoseUnits := validator.NewStandardStringAsBloodGlucoseUnits(standardContext, "necromancer", &units)
				Expect(standardStringAsBloodGlucoseUnits).ToNot(BeNil())
				Expect(standardStringAsBloodGlucoseUnits.Exists()).To(BeIdenticalTo(standardStringAsBloodGlucoseUnits))
				Expect(standardContext.Errors()).To(BeEmpty())
			},
			Entry("has mmol/l units", "mmol/l"),
			Entry("has mmol/L units", "mmol/L"),
			Entry("has mg/dl units", "mg/dl"),
			Entry("has mg/dL units", "mg/dL"),
		)
	})
})
