package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/ant0ine/go-json-rest/rest"

	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("AccessLog", func() {
	Context("NewAccessLog", func() {
		It("returns successfully", func() {
			Expect(middleware.NewAccessLog()).ToNot(BeNil())
		})
	})

	Context("with access log middleware, handler, request, and response", func() {
		var accessLogMiddleware *middleware.AccessLog
		var hndlr rest.HandlerFunc
		var req *rest.Request
		var res *testRest.ResponseWriter

		BeforeEach(func() {
			var err error
			accessLogMiddleware, err = middleware.NewAccessLog()
			Expect(err).ToNot(HaveOccurred())
			Expect(accessLogMiddleware).ToNot(BeNil())
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {}
			elapsedTime := 1 * time.Second
			startTime := time.Now()
			req = testRest.NewRequest()
			req.Env["BYTE_WRITTEN"] = int64(128000)
			req.Env["ELAPSED_TIME"] = &elapsedTime
			req.Env["ERRORS"] = []*service.Error{
				{
					Code:   "test-error-code",
					Status: 400,
					Title:  "test-error-title",
					Detail: "test-error-detail",
				},
			}
			req.Env["LOGGER"] = logNull.NewLogger()
			req.Env["REMOTE_USER"] = "gollum"
			req.Env["START_TIME"] = &startTime
			req.Env["STATUS_CODE"] = 400
			req.Request.Header.Set("User-Agent", "gomega")
			req.Request.Header.Set("Referer", "golang")
			req.Request.RemoteAddr = "127.0.0.1:1234"
			res = testRest.NewResponseWriter()
		})

		AfterEach(func() {
			res.AssertOutputsEmpty()
		})

		It("is successful", func() {
			accessLogMiddleware.MiddlewareFunc(hndlr)(res, req)
		})

		It("does nothing if the handler is nil", func() {
			accessLogMiddleware.MiddlewareFunc(nil)(res, req)
		})

		It("does nothing if the response is nil", func() {
			accessLogMiddleware.MiddlewareFunc(hndlr)(nil, req)
		})

		It("does nothing if the request is nil", func() {
			accessLogMiddleware.MiddlewareFunc(hndlr)(res, nil)
		})
	})
})
