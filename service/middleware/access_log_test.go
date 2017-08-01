package middleware_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
)

var _ = Describe("AccessLog", func() {
	Context("NewAccessLog", func() {
		It("returns successfully", func() {
			Expect(middleware.NewAccessLog()).ToNot(BeNil())
		})
	})

	Context("with access log middleware, handler, request, and response", func() {
		var accessLogMiddleware *middleware.AccessLog
		var handler rest.HandlerFunc
		var request *rest.Request
		var response *TestResponseWriter

		BeforeEach(func() {
			var err error
			accessLogMiddleware, err = middleware.NewAccessLog()
			Expect(err).ToNot(HaveOccurred())
			Expect(accessLogMiddleware).ToNot(BeNil())
			handler = func(response rest.ResponseWriter, request *rest.Request) {}
			elapsedTime := 1 * time.Second
			startTime := time.Now()
			request = NewTestRequest()
			request.Env["BYTE_WRITTEN"] = int64(128000)
			request.Env["ELAPSED_TIME"] = &elapsedTime
			request.Env["ERRORS"] = []*service.Error{
				{
					Code:   "test-error-code",
					Status: 400,
					Title:  "test-error-title",
					Detail: "test-error-detail",
				},
			}
			request.Env["LOGGER"] = log.NewNull()
			request.Env["REMOTE_USER"] = "gollum"
			request.Env["START_TIME"] = &startTime
			request.Env["STATUS_CODE"] = 400
			request.Request.Header.Set("User-Agent", "gomega")
			request.Request.Header.Set("Referer", "golang")
			request.Request.RemoteAddr = "127.0.0.1:1234"
			response = NewTestResponseWriter()
		})

		It("is successful", func() {
			accessLogMiddleware.MiddlewareFunc(handler)(response, request)
		})

		It("does nothing if the handler is nil", func() {
			accessLogMiddleware.MiddlewareFunc(nil)(response, request)
		})

		It("does nothing if the response is nil", func() {
			accessLogMiddleware.MiddlewareFunc(handler)(nil, request)
		})

		It("does nothing if the request is nil", func() {
			accessLogMiddleware.MiddlewareFunc(nil)(response, nil)
		})
	})
})
