package base_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureTest "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Base", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(structureBase.New()).ToNot(BeNil())
		})
	})

	Context("with source, meta, and new base", func() {
		var src *structureTest.Source
		var meta interface{}
		var base *structureBase.Base

		BeforeEach(func() {
			src = structureTest.NewSource()
			src.ParameterOutput = pointer.FromString(errorsTest.NewSourceParameter())
			src.PointerOutput = pointer.FromString(errorsTest.NewSourcePointer())
			meta = errorsTest.NewMeta()
			base = structureBase.New()
			Expect(base).ToNot(BeNil())
		})

		AfterEach(func() {
			src.Expectations()
		})

		Context("Origin", func() {
			It("returns OriginExternal if default", func() {
				Expect(base.Origin()).To(Equal(structure.OriginExternal))
			})

			It("returns set origin", func() {
				Expect(base.WithOrigin(structure.OriginInternal).Origin()).To(Equal(structure.OriginInternal))
			})
		})

		Context("HasSource", func() {
			It("returns false if no source set", func() {
				Expect(base.WithSource(nil).HasSource()).To(BeFalse())
			})

			It("returns true if source set", func() {
				Expect(base.WithSource(src).HasSource()).To(BeTrue())
			})
		})

		Context("Source", func() {
			It("returns default source", func() {
				Expect(base.Source()).To(BeNil())
			})

			It("returns set source", func() {
				Expect(base.WithSource(src).Source()).To(Equal(src))
			})
		})

		Context("HasMeta", func() {
			It("returns false if no meta set", func() {
				Expect(base.WithMeta(nil).HasMeta()).To(BeFalse())
			})

			It("returns true if meta set", func() {
				Expect(base.WithMeta(meta).HasMeta()).To(BeTrue())
			})
		})

		Context("Meta", func() {
			It("returns default meta", func() {
				Expect(base.Meta()).To(BeNil())
			})

			It("returns set meta", func() {
				Expect(base.WithMeta(meta).Meta()).To(Equal(meta))
			})
		})

		Context("HasError", func() {
			It("returns false if no errors reported", func() {
				Expect(base.HasError()).To(BeFalse())
			})

			It("returns true if any errors reported", func() {
				base.ReportError(errorsTest.RandomError())
				Expect(base.HasError()).To(BeTrue())
			})
		})

		Context("Error", func() {
			It("returns nil if no error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns errors if any errors", func() {
				err1 := errorsTest.RandomError()
				base.ReportError(err1)
				err2 := errorsTest.RandomError()
				base.ReportError(err2)
				err3 := errorsTest.RandomError()
				base.ReportError(err3)
				Expect(base.Error()).To(Equal(errors.Append(err1, err2, err3)))
			})
		})

		Context("ReportError", func() {
			It("does not add error if nil", func() {
				base.ReportError(nil)
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("reports the error", func() {
				err := errorsTest.RandomError()
				base.ReportError(err)
				Expect(base.Error()).To(Equal(errors.Append(err)))
			})

			It("reports the error with source", func() {
				err := errorsTest.RandomError()
				base.WithSource(src).ReportError(err)
				Expect(base.Error()).To(Equal(errors.WithSource(err, src)))
			})

			It("reports the error with meta", func() {
				err := errorsTest.RandomError()
				base.WithMeta(meta).ReportError(err)
				Expect(base.Error()).To(Equal(errors.WithMeta(err, meta)))
			})

			It("reports the error with source and meta", func() {
				err := errorsTest.RandomError()
				base.WithSource(src).WithMeta(meta).ReportError(err)
				Expect(base.Error()).To(Equal(errors.WithMeta(errors.WithSource(err, src), meta)))
			})

			It("reports the error on a offspring and the ancestor has it", func() {
				err := errorsTest.RandomError()
				result := base.WithMeta(meta).WithMeta(meta).WithMeta(meta)
				result.ReportError(err)
				Expect(result.Error()).To(Equal(errors.WithMeta(err, meta)))
				Expect(base.Error()).To(Equal(result.Error()))
			})
		})

		Context("WithOrigin", func() {
			It("returns a new base with origin", func() {
				result := base.WithOrigin(structure.OriginInternal)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(base))
				Expect(result.Error()).ToNot(HaveOccurred())
				Expect(result.Origin()).To(Equal(structure.OriginInternal))
				Expect(base.Origin()).To(Equal(structure.OriginExternal))
			})
		})

		Context("WithSource", func() {
			It("returns a new base with source", func() {
				result := base.WithSource(src)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(base))
				Expect(result.Error()).ToNot(HaveOccurred())
			})
		})

		Context("WithMeta", func() {
			It("returns a new base with meta", func() {
				result := base.WithMeta(meta)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(base))
				Expect(result.Error()).ToNot(HaveOccurred())
			})
		})

		Context("WithReference", func() {
			It("returns a new base without change if no source", func() {
				reference := structureTest.NewReference()
				result := base.WithReference(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(base))
				Expect(result).ToNot(BeIdenticalTo(base))
				Expect(result.Error()).ToNot(HaveOccurred())
			})

			It("returns a new base with new source if source", func() {
				src.WithReferenceOutputs = []structure.Source{structureTest.NewSource()}
				reference := structureTest.NewReference()
				result := base.WithSource(src).WithReference(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(base))
				Expect(result.Error()).ToNot(HaveOccurred())
				Expect(src.WithReferenceInputs).To(Equal([]string{reference}))
			})
		})
	})
})
