package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Bool", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewBool", func() {
		It("returns successfully", func() {
			value := true
			Expect(structureValidator.NewBool(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool

		BeforeEach(func() {
			validator = structureValidator.NewBool(base, nil)
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

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with true value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool
		var value bool

		BeforeEach(func() {
			value = true
			validator = structureValidator.NewBool(base, &value)
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

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotFalse())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with false value", func() {
		var validator *structureValidator.Bool
		var result structure.Bool
		var value bool

		BeforeEach(func() {
			value = false
			validator = structureValidator.NewBool(base, &value)
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

		Context("True", func() {
			BeforeEach(func() {
				result = validator.True()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotTrue())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("False", func() {
			BeforeEach(func() {
				result = validator.False()
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
