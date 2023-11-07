package middleware_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/request"
	serviceMiddleware "github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Error", func() {
	Context("NewError", func() {
		It("returns successfully", func() {
			Expect(serviceMiddleware.NewError()).ToNot(BeNil())
		})
	})

	Context("with middleware, request, response, and handler", func() {
		var middleware *serviceMiddleware.Error
		var req *rest.Request
		var res *testRest.ResponseWriter
		var hndlr rest.HandlerFunc

		BeforeEach(func() {
			var err error
			middleware, err = serviceMiddleware.NewError()
			Expect(err).ToNot(HaveOccurred())
			Expect(middleware).ToNot(BeNil())
			req = testRest.NewRequest()
			res = testRest.NewResponseWriter()
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {
				Expect(request.ContextErrorFromContext(req.Context())).ToNot(BeNil())
			}
		})

		AfterEach(func() {
			Expect(request.ContextErrorFromContext(req.Context())).To(BeNil())
			res.AssertOutputsEmpty()
		})

		It("is successful", func() {
			middleware.MiddlewareFunc(hndlr)(res, req)
		})

		It("does nothing if the handler is nil", func() {
			middleware.MiddlewareFunc(nil)(res, req)
		})

		It("does nothing if the response is nil", func() {
			middleware.MiddlewareFunc(hndlr)(nil, req)
		})

		It("does nothing if the request is nil", func() {
			middleware.MiddlewareFunc(hndlr)(res, nil)
		})
	})
})
