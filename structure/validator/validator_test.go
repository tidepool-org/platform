package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Validator", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("New", func() {
		It("returns successfully", func() {
			Expect(structureValidator.New()).ToNot(BeNil())
		})
	})

	Context("NewValidator", func() {
		It("returns successfully", func() {
			Expect(structureValidator.NewValidator(base)).ToNot(BeNil())
		})
	})

	Context("with new validator", func() {
		var validator *structureValidator.Validator

		BeforeEach(func() {
			validator = structureValidator.NewValidator(base)
			Expect(validator).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(validator.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Validate", func() {
			var validatable *testStructure.Validatable

			BeforeEach(func() {
				validatable = testStructure.NewValidatable()
			})

			AfterEach(func() {
				validatable.Expectations()
			})

			It("invokes normalize and returns current errors", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(validator.Validate(validatable)).To(Equal(errors.Normalize(err)))
				Expect(validatable.ValidateInputs).To(Equal([]structure.Validator{validator}))
			})
		})

		Context("Validating", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Validating("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := testStructure.NewValidatable()
				Expect(validator.Validating("reference", value)).ToNot(BeNil())
			})
		})

		Context("Bool", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Bool("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := true
				Expect(validator.Bool("reference", &value)).ToNot(BeNil())
			})
		})

		Context("Float64", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Float64("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := 3.45
				Expect(validator.Float64("reference", &value)).ToNot(BeNil())
			})
		})

		Context("Int", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Int("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := 2
				Expect(validator.Int("reference", &value)).ToNot(BeNil())
			})
		})

		Context("String", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.String("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := "six"
				Expect(validator.String("reference", &value)).ToNot(BeNil())
			})
		})

		Context("StringArray", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.StringArray("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := []string{"seven", "eight"}
				Expect(validator.StringArray("reference", &value)).ToNot(BeNil())
			})
		})

		Context("Time", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Time("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := time.Now()
				Expect(validator.Time("reference", &value)).ToNot(BeNil())
			})
		})

		Context("WithSource", func() {
			It("returns new validator", func() {
				src := testStructure.NewSource()
				result := validator.WithSource(src)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(validator))
			})
		})

		Context("WithMeta", func() {
			It("returns new validator", func() {
				meta := testErrors.NewMeta()
				result := validator.WithMeta(meta)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(validator))
			})
		})

		Context("WithReference", func() {
			It("without source returns new validator", func() {
				reference := testStructure.NewReference()
				result := validator.WithReference(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(validator))
				Expect(result).To(Equal(validator))
			})

			It("with source returns new validator", func() {
				src := testStructure.NewSource()
				src.WithReferenceOutputs = []structure.Source{testStructure.NewSource()}
				reference := testStructure.NewReference()
				resultWithSource := validator.WithSource(src)
				resultWithReference := validator.WithReference(reference)
				Expect(resultWithReference).ToNot(BeNil())
				Expect(resultWithReference).ToNot(Equal(resultWithSource))
			})
		})
	})
})
