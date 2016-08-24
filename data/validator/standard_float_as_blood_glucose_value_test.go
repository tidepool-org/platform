package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/bloodglucose"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("StandardFloatAsBloodGlucoseValue", func() {
	It("NewStandardFloatAsBloodGlucoseValue returns nil if context is nil", func() {
		value := 12.345
		Expect(validator.NewStandardFloatAsBloodGlucoseValue(nil, "shapeshifter", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardFloatAsBloodGlucoseValue *validator.StandardFloatAsBloodGlucoseValue
			var result data.BloodGlucoseValue

			BeforeEach(func() {
				standardFloatAsBloodGlucoseValue = validator.NewStandardFloatAsBloodGlucoseValue(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardFloatAsBloodGlucoseValue).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.Exists()
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
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.InRange(0.0, 12.345)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRangeForUnits", func() {
				It("does not add an error if it has nil units", func() {
					result = standardFloatAsBloodGlucoseValue.InRangeForUnits(nil)
					Expect(standardContext.Errors()).To(BeEmpty())
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})

				DescribeTable("does not add an error when",
					func(units string) {
						result = standardFloatAsBloodGlucoseValue.InRangeForUnits(&units)
						Expect(standardContext.Errors()).To(BeEmpty())
						Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
					},
					Entry("has unknown units", "unknown"),
					Entry("has mmol/l units", "mmol/l"),
					Entry("has mmol/L units", "mmol/L"),
					Entry("has mg/dl units", "mg/dl"),
					Entry("has mg/dL units", "mg/dL"),
				)
			})
		})

		Context("new validator with valid reference and value of 12.345", func() {
			var standardFloatAsBloodGlucoseValue *validator.StandardFloatAsBloodGlucoseValue
			var result data.BloodGlucoseValue

			BeforeEach(func() {
				value := 12.345
				standardFloatAsBloodGlucoseValue = validator.NewStandardFloatAsBloodGlucoseValue(standardContext, "shapeshifter", &value)
			})

			It("exists", func() {
				Expect(standardFloatAsBloodGlucoseValue).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/shapeshifter"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.InRange(0.0, 30.0)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRangeForUnits", func() {
				It("does not add an error if it has nil units", func() {
					result = standardFloatAsBloodGlucoseValue.InRangeForUnits(nil)
					Expect(standardContext.Errors()).To(BeEmpty())
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})

				DescribeTable("does not add an error when",
					func(units string) {
						result = standardFloatAsBloodGlucoseValue.InRangeForUnits(&units)
						Expect(standardContext.Errors()).To(BeEmpty())
						Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
					},
					Entry("has unknown units", "unknown"),
					Entry("has mmol/l units", "mmol/l"),
					Entry("has mmol/L units", "mmol/L"),
					Entry("has mg/dl units", "mg/dl"),
					Entry("has mg/dL units", "mg/dL"),
				)
			})
		})

		Context("new validator with valid reference and value of 4567.8", func() {
			var standardFloatAsBloodGlucoseValue *validator.StandardFloatAsBloodGlucoseValue
			var result data.BloodGlucoseValue

			BeforeEach(func() {
				value := 4567.8
				standardFloatAsBloodGlucoseValue = validator.NewStandardFloatAsBloodGlucoseValue(standardContext, "shapeshifter", &value)
			})

			It("exists", func() {
				Expect(standardFloatAsBloodGlucoseValue).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/shapeshifter"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloatAsBloodGlucoseValue.InRange(0.0, 30.0)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4567.8 is not between 0 and 30"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/shapeshifter"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})
			})

			Context("InRangeForUnits", func() {
				It("does not add an error if it has nil units", func() {
					result = standardFloatAsBloodGlucoseValue.InRangeForUnits(nil)
					Expect(standardContext.Errors()).To(BeEmpty())
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})

				It("does not add an error if it has unknown units", func() {
					units := "unknown"
					result = standardFloatAsBloodGlucoseValue.InRangeForUnits(&units)
					Expect(standardContext.Errors()).To(BeEmpty())
					Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
				})

				DescribeTable("adds an error when",
					func(units string, upperLimit float64) {
						result = standardFloatAsBloodGlucoseValue.InRangeForUnits(&units)
						Expect(standardContext.Errors()).To(HaveLen(1))
						Expect(standardContext.Errors()[0]).ToNot(BeNil())
						Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
						Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
						Expect(standardContext.Errors()[0].Detail).To(Equal(fmt.Sprintf("Value 4567.8 is not between 0 and %d", int(upperLimit))))
						Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
						Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/shapeshifter"))
						Expect(result).To(BeIdenticalTo(standardFloatAsBloodGlucoseValue))
					},
					Entry("has mmol/l units", "mmol/l", bloodglucose.MmolLUpperLimit),
					Entry("has mmol/L units", "mmol/L", bloodglucose.MmolLUpperLimit),
					Entry("has mg/dl units", "mg/dl", bloodglucose.MgdLUpperLimit),
					Entry("has mg/dL units", "mg/dL", bloodglucose.MgdLUpperLimit),
				)
			})
		})
	})
})
