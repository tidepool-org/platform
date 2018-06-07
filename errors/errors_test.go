package errors_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Errors", func() {
	Context("with package and message", func() {
		var msg string

		BeforeEach(func() {
			msg = test.NewText(1, 64)
		})

		Context("New", func() {
			It("returns a formatted error", func() {
				Expect(errors.New(msg)).To(MatchError(msg))
			})
		})

		Context("Newf", func() {
			It("returns a formatted error", func() {
				Expect(errors.Newf("%d %s", 111, msg)).To(MatchError("111 " + msg))
			})
		})

		Context("Wrap", func() {
			It("returns a formatted error", func() {
				wrapped := test.NewText(1, 64)
				err := fmt.Errorf("%s", wrapped)
				Expect(errors.Wrap(err, msg)).To(MatchError(msg + "; " + wrapped))
			})

			It("does not fail when err is nil", func() {
				Expect(errors.Wrap(nil, msg)).To(MatchError(msg))
			})
		})

		Context("Wrapf", func() {
			It("returns a formatted error", func() {
				wrapped := test.NewText(1, 64)
				err := fmt.Errorf("%s", wrapped)
				Expect(errors.Wrapf(err, "%d %s", 222, msg)).To(MatchError("222 " + msg + "; " + wrapped))
			})

			It("does not fail when err is nil", func() {
				Expect(errors.Wrapf(nil, "%d %s", 333, msg)).To(MatchError("333 " + msg))
			})
		})
	})

	// Context("NewSource", func() {
	// 	It("return successfully", func() {
	// 		source := errors.NewSource()
	// 		Expect(source).ToNot(BeNil())
	// 		Expect(source.Parameter).To(BeEmpty())
	// 		Expect(source.Pointer).To(BeEmpty())
	// 	})
	// })

	// Context("NewError", func() {
	// 	It("return successfully", func() {
	// 		code := test.NewText(1, 16)
	// 		title := test.NewText(1, 64)
	// 		detail := test.NewText(1, 64)
	// 		err := errors.Prepared(code, title, detail)
	// 		Expect(err).ToNot(BeNil())
	// 		Expect(err.Code).To(Equal(code))
	// 		Expect(err.Title).To(Equal(title))
	// 		Expect(err.Detail).To(Equal(detail))
	// 	})
	// })

	// Context("with new error", func() {
	// 	var code string
	// 	var title string
	// 	var detail string
	// 	var source *errors.Source
	// 	var meta interface{}
	// 	var err *errors.Error

	// 	BeforeEach(func() {
	// 		code = test.NewText(1, 16)
	// 		title = test.NewText(1, 64)
	// 		detail = test.NewText(1, 64)
	// 		source = errors.NewSource()
	// 		Expect(source).ToNot(BeNil())
	// 		source.Parameter = testErrors.NewSourceParameter()
	// 		source.Pointer = testErrors.NewSourcePointer()
	// 		meta = test.NewText(1, 64)
	// 		err = errors.Prepared(code, title, detail)
	// 		Expect(err).ToNot(BeNil())
	// 		err.Source = source
	// 		err.Meta = meta
	// 	})

	// 	Context("Error", func() {
	// 		It("returns the expected string", func() {
	// 			Expect(err.Error()).To(Equal(detail))
	// 		})
	// 	})

	// 	Context("WithSource", func() {
	// 		It("returns a copy of the error with new source", func() {
	// 			withSource := errors.NewSource()
	// 			withSource.Parameter = testErrors.NewSourceParameter()
	// 			withSource.Pointer = testErrors.NewSourcePointer()
	// 			result := err.WithSource(withSource)
	// 			Expect(result).ToNot(BeNil())
	// 			Expect(result).ToNot(BeIdenticalTo(err))
	// 			Expect(result.Code).To(Equal(code))
	// 			Expect(result.Title).To(Equal(title))
	// 			Expect(result.Detail).To(Equal(detail))
	// 			Expect(result.Source).To(BeIdenticalTo(withSource))
	// 			Expect(result.Meta).To(BeIdenticalTo(meta))
	// 		})
	// 	})

	// 	Context("WithMeta", func() {
	// 		It("returns a copy of the error with new meta", func() {
	// 			withMeta := test.NewText(1, 64)
	// 			result := err.WithMeta(withMeta)
	// 			Expect(result).ToNot(BeNil())
	// 			Expect(result).ToNot(BeIdenticalTo(err))
	// 			Expect(result.Code).To(Equal(code))
	// 			Expect(result.Title).To(Equal(title))
	// 			Expect(result.Detail).To(Equal(detail))
	// 			Expect(result.Source).To(Equal(source))
	// 			Expect(result.Source).ToNot(BeIdenticalTo(source))
	// 			Expect(result.Meta).To(BeIdenticalTo(withMeta))
	// 		})
	// 	})
	// })

	// Context("NewErrors", func() {
	// 	It("return successfully", func() {
	// 		errs := errors.NewErrors()
	// 		Expect(errs).ToNot(BeNil())
	// 		Expect(*errs).To(BeEmpty())
	// 	})
	// })

	// Context("with new errors", func() {
	// 	var errs *errors.Errors

	// 	BeforeEach(func() {
	// 		errs = errors.NewErrors()
	// 		Expect(errs).ToNot(BeNil())
	// 	})

	// 	Context("Error", func() {
	// 		It("returns the expected string", func() {
	// 			expected := []string{}
	// 			for index := 0; index < 3; index++ {
	// 				code := test.NewText(1, 16)
	// 				title := test.NewText(1, 64)
	// 				detail := test.NewText(1, 64)
	// 				err := errors.Prepared(code, title, detail)
	// 				Expect(err).ToNot(BeNil())
	// 				errs.Append(err)
	// 				expected = append(expected, err.Error())
	// 			}
	// 			Expect(errs.Error()).To(Equal(strings.Join(expected, ", ")))
	// 		})
	// 	})

	// 	Context("Append", func() {
	// 		It("successfully appends errors", func() {
	// 			expected := []interface{}{}
	// 			for index := 0; index < 3; index++ {
	// 				code := test.NewText(1, 16)
	// 				title := test.NewText(1, 64)
	// 				detail := test.NewText(1, 64)
	// 				err := errors.Prepared(code, title, detail)
	// 				Expect(err).ToNot(BeNil())
	// 				errs.Append(err)
	// 				expected = append(expected, err)
	// 			}
	// 			Expect(*errs).To(ConsistOf(expected...))
	// 		})

	// 		It("does not append if the error is nil", func() {
	// 			errs.Append(nil)
	// 			Expect(*errs).To(BeEmpty())
	// 		})
	// 	})
	// })

	// Context("NewContextWithError", func() {
	// 	It("returns a new context", func() {
	// 		ctx := context.Background()
	// 		result := errors.NewContextWithError(ctx, errors.NewErrors())
	// 		Expect(result).ToNot(BeNil())
	// 		Expect(result).ToNot(Equal(ctx))
	// 	})
	// })

	// Context("ErrorFromContext", func() {
	// 	It("returns nil if context is nil", func() {
	// 		Expect(errors.ErrorFromContext(nil)).To(BeNil())
	// 	})

	// 	It("returns nil if errors is not in context", func() {
	// 		Expect(errors.ErrorFromContext(context.Background())).To(BeNil())
	// 	})

	// 	It("returns errors", func() {
	// 		errs := errors.NewErrors()
	// 		Expect(errors.ErrorFromContext(errors.NewContextWithError(context.Background(), errs))).To(BeIdenticalTo(errs))
	// 	})
	// })
})
