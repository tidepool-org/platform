package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("StandardStringArray", func() {
	It("NewStandardStringArray returns nil if context is nil", func() {
		value := []string{}
		Expect(validator.NewStandardStringArray(nil, "wight", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(test.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardStringArray *validator.StandardStringArray
			var result data.StringArray

			BeforeEach(func() {
				standardStringArray = validator.NewStandardStringArray(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardStringArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringArray.Exists()
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
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringArray.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardStringArray.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardStringArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThan(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthInRange(0, 1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachOneOf([]string{"1", "seven"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachNotOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachNotOneOf([]string{"seven", "four"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})
		})

		Context("new validator with valid reference and empty string array value", func() {
			var standardStringArray *validator.StandardStringArray
			var result data.StringArray

			BeforeEach(func() {
				value := []string{}
				standardStringArray = validator.NewStandardStringArray(standardContext, "wight", &value)
			})

			It("exists", func() {
				Expect(standardStringArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardStringArray.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardStringArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})
		})

		Context("new validator with valid reference and value with length of 1", func() {
			var standardStringArray *validator.StandardStringArray
			var result data.StringArray

			BeforeEach(func() {
				value := []string{"1"}
				standardStringArray = validator.NewStandardStringArray(standardContext, "wight", &value)
			})

			It("exists", func() {
				Expect(standardStringArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardStringArray.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardStringArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThanOrEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThanOrEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than or equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthInRange(0, 3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachOneOf([]string{"1", "seven"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachNotOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachNotOneOf([]string{"seven", "four"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})
		})

		Context("new validator with valid reference and value with length of 4", func() {
			var standardStringArray *validator.StandardStringArray
			var result data.StringArray

			BeforeEach(func() {
				value := []string{"1", "two", "three", "four"}
				standardStringArray = validator.NewStandardStringArray(standardContext, "wight", &value)
			})

			It("exists", func() {
				Expect(standardStringArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardStringArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardStringArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardStringArray.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardStringArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthNotEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthLessThanOrEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than or equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardStringArray.LengthInRange(0, 3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not between 0 and 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachOneOf([]string{"1", "seven"})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(3))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"two\" is not one of [\"1\", \"seven\"]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight/1"))
					Expect(standardContext.Errors()[1]).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[1].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[1].Detail).To(Equal("Value \"three\" is not one of [\"1\", \"seven\"]"))
					Expect(standardContext.Errors()[1].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Source.Pointer).To(Equal("/wight/2"))
					Expect(standardContext.Errors()[2]).ToNot(BeNil())
					Expect(standardContext.Errors()[2].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[2].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[2].Detail).To(Equal("Value \"four\" is not one of [\"1\", \"seven\"]"))
					Expect(standardContext.Errors()[2].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[2].Source.Pointer).To(Equal("/wight/3"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachNotOneOf", func() {
				BeforeEach(func() {
					result = standardStringArray.EachNotOneOf([]string{"seven", "four"})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-disallowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is one of the disallowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"four\" is one of [\"seven\", \"four\"]"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight/3"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachOneOf with nil allowed values", func() {
				BeforeEach(func() {
					result = standardStringArray.EachOneOf(nil)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(4))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value \"1\" is not one of []"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/wight/0"))
					Expect(standardContext.Errors()[1]).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[1].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[1].Detail).To(Equal("Value \"two\" is not one of []"))
					Expect(standardContext.Errors()[1].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[1].Source.Pointer).To(Equal("/wight/1"))
					Expect(standardContext.Errors()[2]).ToNot(BeNil())
					Expect(standardContext.Errors()[2].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[2].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[2].Detail).To(Equal("Value \"three\" is not one of []"))
					Expect(standardContext.Errors()[2].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[2].Source.Pointer).To(Equal("/wight/2"))
					Expect(standardContext.Errors()[3]).ToNot(BeNil())
					Expect(standardContext.Errors()[3].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[3].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[3].Detail).To(Equal("Value \"four\" is not one of []"))
					Expect(standardContext.Errors()[3].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[3].Source.Pointer).To(Equal("/wight/3"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})

			Context("EachNotOneOf with nil disallowed values", func() {
				BeforeEach(func() {
					result = standardStringArray.EachNotOneOf(nil)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardStringArray))
				})
			})
		})
	})
})
