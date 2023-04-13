package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
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

		Context("Origin", func() {
			It("returns OriginExternal if default", func() {
				Expect(validator.Origin()).To(Equal(structure.OriginExternal))
			})

			It("returns set origin", func() {
				Expect(validator.WithOrigin(structure.OriginInternal).Origin()).To(Equal(structure.OriginInternal))
			})
		})

		Context("HasSource", func() {
			It("returns false if no source set", func() {
				Expect(validator.WithSource(nil).HasSource()).To(BeFalse())
			})

			It("returns true if source set", func() {
				Expect(validator.WithSource(structureTest.NewSource()).HasSource()).To(BeTrue())
			})
		})

		Context("Source", func() {
			It("returns default source", func() {
				Expect(validator.Source()).To(BeNil())
			})

			It("returns set source", func() {
				src := structureTest.NewSource()
				Expect(validator.WithSource(src).Source()).To(Equal(src))
			})
		})

		Context("HasMeta", func() {
			It("returns false if no meta set", func() {
				Expect(validator.WithMeta(nil).HasMeta()).To(BeFalse())
			})

			It("returns true if meta set", func() {
				Expect(validator.WithMeta(errorsTest.NewMeta()).HasMeta()).To(BeTrue())
			})
		})

		Context("Meta", func() {
			It("returns default meta", func() {
				Expect(validator.Meta()).To(BeNil())
			})

			It("returns set meta", func() {
				meta := errorsTest.NewMeta()
				Expect(validator.WithMeta(meta).Meta()).To(Equal(meta))
			})
		})

		Context("HasError", func() {
			It("returns false if no errors reported", func() {
				Expect(validator.HasError()).To(BeFalse())
			})

			It("returns true if any errors reported", func() {
				base.ReportError(errorsTest.RandomError())
				Expect(validator.HasError()).To(BeTrue())
			})
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := errorsTest.RandomError()
				base.ReportError(err)
				Expect(validator.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("ReportError", func() {
			It("reports the error to the base", func() {
				err := errorsTest.RandomError()
				validator.ReportError(err)
				Expect(base.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Validate", func() {
			var validatable *structureTest.Validatable

			BeforeEach(func() {
				validatable = structureTest.NewValidatable()
			})

			AfterEach(func() {
				validatable.Expectations()
			})

			It("invokes normalize and returns current errors", func() {
				err := errorsTest.RandomError()
				base.ReportError(err)
				Expect(validator.Validate(validatable)).To(Equal(errors.Normalize(err)))
				Expect(validatable.ValidateInputs).To(Equal([]structure.Validator{validator}))
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
				value := test.RandomTime()
				Expect(validator.Time("reference", &value)).ToNot(BeNil())
			})
		})

		Context("Object", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Object("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := map[string]interface{}{"a": 1, "b": 2}
				Expect(validator.Object("reference", &value)).ToNot(BeNil())
			})
		})

		Context("Array", func() {
			It("returns a validator when called with nil value", func() {
				Expect(validator.Array("reference", nil)).ToNot(BeNil())
			})

			It("returns a validator when called with non-nil value", func() {
				value := []interface{}{"a", "b"}
				Expect(validator.Array("reference", &value)).ToNot(BeNil())
			})
		})

		Context("WithOrigin", func() {
			It("returns a new validator with origin", func() {
				result := validator.WithOrigin(structure.OriginInternal)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(validator))
				Expect(result.Error()).ToNot(HaveOccurred())
				Expect(result.Origin()).To(Equal(structure.OriginInternal))
				Expect(validator.Origin()).To(Equal(structure.OriginExternal))
			})
		})

		Context("WithSource", func() {
			It("returns new validator", func() {
				src := structureTest.NewSource()
				result := validator.WithSource(src)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(validator))
			})
		})

		Context("WithMeta", func() {
			It("returns new validator", func() {
				result := validator.WithMeta(errorsTest.NewMeta())
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(validator))
			})
		})

		Context("WithReference", func() {
			It("without source returns new validator", func() {
				reference := structureTest.NewReference()
				result := validator.WithReference(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(validator))
				Expect(result).To(Equal(validator))
			})

			It("with source returns new validator", func() {
				src := structureTest.NewSource()
				src.WithReferenceOutputs = []structure.Source{structureTest.NewSource()}
				reference := structureTest.NewReference()
				resultWithSource := validator.WithSource(src)
				resultWithReference := validator.WithReference(reference)
				Expect(resultWithReference).ToNot(BeNil())
				Expect(resultWithReference).ToNot(Equal(resultWithSource))
			})
		})
	})
})
