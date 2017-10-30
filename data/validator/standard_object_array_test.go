package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/null"
)

var _ = Describe("StandardObjectArray", func() {
	It("NewStandardObjectArray returns nil if context is nil", func() {
		value := []map[string]interface{}{}
		Expect(validator.NewStandardObjectArray(nil, "mummy", &value)).To(BeNil())
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
			var standardObjectArray *validator.StandardObjectArray
			var result data.ObjectArray

			BeforeEach(func() {
				standardObjectArray = validator.NewStandardObjectArray(standardContext, nil, nil)
			})

			It("exists", func() {
				Expect(standardObjectArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObjectArray.Exists()
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
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotExists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObjectArray.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThan(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthInRange(0, 1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})
		})

		Context("new validator with valid reference and empty object array value", func() {
			var standardObjectArray *validator.StandardObjectArray
			var result data.ObjectArray

			BeforeEach(func() {
				value := []map[string]interface{}{}
				standardObjectArray = validator.NewStandardObjectArray(standardContext, "mummy", &value)
			})

			It("exists", func() {
				Expect(standardObjectArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObjectArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObjectArray.Empty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})
		})

		Context("new validator with valid reference and value with length of 1", func() {
			var standardObjectArray *validator.StandardObjectArray
			var result data.ObjectArray

			BeforeEach(func() {
				value := []map[string]interface{}{{"one": 1}}
				standardObjectArray = validator.NewStandardObjectArray(standardContext, "mummy", &value)
			})

			It("exists", func() {
				Expect(standardObjectArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObjectArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObjectArray.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthNotEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThanOrEqualTo(1)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThanOrEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 1 is not greater than or equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthInRange(0, 3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})
		})

		Context("new validator with valid reference and value with length of 4", func() {
			var standardObjectArray *validator.StandardObjectArray
			var result data.ObjectArray

			BeforeEach(func() {
				value := []map[string]interface{}{{"one": 1}, {"two": 2}, {"three": 3}, {"four": 4}}
				standardObjectArray = validator.NewStandardObjectArray(standardContext, "mummy", &value)
			})

			It("exists", func() {
				Expect(standardObjectArray).ToNot(BeNil())
			})

			Context("Exists", func() {
				BeforeEach(func() {
					result = standardObjectArray.Exists()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotExists", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotExists()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-exists"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value exists"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value exists"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("Empty", func() {
				BeforeEach(func() {
					result = standardObjectArray.Empty()
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("value-not-empty"))
					Expect(standardContext.Errors()[0].Title).To(Equal("value is not empty"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Value is not empty"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("NotEmpty", func() {
				BeforeEach(func() {
					result = standardObjectArray.NotEmpty()
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthNotEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthNotEqualTo(4)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is equal to 4"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThan(3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthLessThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthLessThanOrEqualTo(1)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not less than or equal to 1"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThan", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThan(3)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthGreaterThanOrEqualTo", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthGreaterThanOrEqualTo(4)
				})

				It("does not add an error", func() {
					Expect(standardContext.Errors()).To(BeEmpty())
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})

			Context("LengthInRange", func() {
				BeforeEach(func() {
					result = standardObjectArray.LengthInRange(0, 3)
				})

				It("adds the expected error", func() {
					Expect(standardContext.Errors()).To(HaveLen(1))
					Expect(standardContext.Errors()[0]).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Code).To(Equal("length-out-of-range"))
					Expect(standardContext.Errors()[0].Title).To(Equal("length is out of range"))
					Expect(standardContext.Errors()[0].Detail).To(Equal("Length 4 is not between 0 and 3"))
					Expect(standardContext.Errors()[0].Source).ToNot(BeNil())
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/mummy"))
				})

				It("returns self", func() {
					Expect(result).To(BeIdenticalTo(standardObjectArray))
				})
			})
		})
	})
})
