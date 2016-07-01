package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("StandardInteger", func() {
	It("NewStandardInteger returns nil if context is nil", func() {
		value := 1
		Expect(validator.NewStandardInteger(nil, "ghast", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(log.NewNullLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardInteger *validator.StandardInteger
			var result data.Integer

			BeforeEach(func() {
				standardInteger = validator.NewStandardInteger(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardInteger).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardInteger.Exists()
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
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardInteger.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.EqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.NotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardInteger.LessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.LessThan(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardInteger.InRange(0, 1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardInteger.OneOf([]int{1, 7})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardInteger.NotOneOf([]int{7, 4})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})
		})

		Context("new validator with valid reference and value of 1", func() {
			var standardInteger *validator.StandardInteger
			var result data.Integer

			BeforeEach(func() {
				value := 1
				standardInteger = validator.NewStandardInteger(standardContext, "ghast", &value)
			})

			It("exists", func() {
				Expect(standardInteger).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardInteger.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardInteger.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.EqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.NotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardInteger.LessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.LessThanOrEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 1 is not greater than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThanOrEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 1 is not greater than or equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardInteger.InRange(0, 3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardInteger.OneOf([]int{1, 7})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardInteger.NotOneOf([]int{7, 4})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})
		})

		Context("new validator with valid reference and value of 4", func() {
			var standardInteger *validator.StandardInteger
			var result data.Integer

			BeforeEach(func() {
				value := 4
				standardInteger = validator.NewStandardInteger(standardContext, "ghast", &value)
			})

			It("exists", func() {
				Expect(standardInteger).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardInteger.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardInteger.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.EqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.NotEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThan", func() {
				BeforeEach(func() {
					result = standardInteger.LessThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not less than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("LessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.LessThanOrEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not less than or equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThan", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("GreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardInteger.GreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("InRange", func() {
				BeforeEach(func() {
					result = standardInteger.InRange(0, 3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not between 0 and 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardInteger.OneOf([]int{1, 7})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not one of [1, 7]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardInteger.NotOneOf([]int{7, 4})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-disallowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is one of the disallowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is one of [7, 4]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("OneOf with nil allowed values", func() {
				BeforeEach(func() {
					result = standardInteger.OneOf(nil)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value 4 is not one of []"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/ghast"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})

			Context("NotOneOf with nil disallowed values", func() {
				BeforeEach(func() {
					result = standardInteger.NotOneOf(nil)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardInteger))
				})
			})
		})
	})
})
