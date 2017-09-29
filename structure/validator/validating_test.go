package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Validating", func() {
	var base *structureBase.Base
	var validatable *testStructure.Validatable

	BeforeEach(func() {
		base = structureBase.New()
		validatable = testStructure.NewValidatable()
	})

	AfterEach(func() {
		validatable.Expectations()
	})

	Context("NewValidating", func() {
		It("returns successfully", func() {
			Expect(structureValidator.NewValidating(base, validatable)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.Validating
		var result structure.Validating

		BeforeEach(func() {
			validator = structureValidator.NewValidating(base, nil)
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

		Context("Validate", func() {
			BeforeEach(func() {
				result = validator.Validate()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with validator", func() {
		var validator *structureValidator.Validating
		var result structure.Validating

		BeforeEach(func() {
			validator = structureValidator.NewValidating(base, validatable)
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

		Context("Validate", func() {
			BeforeEach(func() {
				result = validator.Validate()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
				Expect(validatable.ValidateInvocations).To(Equal(1))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
