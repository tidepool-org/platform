package normalizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	testStructure "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Normalizer", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("New", func() {
		It("returns successfully", func() {
			Expect(structureNormalizer.New()).ToNot(BeNil())
		})
	})

	Context("NewNormalizer", func() {
		It("returns successfully", func() {
			Expect(structureNormalizer.NewNormalizer(base)).ToNot(BeNil())
		})
	})

	Context("with new normalizer", func() {
		var normalizer *structureNormalizer.Normalizer

		BeforeEach(func() {
			normalizer = structureNormalizer.NewNormalizer(base)
			Expect(normalizer).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Normalize", func() {
			var normalizable *testStructure.Normalizable

			BeforeEach(func() {
				normalizable = testStructure.NewNormalizable()
			})

			AfterEach(func() {
				normalizable.Expectations()
			})

			It("invokes normalize and returns current errors", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(normalizer.Normalize(normalizable)).To(Equal(errors.Normalize(err)))
				Expect(normalizable.NormalizeInputs).To(Equal([]structure.Normalizer{normalizer}))
			})
		})

		Context("WithSource", func() {
			It("returns new normalizer", func() {
				src := testStructure.NewSource()
				result := normalizer.WithSource(src)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(normalizer))
			})
		})

		Context("WithMeta", func() {
			It("returns new normalizer", func() {
				meta := testErrors.NewMeta()
				result := normalizer.WithMeta(meta)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(normalizer))
			})
		})

		Context("WithReference", func() {
			It("without source returns new normalizer", func() {
				reference := testStructure.NewReference()
				result := normalizer.WithReference(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(normalizer))
				Expect(result).To(Equal(normalizer))
			})

			It("with source returns new normalizer", func() {
				src := testStructure.NewSource()
				src.WithReferenceOutputs = []structure.Source{testStructure.NewSource()}
				reference := testStructure.NewReference()
				resultWithSource := normalizer.WithSource(src)
				resultWithReference := normalizer.WithReference(reference)
				Expect(resultWithReference).ToNot(BeNil())
				Expect(resultWithReference).ToNot(Equal(resultWithSource))
			})
		})
	})
})
