package middleware_test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	testRequest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/service/middleware"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Trace", func() {
	Context("NewTrace", func() {
		It("returns successfully", func() {
			Expect(middleware.NewTrace()).ToNot(BeNil())
		})
	})

	Context("with trace middleware, handler, request, and response", func() {
		var traceMiddleware *middleware.Trace
		var req *rest.Request
		var res *testRest.ResponseWriter
		var traceRequest string
		var traceSession string
		var hndlr rest.HandlerFunc

		BeforeEach(func() {
			var err error
			traceMiddleware, err = middleware.NewTrace()
			Expect(err).ToNot(HaveOccurred())
			Expect(traceMiddleware).ToNot(BeNil())
			req = testRest.NewRequest()
			res = testRest.NewResponseWriter()
			res.HeaderOutput = &http.Header{}
			traceRequest = testRequest.NewTraceRequest()
			req.Request.Header.Set("X-Tidepool-Trace-Request", traceRequest)
			traceSession = testRequest.NewTraceSession()
			req.Request.Header.Set("X-Tidepool-Trace-Session", traceSession)
		})

		AfterEach(func() {
			res.AssertOutputsEmpty()
		})

		Context("without logger", func() {
			BeforeEach(func() {
				hndlr = func(res rest.ResponseWriter, req *rest.Request) {
					Expect(req.Env["LOGGER"]).To(BeNil())
					Expect(req.Env["TRACE-REQUEST"]).To(Equal(traceRequest))
					Expect(req.Env["TRACE-SESSION"]).To(Equal(traceSession))
					ctx := req.Context()
					Expect(log.LoggerFromContext(ctx)).To(BeNil())
					Expect(request.TraceRequestFromContext(ctx)).To(Equal(traceRequest))
					Expect(request.TraceSessionFromContext(ctx)).To(Equal(traceSession))
					Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest}))
					Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession}))
				}
			})

			AfterEach(func() {
				Expect(req.Env["LOGGER"]).To(BeNil())
				Expect(req.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(req.Env["TRACE-SESSION"]).To(BeNil())
				ctx := req.Context()
				Expect(request.TraceRequestFromContext(ctx)).To(BeEmpty())
				Expect(request.TraceSessionFromContext(ctx)).To(BeEmpty())
				Expect(log.LoggerFromContext(ctx)).To(BeNil())
			})

			It("is successful", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(res, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest}))
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession}))
			})

			It("does nothing if the handler is nil", func() {
				traceMiddleware.MiddlewareFunc(nil)(res, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})

			It("does nothing if the response is nil", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(nil, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})

			It("does nothing if the request is nil", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(res, nil)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})
		})

		Context("with logger", func() {
			var lgr log.Logger

			BeforeEach(func() {
				lgr = logNull.NewLogger()
				hndlr = func(res rest.ResponseWriter, req *rest.Request) {
					Expect(req.Env["LOGGER"]).ToNot(BeNil())
					Expect(req.Env["LOGGER"]).ToNot(Equal(lgr))
					Expect(req.Env["TRACE-REQUEST"]).To(Equal(traceRequest))
					Expect(req.Env["TRACE-SESSION"]).To(Equal(traceSession))
					ctx := req.Context()
					Expect(log.LoggerFromContext(ctx)).ToNot(BeNil())
					Expect(log.LoggerFromContext(ctx)).ToNot(Equal(lgr))
					Expect(request.TraceRequestFromContext(ctx)).To(Equal(traceRequest))
					Expect(request.TraceSessionFromContext(ctx)).To(Equal(traceSession))
					Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest}))
					Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession}))
				}
				req.Env["LOGGER"] = lgr
				req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), lgr))
			})

			AfterEach(func() {
				Expect(req.Env["LOGGER"]).To(Equal(lgr))
				Expect(req.Env["TRACE-REQUEST"]).To(BeNil())
				Expect(req.Env["TRACE-SESSION"]).To(BeNil())
				ctx := req.Context()
				Expect(request.TraceRequestFromContext(ctx)).To(BeEmpty())
				Expect(request.TraceSessionFromContext(ctx)).To(BeEmpty())
				Expect(log.LoggerFromContext(ctx)).To(Equal(lgr))
			})

			It("is successful", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(res, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest}))
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession}))
			})

			It("does nothing if the handler is nil", func() {
				traceMiddleware.MiddlewareFunc(nil)(res, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})

			It("does nothing if the response is nil", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(nil, req)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})

			It("does nothing if the request is nil", func() {
				traceMiddleware.MiddlewareFunc(hndlr)(res, nil)
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})
		})

		It("adds trace request if not specified", func() {
			req.Request.Header.Del("X-Tidepool-Trace-Request")
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {
				Expect(req.Env["TRACE-REQUEST"]).ToNot(BeEmpty())
				Expect(req.Env["TRACE-REQUEST"]).ToNot(Equal(traceRequest))
				Expect(request.TraceRequestFromContext(req.Context())).ToNot(BeEmpty())
				Expect(request.TraceRequestFromContext(req.Context())).ToNot(Equal(traceRequest))
				Expect(res.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Request"]).ToNot(Equal([]string{traceRequest}))
			}
			traceMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.Header()["X-Tidepool-Trace-Request"]).ToNot(BeEmpty())
			Expect(res.Header()["X-Tidepool-Trace-Request"]).ToNot(Equal([]string{traceRequest}))
		})

		It("trims trace request to maximum length", func() {
			traceRequest = test.NewVariableString(65, 256, test.CharsetAlphaNumeric)
			req.Request.Header.Set("X-Tidepool-Trace-Request", traceRequest)
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {
				Expect(req.Env["TRACE-REQUEST"]).To(Equal(traceRequest[0:64]))
				Expect(request.TraceRequestFromContext(req.Context())).To(Equal(traceRequest[0:64]))
				Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest[0:64]}))
			}
			traceMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.Header()["X-Tidepool-Trace-Request"]).To(Equal([]string{traceRequest[0:64]}))
		})

		It("does not add trace session if not specified", func() {
			req.Request.Header.Del("X-Tidepool-Trace-Session")
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {
				Expect(req.Env["TRACE-SESSION"]).To(BeNil())
				Expect(request.TraceSessionFromContext(req.Context())).To(BeEmpty())
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
			}
			traceMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.Header()["X-Tidepool-Trace-Session"]).To(BeEmpty())
		})

		It("trims trace session to maximum length", func() {
			traceSession = test.NewVariableString(65, 256, test.CharsetAlphaNumeric)
			req.Request.Header.Set("X-Tidepool-Trace-Session", traceSession)
			hndlr = func(res rest.ResponseWriter, req *rest.Request) {
				Expect(req.Env["TRACE-SESSION"]).To(Equal(traceSession[0:64]))
				Expect(request.TraceSessionFromContext(req.Context())).To(Equal(traceSession[0:64]))
				Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession[0:64]}))
			}
			traceMiddleware.MiddlewareFunc(hndlr)(res, req)
			Expect(res.Header()["X-Tidepool-Trace-Session"]).To(Equal([]string{traceSession[0:64]}))
		})
	})
})
