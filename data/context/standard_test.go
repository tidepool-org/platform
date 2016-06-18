package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Standard", func() {
	It("NewStandard returns an error if logger is nil", func() {
		standard, err := context.NewStandard(nil)
		Expect(err).To(MatchError("context: logger is missing"))
		Expect(standard).To(BeNil())
	})

	Context("NewStandard", func() {
		var standard *context.Standard
		var err error

		BeforeEach(func() {
			standard, err = context.NewStandard(test.NewLogger())
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("has a contained Errors that is empty", func() {
			Expect(standard.Errors()).To(BeEmpty())
		})

		It("ignores sending a nil error to AppendError", func() {
			standard.AppendError("ignore", nil)
			Expect(standard.Errors()).To(BeEmpty())
		})

		It("Logger returns a logger", func() {
			Expect(standard.Logger()).ToNot(BeNil())
		})

		Context("SetMeta", func() {
			It("sets the meta on the context", func() {
				meta := "metametameta"
				standard.SetMeta(meta)
				Expect(standard.Meta()).To(BeIdenticalTo(meta))
			})
		})

		Context("ResolveReference", func() {
			It("correctly returns the resolved reference", func() {
				Expect(standard.ResolveReference("reference")).To(Equal("/reference"))
			})
		})

		Context("AppendError with an error with a nil reference", func() {
			var nilReferenceError *service.Error

			BeforeEach(func() {
				nilReferenceError = &service.Error{}
				standard.AppendError(nil, nilReferenceError)
			})

			It("has errors", func() {
				Expect(standard.Errors()).ToNot(BeEmpty())
			})

			It("has the error", func() {
				Expect(standard.Errors()).To(ConsistOf(nilReferenceError))
			})

			It("added the <nil> source pointer", func() {
				Expect(standard.Errors()).To(HaveLen(1))
				Expect(standard.Errors()[0]).ToNot(BeNil())
				Expect(standard.Errors()[0].Source).ToNot(BeNil())
				Expect(standard.Errors()[0].Source.Pointer).To(Equal("/<nil>"))
			})
		})

		Context("AppendError with Meta and an error", func() {
			var meta string
			var firstError *service.Error

			BeforeEach(func() {
				meta = "metamorphize"
				standard.SetMeta(meta)
				firstError = &service.Error{}
				standard.AppendError("first", firstError)
			})

			It("has errors", func() {
				Expect(standard.Errors()).ToNot(BeEmpty())
			})

			It("has the error", func() {
				Expect(standard.Errors()).To(ConsistOf(firstError))
			})

			It("added the error source pointer", func() {
				Expect(standard.Errors()).To(HaveLen(1))
				Expect(standard.Errors()[0]).ToNot(BeNil())
				Expect(standard.Errors()[0].Source).ToNot(BeNil())
				Expect(standard.Errors()[0].Source.Pointer).To(Equal("/first"))
				Expect(standard.Errors()[0].Meta).To(Equal(meta))
			})
		})

		Context("AppendError with a first error", func() {
			var firstError *service.Error

			BeforeEach(func() {
				firstError = &service.Error{}
				standard.AppendError("first", firstError)
			})

			It("has errors", func() {
				Expect(standard.Errors()).ToNot(BeEmpty())
			})

			It("has the error", func() {
				Expect(standard.Errors()).To(ConsistOf(firstError))
			})

			It("added the error source pointer", func() {
				Expect(standard.Errors()).To(HaveLen(1))
				Expect(standard.Errors()[0]).ToNot(BeNil())
				Expect(standard.Errors()[0].Source).ToNot(BeNil())
				Expect(standard.Errors()[0].Source.Pointer).To(Equal("/first"))
			})

			Context("and AppendError with a second error", func() {
				var secondError *service.Error

				BeforeEach(func() {
					secondError = &service.Error{}
					standard.AppendError("second", secondError)
				})

				It("has errors", func() {
					Expect(standard.Errors()).ToNot(BeEmpty())
				})

				It("has both errors", func() {
					Expect(standard.Errors()).To(ConsistOf(firstError, secondError))
				})

				It("added the error source pointer", func() {
					Expect(standard.Errors()).To(HaveLen(2))
					Expect(standard.Errors()[0]).ToNot(BeNil())
					Expect(standard.Errors()[0].Source).ToNot(BeNil())
					Expect(standard.Errors()[0].Source.Pointer).To(Equal("/first"))
					Expect(standard.Errors()[1]).ToNot(BeNil())
					Expect(standard.Errors()[1].Source).ToNot(BeNil())
					Expect(standard.Errors()[1].Source.Pointer).To(Equal("/second"))
				})
			})
		})

		Context("creating a child context", func() {
			var child data.Context

			BeforeEach(func() {
				child = standard.NewChildContext("child")
			})

			It("exists", func() {
				Expect(child).ToNot(BeNil())
			})

			It("Logger returns a logger", func() {
				Expect(child.Logger()).ToNot(BeNil())
			})

			Context("ResolveReference", func() {
				It("correctly returns the resolved reference", func() {
					Expect(child.ResolveReference("reference")).To(Equal("/child/reference"))
				})
			})

			Context("AppendError with a first error", func() {
				var firstError *service.Error

				BeforeEach(func() {
					firstError = &service.Error{}
					child.AppendError("first", firstError)
				})

				It("has errors", func() {
					Expect(standard.Errors()).ToNot(BeEmpty())
				})

				It("has the error", func() {
					Expect(standard.Errors()).To(ConsistOf(firstError))
				})

				It("added the error source pointer", func() {
					Expect(standard.Errors()).To(HaveLen(1))
					Expect(standard.Errors()[0]).ToNot(BeNil())
					Expect(standard.Errors()[0].Source).ToNot(BeNil())
					Expect(standard.Errors()[0].Source.Pointer).To(Equal("/child/first"))
				})

				Context("and AppendError with a second error to the parent context", func() {
					var secondError *service.Error

					BeforeEach(func() {
						secondError = &service.Error{}
						standard.AppendError("second", secondError)
					})

					It("has errors", func() {
						Expect(standard.Errors()).ToNot(BeEmpty())
					})

					It("has both errors", func() {
						Expect(standard.Errors()).To(ConsistOf(firstError, secondError))
					})

					It("added the error source pointer", func() {
						Expect(standard.Errors()).To(HaveLen(2))
						Expect(standard.Errors()[0]).ToNot(BeNil())
						Expect(standard.Errors()[0].Source).ToNot(BeNil())
						Expect(standard.Errors()[0].Source.Pointer).To(Equal("/child/first"))
						Expect(standard.Errors()[1]).ToNot(BeNil())
						Expect(standard.Errors()[1].Source).ToNot(BeNil())
						Expect(standard.Errors()[1].Source.Pointer).To(Equal("/second"))
					})
				})
			})

			Context("creating a grandchild of the child context", func() {
				var grandchild data.Context

				BeforeEach(func() {
					grandchild = child.NewChildContext("grandchild")
				})

				It("exists", func() {
					Expect(grandchild).ToNot(BeNil())
				})

				It("Logger returns a logger", func() {
					Expect(grandchild.Logger()).ToNot(BeNil())
				})

				Context("ResolveReference", func() {
					It("correctly returns the resolved reference", func() {
						Expect(grandchild.ResolveReference("reference")).To(Equal("/child/grandchild/reference"))
					})
				})

				Context("AppendError with a first error", func() {
					var firstError *service.Error

					BeforeEach(func() {
						firstError = &service.Error{}
						grandchild.AppendError("first", firstError)
					})

					It("has errors", func() {
						Expect(standard.Errors()).ToNot(BeEmpty())
					})

					It("has the error", func() {
						Expect(standard.Errors()).To(ConsistOf(firstError))
					})

					It("added the error source pointer", func() {
						Expect(standard.Errors()).To(HaveLen(1))
						Expect(standard.Errors()[0]).ToNot(BeNil())
						Expect(standard.Errors()[0].Source).ToNot(BeNil())
						Expect(standard.Errors()[0].Source.Pointer).To(Equal("/child/grandchild/first"))
					})
				})
			})
		})
	})
})
