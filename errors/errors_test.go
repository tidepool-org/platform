package errors_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	errorsTest "github.com/tidepool-org/platform/errors/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e "errors"

	"github.com/tidepool-org/platform/errors"

	//errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
)

type SerializableWrapper struct {
	Value *errors.Serializable `bson:"value"`
}

var _ = Describe("Errors", func() {
	Context("with package and message", func() {
		var msg string

		BeforeEach(func() {
			msg = test.RandomStringFromRange(1, 64)
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
				wrapped := test.RandomStringFromRange(1, 64)
				err := fmt.Errorf("%s", wrapped)
				Expect(errors.Wrap(err, msg)).To(MatchError(msg + "; " + wrapped))
			})

			It("does not fail when err is nil", func() {
				Expect(errors.Wrap(nil, msg)).To(MatchError(msg))
			})
		})

		Context("Wrapf", func() {
			It("returns a formatted error", func() {
				wrapped := test.RandomStringFromRange(1, 64)
				err := fmt.Errorf("%s", wrapped)
				Expect(errors.Wrapf(err, "%d %s", 222, msg)).To(MatchError("222 " + msg + "; " + wrapped))
			})

			It("does not fail when err is nil", func() {
				Expect(errors.Wrapf(nil, "%d %s", 333, msg)).To(MatchError("333 " + msg))
			})
		})
	})

	Context("Marshal / Unmarshal", func() {
		DescribeTable("works correctly",
			func(err error) {
				wrapped := SerializableWrapper{
					Value: errors.NewSerializable(err),
				}

				serialized, err := bson.Marshal(wrapped)
				Expect(err).ToNot(HaveOccurred())

				deserialized := SerializableWrapper{}
				err = bson.Unmarshal(serialized, &deserialized)
				Expect(err).ToNot(HaveOccurred())
				Expect(deserialized).To(Equal(deserialized))
			},
			Entry("with std lib error", e.New("std lib err")),
			Entry("with a single error object", errorsTest.RandomError()),
			Entry("with an array of object errors", errors.Append(errorsTest.RandomError(), errorsTest.RandomError())),
			Entry("with a mixed type errors", errors.Append(errorsTest.RandomError(), e.New("std lib err"))),
			Entry("with nested errors", errors.Append(errorsTest.RandomError(), errorsTest.RandomError())),
		)
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
	// 		code := test.RandomStringFromRange(1, 16)
	// 		title := test.RandomStringFromRange(1, 64)
	// 		detail := test.RandomStringFromRange(1, 64)
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
	// 		code = test.RandomStringFromRange(1, 16)
	// 		title = test.RandomStringFromRange(1, 64)
	// 		detail = test.RandomStringFromRange(1, 64)
	// 		source = errors.NewSource()
	// 		Expect(source).ToNot(BeNil())
	// 		source.Parameter = errorsTest.NewSourceParameter()
	// 		source.Pointer = errorsTest.NewSourcePointer()
	// 		meta = test.RandomStringFromRange(1, 64)
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
	// 			withSource.Parameter = errorsTest.NewSourceParameter()
	// 			withSource.Pointer = errorsTest.NewSourcePointer()
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
	// 			withMeta := test.RandomStringFromRange(1, 64)
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
	// 				code := test.RandomStringFromRange(1, 16)
	// 				title := test.RandomStringFromRange(1, 64)
	// 				detail := test.RandomStringFromRange(1, 64)
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
	// 				code := test.RandomStringFromRange(1, 16)
	// 				title := test.RandomStringFromRange(1, 64)
	// 				detail := test.RandomStringFromRange(1, 64)
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
