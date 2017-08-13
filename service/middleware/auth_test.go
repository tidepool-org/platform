package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	"github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Auth", func() {
	var authClient *testAuth.Client

	BeforeEach(func() {
		authClient = testAuth.NewClient()
		Expect(authClient).ToNot(BeNil())
	})

	AfterEach(func() {
		Expect(authClient.UnusedOutputsCount()).To(Equal(0))
	})

	Context("NewAuth", func() {
		It("returns an error if auth client is missing", func() {
			authMiddleware, err := middleware.NewAuth(nil)
			Expect(err).To(MatchError("middleware: auth client is missing"))
			Expect(authMiddleware).To(BeNil())
		})

		It("returns successfully", func() {
			authMiddleware, err := middleware.NewAuth(authClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(authMiddleware).ToNot(BeNil())
		})
	})

	Context("with auth middleware", func() {
		var authMiddleware *middleware.Auth
		var handlerFunc rest.HandlerFunc

		BeforeEach(func() {
			var err error
			authMiddleware, err = middleware.NewAuth(authClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(authMiddleware).ToNot(BeNil())
			handlerFunc = func(response rest.ResponseWriter, request *rest.Request) {
				Expect(response).ToNot(BeNil())
				Expect(request).ToNot(BeNil())
				Expect(service.GetRequestAuthDetails(request)).To(BeNil())
				response.WriteHeader(0)
			}
		})

		Context("MiddlewareFunc", func() {
			It("returns successfully", func() {
				Expect(authMiddleware.MiddlewareFunc(handlerFunc)).ToNot(BeNil())
			})

			It("does nothing if handlerFunc is nil", func() {
				middlewareFunc := authMiddleware.MiddlewareFunc(nil)
				Expect(middlewareFunc).ToNot(BeNil())
				middlewareFunc(testRest.NewResponseWriter(), testRest.NewRequest())
			})
		})

		Context("with response, request, and middleware func", func() {
			var token string
			var response *testRest.ResponseWriter
			var request *rest.Request
			var middlewareFunc rest.HandlerFunc

			BeforeEach(func() {
				token = id.New()
				response = testRest.NewResponseWriter()
				request = testRest.NewRequest()
				request.Header.Set(auth.TidepoolAuthTokenHeaderName, token)
				middlewareFunc = authMiddleware.MiddlewareFunc(func(response rest.ResponseWriter, request *rest.Request) { handlerFunc(response, request) })
				Expect(middlewareFunc).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(response.UnusedOutputsCount()).To(Equal(0))
			})

			It("does nothing if response is nil", func() {
				middlewareFunc(nil, request)
				Expect(response.WriteHeaderInputs).To(BeEmpty())
			})

			It("does nothing if request is nil", func() {
				middlewareFunc(response, nil)
				Expect(response.WriteHeaderInputs).To(BeEmpty())
			})

			It("does nothing if the token is missing", func() {
				request.Header.Del(auth.TidepoolAuthTokenHeaderName)
				middlewareFunc(response, request)
				Expect(response.WriteHeaderInputs).To(Equal([]int{0}))
			})

			Context("with auth details", func() {
				var authDetails *testAuth.Details

				BeforeEach(func() {
					authDetails = testAuth.NewDetails()
					Expect(authDetails).ToNot(BeNil())
					authClient.ValidateTokenOutputs = []testAuth.ValidateTokenOutput{{Details: authDetails, Error: nil}}
				})

				It("does nothing if the token is unauthorized", func() {
					authClient.ValidateTokenOutputs = []testAuth.ValidateTokenOutput{{Details: nil, Error: client.NewUnauthorizedError()}}
					middlewareFunc(response, request)
					Expect(response.WriteHeaderInputs).To(Equal([]int{0}))
				})

				It("responds with internal service failure if the token has other error", func() {
					response.WriteJsonOutputs = []error{nil}
					authClient.ValidateTokenOutputs = []testAuth.ValidateTokenOutput{{Details: nil, Error: errors.New("test")}}
					middlewareFunc(response, request)
					Expect(response.WriteHeaderInputs).To(Equal([]int{500}))
					Expect(response.WriteJsonInputs).To(HaveLen(1))
					Expect(response.WriteJsonInputs[0].(*context.JSONResponse).Errors).To(Equal([]*service.Error{service.ErrorInternalServerFailure()}))
				})

				It("returns succesfully and sets auth details", func() {
					handlerFunc = func(response rest.ResponseWriter, request *rest.Request) {
						Expect(response).ToNot(BeNil())
						Expect(request).ToNot(BeNil())
						Expect(service.GetRequestAuthDetails(request)).To(Equal(authDetails))
						response.WriteHeader(234)
					}
					middlewareFunc(response, request)
					Expect(response.WriteHeaderInputs).To(Equal([]int{234}))
					Expect(service.GetRequestAuthDetails(request)).To(BeNil())
				})

				It("returns succesfully and sets and restores auth details", func() {
					previousAuthDetails := testAuth.NewDetails()
					service.SetRequestAuthDetails(request, previousAuthDetails)
					handlerFunc = func(response rest.ResponseWriter, request *rest.Request) {
						Expect(response).ToNot(BeNil())
						Expect(request).ToNot(BeNil())
						Expect(service.GetRequestAuthDetails(request)).To(Equal(authDetails))
						response.WriteHeader(345)
					}
					middlewareFunc(response, request)
					Expect(response.WriteHeaderInputs).To(Equal([]int{345}))
					Expect(service.GetRequestAuthDetails(request)).To(Equal(previousAuthDetails))
				})
			})
		})
	})
})
