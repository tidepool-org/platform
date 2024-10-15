package errors_test

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

	Context("AsSource", func() {
		It("returns nil if the error is nil", func() {
			Expect(errors.AsSource(nil)).To(BeNil())
		})

		It("returns nil if the error is a standard error", func() {
			err := stdErrors.New("standard library error")
			Expect(errors.AsSource(err)).To(BeNil())
		})

		It("returns nil if the error is an object without a source", func() {
			err := errors.New("object error")
			Expect(errors.AsSource(err)).To(BeNil())
		})

		It("returns a source if the error is an object with a parameter source", func() {
			err := errorsTest.WithParameterSource(errors.New("object error"), "parameter")
			source := errors.AsSource(err)
			Expect(source).ToNot(BeNil())
			Expect(source.Parameter()).To(Equal("parameter"))
			Expect(source.Pointer()).To(Equal(""))
		})

		It("returns a source if the error is an object with a pointer source", func() {
			err := errorsTest.WithPointerSource(errors.New("object error"), "pointer")
			source := errors.AsSource(err)
			Expect(source).ToNot(BeNil())
			Expect(source.Parameter()).To(Equal(""))
			Expect(source.Pointer()).To(Equal("pointer"))
		})

		It("returns nil if the error is an array", func() {
			firstErr := errorsTest.WithParameterSource(errors.New("first error"), "first parameter")
			middleErr := errorsTest.WithParameterSource(errors.New("middle error"), "middle parameter")
			lastErr := errorsTest.WithParameterSource(errors.New("last error"), "last parameter")
			err := errors.Append(firstErr, middleErr, lastErr)
			Expect(errors.AsSource(err)).To(BeNil())
		})
	})

	Context("ToArray", func() {
		It("returns nil if the error is nil", func() {
			Expect(errors.ToArray(nil)).To(BeNil())
		})

		It("returns an error array with just the one error itself if the error is a standard error", func() {
			err := stdErrors.New("standard library error")
			Expect(errors.ToArray(err)).To(Equal([]error{err}))
		})

		It("returns an error array with just the one error itself if the error is an object", func() {
			err := errors.New("object error")
			Expect(errors.ToArray(err)).To(Equal([]error{err}))
		})

		It("returns all of the errors if the error is an array", func() {
			firstErr := errors.New("first error")
			middleErr := errors.New("middle error")
			lastErr := errors.New("last error")
			err := errors.Append(firstErr, middleErr, lastErr)
			Expect(errors.ToArray(err)).To(Equal([]error{firstErr, middleErr, lastErr}))
		})
	})

	Context("First", func() {
		It("returns nil if the error is nil", func() {
			Expect(errors.First(nil)).To(BeNil())
		})

		It("returns the error itself if the error is a standard error", func() {
			err := stdErrors.New("standard library error")
			Expect(errors.First(err)).To(BeIdenticalTo(err))
		})

		It("returns the error itself if the error is an object", func() {
			err := errors.New("object error")
			Expect(errors.First(err)).To(BeIdenticalTo(err))
		})

		It("returns the first error itself if the error is an array", func() {
			firstErr := errors.New("first error")
			middleErr := errors.New("middle error")
			lastErr := errors.New("last error")
			err := errors.Append(firstErr, middleErr, lastErr)
			Expect(errors.First(err)).To(BeIdenticalTo(firstErr))
		})
	})

	Context("Last", func() {
		It("returns nil if the error is nil", func() {
			Expect(errors.Last(nil)).To(BeNil())
		})

		It("returns the error itself if the error is a standard error", func() {
			err := stdErrors.New("standard library error")
			Expect(errors.Last(err)).To(BeIdenticalTo(err))
		})

		It("returns the error itself if the error is an object", func() {
			err := errors.New("object error")
			Expect(errors.Last(err)).To(BeIdenticalTo(err))
		})

		It("returns the first error itself if the error is an array", func() {
			firstErr := errors.New("first error")
			middleErr := errors.New("middle error")
			lastErr := errors.New("last error")
			err := errors.Append(firstErr, middleErr, lastErr)
			Expect(errors.Last(err)).To(BeIdenticalTo(lastErr))
		})
	})

	Context("NewSerializable", func() {
		It("returns nil if the error is nil", func() {
			serializable := errors.NewSerializable(nil)
			Expect(serializable).To(BeNil())
		})

		It("returns a serializable if the error is not nil", func() {
			err := errorsTest.RandomError()
			serializable := errors.NewSerializable(err)
			Expect(serializable).ToNot(BeNil())
			Expect(serializable.Error).To(Equal(err))
		})
	})

	Context("Serializable", func() {

		DescribeTable("parses, validates, and normalizes successfully",
			func(inputJSON string) {
				serializableJSON := `{"error": ` + inputJSON + `}`

				serializableObject := &map[string]any{}
				err := json.Unmarshal([]byte(serializableJSON), serializableObject)
				Expect(err).ToNot(HaveOccurred())

				serializable := &errors.Serializable{}

				parser := structureParser.NewObject(logTest.NewLogger(), serializableObject)
				Expect(parser).ToNot(BeNil())
				serializable.Parse("error", parser)
				Expect(parser.Error()).ToNot(HaveOccurred())

				validator := structureValidator.New(logTest.NewLogger())
				Expect(validator).ToNot(BeNil())
				serializable.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())

				normalizer := structureNormalizer.New(logTest.NewLogger())
				Expect(normalizer).ToNot(BeNil())
				serializable.Normalize(normalizer)
				Expect(normalizer.Error()).ToNot(HaveOccurred())

				outputJSON, err := json.Marshal(serializable)
				Expect(err).ToNot(HaveOccurred())
				Expect(outputJSON).To(MatchJSON(inputJSON))
			},
			Entry("an array", `
					[
						{
							"detail": "standard library error"
						},
						{
							"code": "ñnホôÓ6😀🍕",
							"title": "🦖þ\":;💔🏄VFb",
							"detail": "Üg🍕60🧠{XźsYฝuGטíñnホôÓ6😀🍕,øูLAn🦖þ\":;💔🏄VFb🏄วは",
							"source": {
								"parameter": "b🏄วは",
								"pointer": ",øูLAn"
							},
							"meta": "supermeta",
							"caller": {
								"package": "github.com/tidepool-org/platform/errors/test",
								"function": "RandomError",
								"file": "errors/test/errors.go",
								"line": 28
							}
						},
						{
							"detail": "wrapped error",
							"caller": {
								"package": "github.com/tidepool-org/platform/errors_test",
								"function": "1",
								"file": "errors/errors_test.go",
								"line": 145
							},
							"cause": {
								"detail": "o0cä",
								"caller": {
									"package": "github.com/tidepool-org/platform/errors/test",
									"function": "RandomError",
									"file": "errors/test/errors.go",
									"line": 28
								}
							}
						}
					]
				`),
			Entry("an object", `
					{
						"code": "ñnホôÓ6😀🍕",
						"title": "🦖þ\":;💔🏄VFb",
						"detail": "wrapped error",
						"source": {
							"parameter": "b🏄วは",
							"pointer": ",øูLAn"
						},
						"caller": {
							"package": "github.com/tidepool-org/platform/errors_test",
							"function": "1",
							"file": "errors/errors_test.go",
							"line": 145
						},
						"cause": {
							"detail": "o0cä",
							"caller": {
								"package": "github.com/tidepool-org/platform/errors/test",
								"function": "RandomError",
								"file": "errors/test/errors.go",
								"line": 28
							}
						}
					}
				`),
			Entry("a string", `"test error"`),
		)

		DescribeTable("reports parse errors",
			func(inputJSON string, expectedError string) {
				serializableJSON := `{"error": ` + inputJSON + `}`

				serializableObject := &map[string]any{}
				err := json.Unmarshal([]byte(serializableJSON), serializableObject)
				Expect(err).ToNot(HaveOccurred())

				serializable := &errors.Serializable{}

				parser := structureParser.NewObject(logTest.NewLogger(), serializableObject)
				Expect(parser).ToNot(BeNil())
				serializable.Parse("error", parser)
				Expect(parser.Error()).To(MatchError(expectedError))
			},
			Entry("an array", `
					[
						{
							"detail": "standard library error"
						},
						{
							"code": true,
							"title": true,
							"detail": true,
							"source": {
								"parameter": true,
								"pointer": true,
								"extra": true
							},
							"meta": "supermeta",
							"caller": {
								"package": true,
								"function": true,
								"file": true,
								"line": true,
								"extra": true
							},
							"extra": true
						},
						{
							"detail": "wrapped error",
							"caller": {
								"package": "github.com/tidepool-org/platform/errors_test",
								"function": "1",
								"file": "errors/errors_test.go",
								"line": 145
							},
							"cause": {
								"detail": "o0cä",
								"caller": {
									"package": "github.com/tidepool-org/platform/errors/test",
									"function": "RandomError",
									"file": "errors/test/errors.go",
									"line": 28
								}
							}
						}
					]
				`, "type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not string, but bool, not parsed, type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not int, but bool, not parsed, not parsed"),
			Entry("an object", `
					{
						"code": true,
						"title": true,
						"detail": true,
						"source": {
							"parameter": true,
							"pointer": true,
							"extra": true
						},
						"meta": "supermeta",
						"caller": {
							"package": true,
							"function": true,
							"file": true,
							"line": true,
							"extra": true
						},
						"extra": true
					}
				`, "type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not string, but bool, not parsed, type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not int, but bool, not parsed, not parsed"),
		)

		DescribeTable("reports validate errors",
			func(inputJSON string, expectedError string) {
				serializableJSON := `{"error": ` + inputJSON + `}`

				serializableObject := &map[string]any{}
				err := json.Unmarshal([]byte(serializableJSON), serializableObject)
				Expect(err).ToNot(HaveOccurred())

				serializable := &errors.Serializable{}

				parser := structureParser.NewObject(logTest.NewLogger(), serializableObject)
				Expect(parser).ToNot(BeNil())
				serializable.Parse("error", parser)
				Expect(parser.Error()).ToNot(HaveOccurred())

				validator := structureValidator.New(logTest.NewLogger())
				Expect(validator).ToNot(BeNil())
				serializable.Validate(validator)
				Expect(validator.Error()).To(MatchError(expectedError))
			},
			Entry("an array", `
					[
						{
							"detail": "standard library error"
						},
						{
							"code": "ñnホôÓ6😀🍕",
							"title": "🦖þ\":;💔🏄VFb",
							"detail": "Üg🍕60🧠{XźsYฝuGטíñnホôÓ6😀🍕,øูLAn🦖þ\":;💔🏄VFb🏄วは",
							"source": {
								"parameter": "",
								"pointer": ""
							},
							"meta": "supermeta",
							"caller": {
								"package": "",
								"function": "",
								"file": "",
								"line": -1
							}
						},
						{
							"detail": "wrapped error",
							"caller": {
								"package": "github.com/tidepool-org/platform/errors_test",
								"function": "1",
								"file": "errors/errors_test.go",
								"line": 145
							},
							"cause": {
								"detail": "o0cä",
								"caller": {
									"package": "github.com/tidepool-org/platform/errors/test",
									"function": "RandomError",
									"file": "errors/test/errors.go",
									"line": 28
								}
							}
						}
					]
				`, "value is empty, value is empty, value is empty, value is empty, value -1 is not greater than or equal to 0"),
			Entry("an object", `
					{
						"code": "ñnホôÓ6😀🍕",
						"title": "🦖þ\":;💔🏄VFb",
						"detail": "Üg🍕60🧠{XźsYฝuGטíñnホôÓ6😀🍕,øูLAn🦖þ\":;💔🏄VFb🏄วは",
						"source": {
							"parameter": "",
							"pointer": ""
						},
						"meta": "supermeta",
						"caller": {
							"package": "",
							"function": "",
							"file": "",
							"line": -1
						}
					}
				`, "value is empty, value is empty, value is empty, value is empty, value -1 is not greater than or equal to 0"),
		)

		Context("marshaling and unmarshaling with JSON", func() {
			DescribeTable("work correctly",
				func(err error) {
					wrapped := SerializableWrapper{
						Value: errors.NewSerializable(err),
					}

					serializedJSON, err := json.Marshal(wrapped)
					Expect(err).ToNot(HaveOccurred())
					Expect(serializedJSON).ToNot(BeEmpty())

					deserialized := SerializableWrapper{}
					err = json.Unmarshal(serializedJSON, &deserialized)
					Expect(err).ToNot(HaveOccurred())

					Expect(deserialized).To(Equal(wrapped))
				},
				Entry("with standard library error", stdErrors.New("standard library error")), // Standard library errors used for testing platform errors only
				Entry("with a single error object", errorsTest.RandomError()),
				Entry("with an array of object errors", errors.Append(errorsTest.RandomError(), errorsTest.RandomError())),
				Entry("with a mixed type errors", errors.Append(errorsTest.RandomError(), stdErrors.New("standard library error"))), // Standard library errors used for testing platform errors only
				Entry("with wrapped errors", errors.Wrap(errorsTest.RandomError(), "wrapped error")),
			)
		})

		Context("marshaling and unmarshaling with BSON", func() {
			DescribeTable("work correctly",
				func(err error) {
					wrapped := SerializableWrapper{
						Value: errors.NewSerializable(err),
					}

					serializedBSON, err := bson.Marshal(wrapped)
					Expect(err).ToNot(HaveOccurred())
					Expect(serializedBSON).ToNot(BeEmpty())

					deserialized := SerializableWrapper{}
					err = bson.Unmarshal(serializedBSON, &deserialized)
					Expect(err).ToNot(HaveOccurred())

					Expect(deserialized).To(Equal(wrapped))
				},
				Entry("with standard library error", stdErrors.New("standard library error")), // Standard library errors used for testing platform errors only
				Entry("with a single error object", errorsTest.RandomError()),
				Entry("with an array of object errors", errors.Append(errorsTest.RandomError(), errorsTest.RandomError())),
				Entry("with a mixed type errors", errors.Append(errorsTest.RandomError(), stdErrors.New("standard library error"))), // Standard library errors used for testing platform errors only
				Entry("with wrapped errors", errors.Wrap(errorsTest.RandomError(), "wrapped error")),
			)
		})
	})

	Context("NewContextWithError", func() {
		It("returns a new context", func() {
			ctx := context.Background()
			result := errors.NewContextWithError(ctx, errorsTest.RandomError())
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(Equal(ctx))
		})
	})

	Context("ErrorFromContext", func() {
		It("returns nil if context is nil", func() {
			var ctx context.Context
			Expect(errors.ErrorFromContext(ctx)).To(BeNil())
		})

		It("returns nil if errors is not in context", func() {
			Expect(errors.ErrorFromContext(context.Background())).To(BeNil())
		})

		It("returns errors", func() {
			err := errorsTest.RandomError()
			Expect(errors.ErrorFromContext(errors.NewContextWithError(context.Background(), err))).To(BeIdenticalTo(err))
		})
	})
})
