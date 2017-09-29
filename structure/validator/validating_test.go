package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Validating", func() {
	var base *testStructure.Base
	var validatable *testStructure.Validatable

	BeforeEach(func() {
		base = testStructure.NewBase()
		validatable = testStructure.NewValidatable()
	})

	AfterEach(func() {
		validatable.Expectations()
		base.Expectations()
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
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value does not exist"))
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
				Expect(base.ReportErrorInputs).To(BeEmpty())
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
				Expect(base.ReportErrorInputs).To(BeEmpty())
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
				Expect(base.ReportErrorInputs).To(BeEmpty())
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
				Expect(base.ReportErrorInputs).To(HaveLen(1))
				Expect(base.ReportErrorInputs[0]).To(MatchError("value exists"))
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
				Expect(base.ReportErrorInputs).To(BeEmpty())
				Expect(validatable.ValidateInvocations).To(Equal(1))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})
