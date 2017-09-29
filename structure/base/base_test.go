package base_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/tidepool-org/platform/errors"
// 	testErrors "github.com/tidepool-org/platform/errors/test"
// 	"github.com/tidepool-org/platform/structure"
// 	structureBase "github.com/tidepool-org/platform/structure/base"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// 	"github.com/tidepool-org/platform/test"
// )

// var _ = Describe("Base", func() {
// 	var source *testStructure.Source
// 	var meta interface{}

// 	BeforeEach(func() {
// 		source = testStructure.NewSource()
// 		meta = test.NewText(1, 128)
// 	})

// 	AfterEach(func() {
// 		source.Expectations()
// 	})

// 	Context("New", func() {
// 		It("returns successfully", func() {
// 			Expect(structureBase.New()).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new base", func() {
// 		var base *structureBase.Base

// 		BeforeEach(func() {
// 			base = structureBase.New()
// 			Expect(base).ToNot(BeNil())
// 		})

// 		Context("Source", func() {
// 			It("returns nil if source not specified", func() {
// 				Expect(base.Source()).To(BeNil())
// 			})

// 			It("returns source if source specified", func() {
// 				Expect(base.WithSource(source).Source()).To(Equal(source))
// 			})
// 		})

// 		Context("Meta", func() {
// 			It("returns nil if meta not specified", func() {
// 				Expect(base.Meta()).To(BeNil())
// 			})

// 			It("returns meta if meta specified", func() {
// 				Expect(base.WithMeta(meta).Meta()).To(Equal(meta))
// 			})
// 		})

// 		Context("Errors", func() {
// 			It("returns nil if no errors", func() {
// 				Expect(base.Errors()).To(BeNil())
// 			})

// 			It("returns errors if any errors", func() {
// 				err1 := errors.ErrorInternal()
// 				base.ReportError(err1)
// 				err2 := errors.ErrorInternal()
// 				base.ReportError(err2)
// 				errs := base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err1, err2))
// 			})
// 		})

// 		Context("ReportError", func() {
// 			It("does not add the error if nil", func() {
// 				base.ReportError(nil)
// 				Expect(base.Errors()).To(BeNil())
// 			})

// 			It("adds the error", func() {
// 				err := errors.ErrorInternal()
// 				base.ReportError(err)
// 				errs := base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err))
// 			})

// 			It("adds the error with source", func() {
// 				errSource := errors.NewSource()
// 				errSource.Parameter = testErrors.NewSourceParameter()
// 				errSource.Pointer = testErrors.NewSourcePointer()
// 				source.SourceOutputs = []*errors.Source{errSource}
// 				err := errors.ErrorInternal()
// 				base.WithSource(source).ReportError(err)
// 				errs := base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err.WithSource(errSource)))
// 			})

// 			It("adds the error with meta", func() {
// 				err := errors.ErrorInternal()
// 				base.WithMeta(meta).ReportError(err)
// 				errs := base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err.WithMeta(meta)))
// 			})

// 			It("adds the error with source and meta", func() {
// 				errSource := errors.NewSource()
// 				errSource.Parameter = testErrors.NewSourceParameter()
// 				errSource.Pointer = testErrors.NewSourcePointer()
// 				source.SourceOutputs = []*errors.Source{errSource}
// 				err := errors.ErrorInternal()
// 				base.WithSource(source).WithMeta(meta).ReportError(err)
// 				errs := base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err.WithSource(errSource).WithMeta(meta)))
// 			})

// 			It("adds the error on a offspring and the ancestor has it", func() {
// 				err := errors.ErrorInternal()
// 				result := base.WithMeta(meta).WithMeta(meta).WithMeta(meta)
// 				result.ReportError(err)
// 				errs := result.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err.WithMeta(meta)))
// 				errs = base.Errors()
// 				Expect(errs).ToNot(BeNil())
// 				Expect(*errs).To(ConsistOf(err.WithMeta(meta)))
// 			})

// 			Context("WithSource", func() {
// 				It("returns a new base with source", func() {
// 					result := base.WithSource(source)
// 					Expect(result).ToNot(BeNil())
// 					Expect(result).ToNot(BeIdenticalTo(base))
// 					Expect(result.Source()).To(Equal(source))
// 					Expect(result.Meta()).To(BeNil())
// 					Expect(result.Errors()).To(BeNil())
// 				})
// 			})

// 			Context("WithMeta", func() {
// 				It("returns a new base with meta", func() {
// 					result := base.WithMeta(meta)
// 					Expect(result).ToNot(BeNil())
// 					Expect(result).ToNot(BeIdenticalTo(base))
// 					Expect(result.Source()).To(BeNil())
// 					Expect(result.Meta()).To(Equal(meta))
// 					Expect(result.Errors()).To(BeNil())
// 				})
// 			})

// 			Context("WithReference", func() {
// 				It("returns a new base without change if no source", func() {
// 					reference := testStructure.NewReference()
// 					result := base.WithReference(reference)
// 					Expect(result).ToNot(BeNil())
// 					Expect(result).To(Equal(base))
// 					Expect(result).ToNot(BeIdenticalTo(base))
// 					Expect(result.Source()).To(BeNil())
// 					Expect(result.Meta()).To(BeNil())
// 					Expect(result.Errors()).To(BeNil())
// 				})

// 				It("returns a new base with new source if source", func() {
// 					newSource := testStructure.NewSource()
// 					source.WithReferenceOutputs = []structure.Source{newSource}
// 					reference := testStructure.NewReference()
// 					result := base.WithSource(source).WithReference(reference)
// 					Expect(result).ToNot(BeNil())
// 					Expect(result).ToNot(Equal(base))
// 					Expect(result.Source()).To(Equal(newSource))
// 					Expect(result.Meta()).To(BeNil())
// 					Expect(result.Errors()).To(BeNil())
// 					Expect(source.WithReferenceInputs).To(Equal([]string{reference}))
// 				})
// 			})
// 		})
// 	})
// })
