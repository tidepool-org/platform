package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/null"
)

var _ = Describe("StandardString", func() {
	It("NewStandardString returns nil if context is nil", func() {
		value := ""
		Expect(validator.NewStandardString(nil, "skeleton", &value)).To(BeNil())
	})

	Context("with context", func() {
		var standardContext *context.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
		})

		Context("new validator with nil reference and nil value", func() {
			var standardString *validator.StandardString
			var result data.String

			BeforeEach(func() {
				standardString = validator.NewStandardString(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardString).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardString.Exists()
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
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardString.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardString.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardString.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardString.EqualTo("1")
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.NotEqualTo("four")
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThan(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardString.LengthInRange(0, 1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardString.OneOf([]string{"1", "seven"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardString.NotOneOf([]string{"seven", "four"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})
		})

		Context("new validator with valid reference and empty string value", func() {
			var standardString *validator.StandardString
			var result data.String

			BeforeEach(func() {
				value := ""
				standardString = validator.NewStandardString(standardContext, "skeleton", &value)
			})

			It("exists", func() {
				Expect(standardString).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardString.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardString.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardString.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardString.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})
		})

		Context("new validator with valid reference and value of 1", func() {
			var standardString *validator.StandardString
			var result data.String

			BeforeEach(func() {
				value := "1"
				standardString = validator.NewStandardString(standardContext, "skeleton", &value)
			})

			It("exists", func() {
				Expect(standardString).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardString.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardString.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardString.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardString.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardString.EqualTo("1")
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.NotEqualTo("four")
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThanOrEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThanOrEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than or equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardString.LengthInRange(0, 3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardString.OneOf([]string{"1", "seven"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardString.NotOneOf([]string{"seven", "four"})
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})
		})

		Context("new validator with valid reference and value with length of four", func() {
			var standardString *validator.StandardString
			var result data.String

			BeforeEach(func() {
				value := "four"
				standardString = validator.NewStandardString(standardContext, "skeleton", &value)
			})

			It("exists", func() {
				Expect(standardString).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardString.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardString.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardString.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardString.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("EqualTo", func() {
				BeforeEach(func() {
					result = standardString.EqualTo("1")
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal(`Value "four" is not equal to "1"`))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.NotEqualTo("four")
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal(`Value "four" is equal to "four"`))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthNotEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthLessThanOrEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than or equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardString.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardString.LengthInRange(0, 3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not between 0 and 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("OneOf", func() {
				BeforeEach(func() {
					result = standardString.OneOf([]string{"1", "seven"})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal(`Value "four" is not one of ["1", "seven"]`))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotOneOf", func() {
				BeforeEach(func() {
					result = standardString.NotOneOf([]string{"seven", "four"})
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-disallowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is one of the disallowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal(`Value "four" is one of ["seven", "four"]`))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("OneOf with nil allowed values", func() {
				BeforeEach(func() {
					result = standardString.OneOf(nil)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-allowed"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not one of the allowed values"))
					Expect(standardContext.Errors()[0].Detail).To(Equal(`Value "four" is not one of []`))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/skeleton"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})

			Context("NotOneOf with nil disallowed values", func() {
				BeforeEach(func() {
					result = standardString.NotOneOf(nil)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardString))
				})
			})
		})
	})
})
