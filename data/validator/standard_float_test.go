package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("StandardFloat", func() {
	It("NewStandardFloat returns nil if context is nil", func() {
		value := 1.23
		Expect(validator.NewStandardFloat(nil, "vampyre", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(standardContext).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardFloat *validator.StandardFloat
			var result data.Float

			BeforeEach(func() {
				standardFloat = validator.NewStandardFloat(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardFloat).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloat.Exists()
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
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloat.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.EqualTo(1.2)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.NotEqualTo(4.5)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardFloat.LessThan(3.)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.LessThan(1.2)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThan(3.)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThanOrEqualTo(4.5)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloat.InRange(0.0, 1.2)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardFloat.OneOf([]float64{1.2, 7.8})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardFloat.NotOneOf([]float64{7.8, 4.5})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})
		})

		Context("new validator with valid reference and value of 1.2", func() {
			var standardFloat *validator.StandardFloat
			var result data.Float

			BeforeEach(func() {
				value := 1.2
				standardFloat = validator.NewStandardFloat(standardContext, "vampyre", &value)
			})

			It("exists", func() {
				Expect(standardFloat).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloat.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloat.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.EqualTo(1.2)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.NotEqualTo(4.5)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardFloat.LessThan(3.)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.LessThanOrEqualTo(1.2)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThan(3.)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 1.2 is not greater than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThanOrEqualTo(4.5)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 1.2 is not greater than or equal to 4.5"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloat.InRange(0.0, 3.0)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardFloat.OneOf([]float64{1.2, 7.8})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardFloat.NotOneOf([]float64{7.8, 4.5})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})
		})

		Context("new validator with valid reference and value of 4.5", func() {
			var standardFloat *validator.StandardFloat
			var result data.Float

			BeforeEach(func() {
				value := 4.5
				standardFloat = validator.NewStandardFloat(standardContext, "vampyre", &value)
			})

			It("exists", func() {
				Expect(standardFloat).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardFloat.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardFloat.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.EqualTo(1.2)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not equal to 1.2"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.NotEqualTo(4.5)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is equal to 4.5"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardFloat.LessThan(3.)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not less than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.LessThanOrEqualTo(1.2)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not less than or equal to 1.2"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThan(3.)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardFloat.GreaterThanOrEqualTo(4.5)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardFloat.InRange(0.0, 3.0)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not between 0 and 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardFloat.OneOf([]float64{1.2, 7.8})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not one of [1.2, 7.8]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardFloat.NotOneOf([]float64{7.8, 4.5})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-disallowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is one of the disallowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is one of [7.8, 4.5]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("OneOf with nil allowed values", func() {
				BeforeEach(func() {
					result = standardFloat.OneOf(nil)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4.5 is not one of []"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/vampyre"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})

			Context("NotOneOf with nil disallowed values", func() {
				BeforeEach(func() {
					result = standardFloat.NotOneOf(nil)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardFloat))
				})
			})
		})
	})
})
