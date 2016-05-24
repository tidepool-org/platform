package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Standard", func() {
	Describe("New", func() {
		var standard *context.Standard

		BeforeEach(func() {
			standard = context.NewStandard()
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
