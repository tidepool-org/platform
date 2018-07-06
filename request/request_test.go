package request_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"github.com/ant0ine/go-json-rest/rest"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/request"
	testHTTP "github.com/tidepool-org/platform/test/http"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Request", func() {
	Context("DecodeRequestPathParameter", func() {
		var req *rest.Request
		var key string
		var value string
		var validator func(value string) bool

		BeforeEach(func() {
			req = testRest.NewRequest()
			key = testHTTP.NewParameterKey()
			value = testHTTP.NewParameterValue()
			validator = func(value string) bool { return true }
			req.PathParams[key] = value
		})

		It("returns error if the request is missing", func() {
			result, err := request.DecodeRequestPathParameter(nil, key, validator)
			Expect(err).To(MatchError("request is missing"))
			Expect(result).To(BeEmpty())
		})

		It("returns error if parameter is not found", func() {
			delete(req.PathParams, key)
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			errorsTest.ExpectEqual(err, request.ErrorParameterMissing(key))
			Expect(result).To(BeEmpty())
		})

		It("returns error if parameter is empty", func() {
			req.PathParams[key] = ""
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			errorsTest.ExpectEqual(err, request.ErrorParameterMissing(key))
			Expect(result).To(BeEmpty())
		})

		It("returns error if validator returns false", func() {
			result, err := request.DecodeRequestPathParameter(req, key, func(value string) bool { return false })
			errorsTest.ExpectEqual(err, request.ErrorParameterInvalid(key))
			Expect(result).To(BeEmpty())
		})

		It("returns successfully if validator returns true", func() {
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(value))
		})

		It("returns successfully if validator is not specified", func() {
			result, err := request.DecodeRequestPathParameter(req, key, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(value))
		})
	})

	Context("ContextError", func() {
		Context("NewContextError", func() {
			It("return successfully", func() {
				Expect(request.NewContextError()).ToNot(BeNil())
			})
		})

		Context("with context error", func() {
			var contextError *request.ContextError

			BeforeEach(func() {
				contextError = request.NewContextError()
				Expect(contextError).ToNot(BeNil())
			})

			It("does not have an error by default", func() {
				Expect(contextError.Get()).To(BeNil())
			})

			Context("Get", func() {
				It("returns the error", func() {
					err := errorsTest.NewError()
					contextError.Set(err)
					Expect(contextError.Get()).To(Equal(err))
				})
			})

			Context("Set", func() {
				It("set the error", func() {
					err := errorsTest.NewError()
					contextError.Set(err)
					Expect(contextError.Get()).To(Equal(err))
				})
			})
		})

		Context("with context", func() {
			var ctx context.Context

			BeforeEach(func() {
				ctx = context.Background()
			})

			Context("NewContextWithContextError", func() {
				It("returns a context with a context error", func() {
					Expect(request.ContextErrorFromContext(request.NewContextWithContextError(ctx))).ToNot(BeNil())
				})
			})

			Context("ContextErrorFromContext", func() {
				It("returns nil if it does not exist in the context", func() {
					Expect(request.ContextErrorFromContext(ctx)).To(BeNil())
				})

				It("returns the context error if it exists in the context", func() {
					Expect(request.ContextErrorFromContext(request.NewContextWithContextError(ctx))).ToNot(BeNil())
				})
			})

			Context("Get", func() {
				It("returns nil", func() {
					request.SetErrorToContext(ctx, errorsTest.NewError())
					Expect(request.GetErrorFromContext(ctx)).To(BeNil())
				})
			})

			Context("Set", func() {
				It("does not set the error", func() {
					request.SetErrorToContext(ctx, errorsTest.NewError())
					Expect(request.GetErrorFromContext(ctx)).To(BeNil())
				})
			})

			Context("with context error", func() {
				BeforeEach(func() {
					ctx = request.NewContextWithContextError(ctx)
				})

				It("does not have an error by default", func() {
					Expect(request.GetErrorFromContext(ctx)).To(BeNil())
				})

				Context("Get", func() {
					It("returns the error", func() {
						err := errorsTest.NewError()
						request.SetErrorToContext(ctx, err)
						Expect(request.GetErrorFromContext(ctx)).To(Equal(err))
					})
				})

				Context("Set", func() {
					It("set the error", func() {
						err := errorsTest.NewError()
						request.SetErrorToContext(ctx, err)
						Expect(request.GetErrorFromContext(ctx)).To(Equal(err))
					})
				})
			})
		})
	})
})
