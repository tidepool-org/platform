package middleware_test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Recover", func() {
	Context("NewRecover", func() {
		It("returns successfully", func() {
			Expect(middleware.NewRecover()).ToNot(BeNil())
		})
	})

	Context("with recover middleware, handler, request, and response", func() {
		var recoverMiddleware *middleware.Recover
		var req *rest.Request
		var res *testRest.ResponseWriter
		var hndlr rest.HandlerFunc

		BeforeEach(func() {
			var err error
			recoverMiddleware, err = middleware.NewRecover()
			Expect(err).ToNot(HaveOccurred())
			Expect(recoverMiddleware).ToNot(BeNil())
			req = testRest.NewRequest()
			res = testRest.NewResponseWriter()
			hndlr = func(res rest.ResponseWriter, req *rest.Request) { panic("test-panic") }
		})

		AfterEach(func() {
			res.AssertOutputsEmpty()
		})

		It("is successful", func() {
			res.HeaderOutput = &http.Header{}
			res.WriteJsonOutputs = []error{nil}
			recoverMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.WriteHeaderInputs).To(Equal([]int{500}))
			Expect(res.WriteJsonInputs).To(HaveLen(1))
		})

		It("does nothing if the handler is nil", func() {
			recoverMiddleware.MiddlewareFunc(nil)(res, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteJsonInputs).To(BeEmpty())
		})

		It("does nothing if the response is nil", func() {
			recoverMiddleware.MiddlewareFunc(hndlr)(nil, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteJsonInputs).To(BeEmpty())
		})

		It("does nothing if the request is nil", func() {
			recoverMiddleware.MiddlewareFunc(hndlr)(res, nil)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteJsonInputs).To(BeEmpty())
		})

		It("does nothing if there is no panic", func() {
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {}
			recoverMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteJsonInputs).To(BeEmpty())
		})
	})
})
