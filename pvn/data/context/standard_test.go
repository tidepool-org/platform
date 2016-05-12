package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Standard", func() {

	Describe("newly created", func() {

		var standard *context.Standard

		BeforeEach(func() {
			standard = context.NewStandard()
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("has a contained Errors that is empty", func() {
			Expect(standard.Errors).ToNot(BeNil())
			Expect(standard.Errors.HasErrors()).To(BeFalse())
			Expect(standard.Errors.GetErrors()).To(BeEmpty())
		})

		It("ignores a nil reference to AppendError", func() {
			standard.AppendError(nil, &service.Error{})
			Expect(standard.Errors.HasErrors()).To(BeFalse())
			Expect(standard.Errors.GetErrors()).To(BeEmpty())
		})

		It("ignores a nil err to AppendError", func() {
			standard.AppendError("ignore", nil)
			Expect(standard.Errors.HasErrors()).To(BeFalse())
			Expect(standard.Errors.GetErrors()).To(BeEmpty())
		})

		Describe("after appending a first error", func() {
			var firstError *service.Error

			BeforeEach(func() {
				firstError = &service.Error{}
				standard.AppendError("first", firstError)
			})

			It("has errors", func() {
				Expect(standard.Errors.HasErrors()).To(BeTrue())
			})

			It("has the error", func() {
				Expect(standard.Errors.GetErrors()).To(ConsistOf(firstError))
			})

			It("added the error source pointer", func() {
				Expect(standard.Errors.GetErrors()).To(HaveLen(1))
				Expect(standard.Errors.GetError(0)).ToNot(BeNil())
				Expect(standard.Errors.GetError(0).Source).ToNot(BeNil())
				Expect(standard.Errors.GetError(0).Source.Pointer).To(Equal("first"))
			})

			Describe("and appending a second error", func() {
				var secondError *service.Error

				BeforeEach(func() {
					secondError = &service.Error{}
					standard.AppendError("second", secondError)
				})

				It("has errors", func() {
					Expect(standard.Errors.HasErrors()).To(BeTrue())
				})

				It("has both errors", func() {
					Expect(standard.Errors.GetErrors()).To(ConsistOf(firstError, secondError))
				})

				It("added the error source pointer", func() {
					Expect(standard.Errors.GetErrors()).To(HaveLen(2))
					Expect(standard.Errors.GetError(0)).ToNot(BeNil())
					Expect(standard.Errors.GetError(0).Source).ToNot(BeNil())
					Expect(standard.Errors.GetError(0).Source.Pointer).To(Equal("first"))
					Expect(standard.Errors.GetError(1)).ToNot(BeNil())
					Expect(standard.Errors.GetError(1).Source).ToNot(BeNil())
					Expect(standard.Errors.GetError(1).Source.Pointer).To(Equal("second"))
				})
			})
		})

		It("returns a nil if NewChildContext is invoked with a nil child context", func() {
			Expect(standard.NewChildContext(nil)).To(BeNil())
		})

		Describe("creating a child context", func() {
			var child data.Context

			BeforeEach(func() {
				child = standard.NewChildContext("child")
			})

			It("exists", func() {
				Expect(child).ToNot(BeNil())
			})

			Describe("after appending a first error", func() {
				var firstError *service.Error

				BeforeEach(func() {
					firstError = &service.Error{}
					child.AppendError("first", firstError)
				})

				It("has errors", func() {
					Expect(standard.Errors.HasErrors()).To(BeTrue())
				})

				It("has the error", func() {
					Expect(standard.Errors.GetErrors()).To(ConsistOf(firstError))
				})

				It("added the error source pointer", func() {
					Expect(standard.Errors.GetErrors()).To(HaveLen(1))
					Expect(standard.Errors.GetError(0)).ToNot(BeNil())
					Expect(standard.Errors.GetError(0).Source).ToNot(BeNil())
					Expect(standard.Errors.GetError(0).Source.Pointer).To(Equal("child/first"))
				})

				Describe("and appending a second error to the parent context", func() {
					var secondError *service.Error

					BeforeEach(func() {
						secondError = &service.Error{}
						standard.AppendError("second", secondError)
					})

					It("has errors", func() {
						Expect(standard.Errors.HasErrors()).To(BeTrue())
					})

					It("has both errors", func() {
						Expect(standard.Errors.GetErrors()).To(ConsistOf(firstError, secondError))
					})

					It("added the error source pointer", func() {
						Expect(standard.Errors.GetErrors()).To(HaveLen(2))
						Expect(standard.Errors.GetError(0)).ToNot(BeNil())
						Expect(standard.Errors.GetError(0).Source).ToNot(BeNil())
						Expect(standard.Errors.GetError(0).Source.Pointer).To(Equal("child/first"))
						Expect(standard.Errors.GetError(1)).ToNot(BeNil())
						Expect(standard.Errors.GetError(1).Source).ToNot(BeNil())
						Expect(standard.Errors.GetError(1).Source.Pointer).To(Equal("second"))
					})
				})
			})

			Describe("creating a grandchild of the child context", func() {
				var grandchild data.Context

				BeforeEach(func() {
					grandchild = child.NewChildContext("grandchild")
				})

				It("exists", func() {
					Expect(grandchild).ToNot(BeNil())
				})

				Describe("after appending a first error", func() {
					var firstError *service.Error

					BeforeEach(func() {
						firstError = &service.Error{}
						grandchild.AppendError("first", firstError)
					})

					It("has errors", func() {
						Expect(standard.Errors.HasErrors()).To(BeTrue())
					})

					It("has the error", func() {
						Expect(standard.Errors.GetErrors()).To(ConsistOf(firstError))
					})

					It("added the error source pointer", func() {
						Expect(standard.Errors.GetErrors()).To(HaveLen(1))
						Expect(standard.Errors.GetError(0)).ToNot(BeNil())
						Expect(standard.Errors.GetError(0).Source).ToNot(BeNil())
						Expect(standard.Errors.GetError(0).Source.Pointer).To(Equal("child/grandchild/first"))
					})
				})
			})
		})
	})
})
