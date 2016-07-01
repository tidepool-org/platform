package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service/middleware"
)

var _ = Describe("Recover", func() {
	Context("NewRecover", func() {
		It("returns successfully", func() {
			recoverMiddleware, err := middleware.NewRecover()
			Expect(err).ToNot(HaveOccurred())
			Expect(recoverMiddleware).ToNot(BeNil())
		})
	})

	Context("with recover middleware, handler, request, and response", func() {
		var recoverMiddleware *middleware.Recover
		var handler rest.HandlerFunc
		var request *rest.Request
		var response *TestResponseWriter

		BeforeEach(func() {
			var err error
			recoverMiddleware, err = middleware.NewRecover()
			Expect(err).ToNot(HaveOccurred())
			Expect(recoverMiddleware).ToNot(BeNil())
			handler = func(response rest.ResponseWriter, request *rest.Request) { panic("test-panic") }
			request = NewTestRequest()
			response = NewTestResponseWriter()
		})

		It("is successful", func() {
			response.WriteJSONOutputs = []error{nil}
			recoverMiddleware.MiddlewareFunc(handler)(response, request)
			Expect(response.WriteHeaderInputs).To(HaveLen(1))
			Expect(response.WriteJSONInputs).To(HaveLen(1))
		})

		It("does nothing if the handler is nil", func() {
			recoverMiddleware.MiddlewareFunc(nil)(response, request)
			Expect(response.WriteHeaderInputs).To(BeEmpty())
			Expect(response.WriteJSONInputs).To(BeEmpty())
		})

		It("does nothing if the response is nil", func() {
			recoverMiddleware.MiddlewareFunc(handler)(nil, request)
			Expect(response.WriteHeaderInputs).To(BeEmpty())
			Expect(response.WriteJSONInputs).To(BeEmpty())
		})

		It("does nothing if the request is nil", func() {
			recoverMiddleware.MiddlewareFunc(nil)(response, nil)
			Expect(response.WriteHeaderInputs).To(BeEmpty())
			Expect(response.WriteJSONInputs).To(BeEmpty())
		})

		It("does nothing if there is no panic", func() {
			handler = func(response rest.ResponseWriter, request *rest.Request) {}
			recoverMiddleware.MiddlewareFunc(handler)(response, request)
			Expect(response.WriteHeaderInputs).To(BeEmpty())
			Expect(response.WriteJSONInputs).To(BeEmpty())
		})
	})
})
