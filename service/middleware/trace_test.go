package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service/middleware"
)

var _ = Describe("Trace", func() {
	Context("NewTrace", func() {
		It("returns successfully", func() {
			Expect(middleware.NewTrace()).ToNot(BeNil())
		})
	})

	Context("with trace middleware, handler, request, and response", func() {
		var traceMiddleware *middleware.Trace
		var handler rest.HandlerFunc
		var request *rest.Request
		var response *TestResponseWriter

		BeforeEach(func() {
			var err error
			traceMiddleware, err = middleware.NewTrace()
			Expect(err).ToNot(HaveOccurred())
			Expect(traceMiddleware).ToNot(BeNil())
			handler = func(response rest.ResponseWriter, request *rest.Request) {
				Expect(request.Env["LOGGER"]).To(BeNil())
				Expect(request.Env["TRACE-REQUEST"]).ToNot(BeEmpty())
				Expect(request.Env["TRACE-SESSION"]).To(Equal("session-starling"))
				Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
			}
			request = NewTestRequest()
			request.Request.Header.Add("X-Tidepool-Trace-Session", "session-starling")
			response = NewTestResponseWriter()
		})

		It("is successful", func() {
			traceMiddleware.MiddlewareFunc(handler)(response, request)
			Expect(request.Env["LOGGER"]).To(BeNil())
			Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			Expect(request.Env["TRACE-SESSION"]).To(BeNil())
			Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
		})

		It("does nothing if the handler is nil", func() {
			traceMiddleware.MiddlewareFunc(nil)(response, request)
			Expect(request.Env["LOGGER"]).To(BeNil())
			Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			Expect(request.Env["TRACE-SESSION"]).To(BeNil())
			Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
		})

		It("does nothing if the response is nil", func() {
			traceMiddleware.MiddlewareFunc(handler)(nil, request)
			Expect(request.Env["LOGGER"]).To(BeNil())
			Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			Expect(request.Env["TRACE-SESSION"]).To(BeNil())
			Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
		})

		It("does nothing if the request is nil", func() {
			traceMiddleware.MiddlewareFunc(nil)(response, nil)
			Expect(request.Env["LOGGER"]).To(BeNil())
			Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			Expect(request.Env["TRACE-SESSION"]).To(BeNil())
			Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
		})

		It("trims session trace to maximum length", func() {
			request.Request.Header.Set("X-Tidepool-Trace-Session", "session-starling-session-starling-session-starling-session-starling")
			handler = func(response rest.ResponseWriter, request *rest.Request) {
				Expect(request.Env["LOGGER"]).To(BeNil())
				Expect(request.Env["TRACE-REQUEST"]).ToNot(BeEmpty())
				Expect(request.Env["TRACE-SESSION"]).To(Equal("session-starling-session-starling-session-starling-session-starl"))
				Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
			}
			traceMiddleware.MiddlewareFunc(handler)(response, request)
			Expect(request.Env["LOGGER"]).To(BeNil())
			Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			Expect(request.Env["TRACE-SESSION"]).To(BeNil())
			Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
		})

		Context("with logger", func() {
			var logger log.Logger

			BeforeEach(func() {
				logger = null.NewLogger()
				handler = func(response rest.ResponseWriter, request *rest.Request) {
					Expect(request.Env["LOGGER"]).ToNot(BeNil())
					Expect(request.Env["TRACE-REQUEST"]).ToNot(BeEmpty())
					Expect(request.Env["TRACE-SESSION"]).To(Equal("session-starling"))
					Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
				}
				request.Env["LOGGER"] = logger
			})

			It("is successful", func() {
				traceMiddleware.MiddlewareFunc(handler)(response, request)
				Expect(request.Env["LOGGER"]).To(Equal(logger))
				Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(request.Env["TRACE-SESSION"]).To(BeNil())
				Expect(response.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
			})

			It("does nothing if the handler is nil", func() {
				traceMiddleware.MiddlewareFunc(nil)(response, request)
				Expect(request.Env["LOGGER"]).To(Equal(logger))
				Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(request.Env["TRACE-SESSION"]).To(BeNil())
				Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
			})

			It("does nothing if the response is nil", func() {
				traceMiddleware.MiddlewareFunc(handler)(nil, request)
				Expect(request.Env["LOGGER"]).To(Equal(logger))
				Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(request.Env["TRACE-SESSION"]).To(BeNil())
				Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
			})

			It("does nothing if the request is nil", func() {
				traceMiddleware.MiddlewareFunc(nil)(response, nil)
				Expect(request.Env["LOGGER"]).To(Equal(logger))
				Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(request.Env["TRACE-SESSION"]).To(BeNil())
				Expect(response.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
			})

		})
	})
})
