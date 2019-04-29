package normalizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	testStructure "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Normalizer", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(dataNormalizer.New()).ToNot(BeNil())
		})
	})

	Context("with new normalizer", func() {
		var normalizer *dataNormalizer.Normalizer

		BeforeEach(func() {
			normalizer = dataNormalizer.New()
			Expect(normalizer).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns no error", func() {
				Expect(normalizer.Error()).ToNot(HaveOccurred())
			})

			It("returns any reported error", func() {
				err := testErrors.RandomError()
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
				err := testErrors.RandomError()
				normalizer.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Normalize", func() {
			var normalizable *testData.Normalizable

			BeforeEach(func() {
				normalizable = testData.NewNormalizable()
			})

			AfterEach(func() {
				normalizable.Expectations()
			})

			It("invokes normalize", func() {
				Expect(normalizer.Normalize(normalizable)).To(Succeed())
				Expect(normalizable.NormalizeInputs).To(Equal([]data.Normalizer{normalizer}))
			})

			It("returns any error", func() {
				err := testErrors.RandomError()
				normalizable.NormalizeStub = func(normalizer data.Normalizer) { normalizer.ReportError(err) }
				Expect(normalizer.Normalize(normalizable)).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Data", func() {
			It("returns no data", func() {
				Expect(normalizer.Data()).To(BeEmpty())
			})

			It("returns data", func() {
				datum1 := testData.NewDatum()
				datum2 := testData.NewDatum()
				normalizer.AddData(datum1, datum2)
				Expect(normalizer.Data()).To(Equal([]data.Datum{datum1, datum2}))
			})
		})

		Context("AddData", func() {
			It("does nothing if data is nil", func() {
				normalizer.AddData(nil, nil)
				Expect(normalizer.Data()).To(BeEmpty())
			})

			It("adds data", func() {
				datum1 := testData.NewDatum()
				datum2 := testData.NewDatum()
				normalizer.AddData(datum1, datum2)
				Expect(normalizer.Data()).To(Equal([]data.Datum{datum1, datum2}))
			})
		})

		Context("WithSource", func() {
			var source *testStructure.Source
			var normalizerWithSource data.Normalizer

			BeforeEach(func() {
				source = testStructure.NewSource()
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
				err := testErrors.RandomError()
				normalizerWithSource.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithSource(err, source)))
			})
		})

		Context("WithMeta", func() {
			var meta interface{}
			var normalizerWithMeta data.Normalizer

			BeforeEach(func() {
				meta = testErrors.NewMeta()
				normalizerWithMeta = normalizer.WithMeta(meta)
			})

			It("returns new normalizer", func() {
				Expect(normalizerWithMeta).ToNot(BeNil())
				Expect(normalizerWithMeta).ToNot(Equal(normalizer))
			})

			It("retains the meta", func() {
				err := testErrors.RandomError()
				normalizerWithMeta.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithMeta(err, meta)))
			})
		})

		Context("WithReference", func() {
			var reference string
			var normalizerWithReference data.Normalizer

			BeforeEach(func() {
				reference = testStructure.NewReference()
				normalizerWithReference = normalizer.WithReference(reference)
			})

			It("returns new normalizer", func() {
				Expect(normalizerWithReference).ToNot(BeNil())
				Expect(normalizerWithReference).ToNot(Equal(normalizer))
			})

			It("retains the reference", func() {
				err := testErrors.RandomError()
				source := testStructure.NewSource()
				source.ParameterOutput = pointer.FromString("")
				source.PointerOutput = pointer.FromString(fmt.Sprintf("/%s", reference))
				normalizerWithReference.ReportError(err)
				Expect(normalizer.Error()).To(Equal(errors.WithSource(err, source)))
			})
		})
	})
})
