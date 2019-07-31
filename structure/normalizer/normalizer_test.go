package normalizer_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureTest "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Normalizer", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(structureNormalizer.New()).ToNot(BeNil())
		})
	})

	Context("NewNormalizer", func() {
		It("returns successfully", func() {
			Expect(structureNormalizer.NewNormalizer(structureBase.New())).ToNot(BeNil())
		})
	})

	Context("with new normalizer", func() {
		var normalizer *structureNormalizer.Normalizer

		BeforeEach(func() {
			normalizer = structureNormalizer.New()
			Expect(normalizer).ToNot(BeNil())
		})

		Context("Origin", func() {
			It("returns OriginExternal if default", func() {
				Expect(normalizer.Origin()).To(Equal(structure.OriginExternal))
			})

			It("returns set origin", func() {
				Expect(normalizer.WithOrigin(structure.OriginInternal).Origin()).To(Equal(structure.OriginInternal))
			})
		})

		Context("HasSource", func() {
			It("returns false if no source set", func() {
				Expect(normalizer.WithSource(nil).HasSource()).To(BeFalse())
			})

			It("returns true if source set", func() {
				Expect(normalizer.WithSource(structureTest.NewSource()).HasSource()).To(BeTrue())
			})
		})

		Context("Source", func() {
			It("returns default source", func() {
				Expect(normalizer.Source()).To(Equal(structure.NewPointerSource()))
			})

			It("returns set source", func() {
				src := structureTest.NewSource()
				Expect(normalizer.WithSource(src).Source()).To(Equal(src))
			})
		})

		Context("HasMeta", func() {
			It("returns false if no meta set", func() {
				Expect(normalizer.WithMeta(nil).HasMeta()).To(BeFalse())
			})

			It("returns true if meta set", func() {
				Expect(normalizer.WithMeta(errorsTest.NewMeta()).HasMeta()).To(BeTrue())
			})
		})

		Context("Meta", func() {
			It("returns default meta", func() {
				Expect(normalizer.Meta()).To(BeNil())
			})

			It("returns set meta", func() {
				meta := errorsTest.NewMeta()
				Expect(normalizer.WithMeta(meta).Meta()).To(Equal(meta))
			})
		})

		Context("HasError", func() {
			It("returns false if no errors reporter", func() {
				Expect(normalizer.HasError()).To(BeFalse())
			})

			It("returns true if any errors reported", func() {
				normalizer.ReportError(errorsTest.RandomError())
				Expect(normalizer.HasError()).To(BeTrue())
			})
		})

		Context("Error", func() {
			It("returns no error", func() {
				Expect(normalizer.Error()).ToNot(HaveOccurred())
			})

			It("returns any reported error", func() {
				err := errorsTest.RandomError()
				normalizer.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("ReportError", func() {
			It("does not report nil error", func() {
				normalizer.ReportError(nil)
				Expect(normalizer.Error()).ToNot(HaveOccurred())
			})

			It("reports the error", func() {
				err := errorsTest.RandomError()
				normalizer.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Normalize", func() {
			var normalizable *structureTest.Normalizable

			BeforeEach(func() {
				normalizable = structureTest.NewNormalizable()
			})

			AfterEach(func() {
				normalizable.Expectations()
			})

			It("invokes normalize", func() {
				Expect(normalizer.Normalize(normalizable)).To(Succeed())
				Expect(normalizable.NormalizeInputs).To(Equal([]structure.Normalizer{normalizer}))
			})

			It("returns any error", func() {
				err := errorsTest.RandomError()
				normalizable.NormalizeStub = func(normalizer structure.Normalizer) { normalizer.ReportError(err) }
				Expect(normalizer.Normalize(normalizable)).To(Equal(errors.Normalize(err)))
			})
		})

		Context("WithOrigin", func() {
			It("returns a new normalizer with origin", func() {
				result := normalizer.WithOrigin(structure.OriginInternal)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(normalizer))
				Expect(result.Error()).ToNot(HaveOccurred())
				Expect(result.Origin()).To(Equal(structure.OriginInternal))
				Expect(normalizer.Origin()).To(Equal(structure.OriginExternal))
			})
		})

		Context("WithSource", func() {
			var source *structureTest.Source
			var normalizerWithSource structure.Normalizer

			BeforeEach(func() {
				source = structureTest.NewSource()
				normalizerWithSource = normalizer.WithSource(source)
			})

			AfterEach(func() {
				source.Expectations()
			})

			It("returns new normalizer", func() {
				Expect(normalizerWithSource).ToNot(BeNil())
				Expect(normalizerWithSource).ToNot(Equal(normalizer))
			})

			It("retains the source", func() {
				source.ParameterOutput = pointer.FromString("123")
				source.PointerOutput = pointer.FromString("/a/b/c")
				err := errorsTest.RandomError()
				normalizerWithSource.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithSource(err, source)))
			})
		})

		Context("WithMeta", func() {
			var meta interface{}
			var normalizerWithMeta structure.Normalizer

			BeforeEach(func() {
				meta = errorsTest.NewMeta()
				normalizerWithMeta = normalizer.WithMeta(meta)
			})

			It("returns new normalizer", func() {
				Expect(normalizerWithMeta).ToNot(BeNil())
				Expect(normalizerWithMeta).ToNot(Equal(normalizer))
			})

			It("retains the meta", func() {
				err := errorsTest.RandomError()
				normalizerWithMeta.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithMeta(err, meta)))
			})
		})

		Context("WithReference", func() {
			var reference string
			var normalizerWithReference structure.Normalizer

			BeforeEach(func() {
				reference = structureTest.NewReference()
				normalizerWithReference = normalizer.WithReference(reference)
			})

			It("returns new normalizer", func() {
				Expect(normalizerWithReference).ToNot(BeNil())
				Expect(normalizerWithReference).ToNot(Equal(normalizer))
			})

			It("retains the reference", func() {
				err := errorsTest.RandomError()
				source := structureTest.NewSource()
				source.ParameterOutput = pointer.FromString("")
				source.PointerOutput = pointer.FromString(fmt.Sprintf("/%s", reference))
				normalizerWithReference.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithSource(err, source)))
			})
		})
	})
})
