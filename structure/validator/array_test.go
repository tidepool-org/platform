package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Array", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewArray", func() {
		It("returns successfully", func() {
			value := []interface{}{"a", "b"}
			Expect(structureValidator.NewArray(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.Array
		var result structure.Array

		BeforeEach(func() {
			validator = structureValidator.NewArray(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with empty value", func() {
		var validator *structureValidator.Array
		var result structure.Array
		var value []interface{}

		BeforeEach(func() {
			value = []interface{}{}
			validator = structureValidator.NewArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with non-empty value", func() {
		var validator *structureValidator.Array
		var result structure.Array
		var value []interface{}

		BeforeEach(func() {
			value = []interface{}{"a", "b"}
			validator = structureValidator.NewArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
