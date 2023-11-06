package middleware_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Logger", func() {
	Context("with logger", func() {
		var lgr log.Logger

		BeforeEach(func() {
			lgr = logNull.NewLogger()
		})

		Context("NewLogger", func() {
			It("returns an error if the logger is missing", func() {
				loggerMiddleware, err := middleware.NewLogger(nil)
				Expect(err).To(MatchError("logger is missing"))
				Expect(loggerMiddleware).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(middleware.NewLogger(lgr)).ToNot(BeNil())
			})
		})

		Context("with logger middleware, handler, request, and response", func() {
			var loggerMiddleware *middleware.Logger
			var req *rest.Request
			var res *testRest.ResponseWriter
			var hndlr rest.HandlerFunc

			BeforeEach(func() {
				var err error
				loggerMiddleware, err = middleware.NewLogger(lgr)
				Expect(err).ToNot(HaveOccurred())
				Expect(loggerMiddleware).ToNot(BeNil())
				req = testRest.NewRequest()
				res = testRest.NewResponseWriter()
				hndlr = func(res rest.ResponseWriter, req *rest.Request) {
					Expect(req.Env["LOGGER"]).To(Equal(lgr))
					Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
				}
			})

			AfterEach(func() {
				Expect(req.Env["LOGGER"]).To(BeNil())
				Expect(log.LoggerFromContext(req.Context())).To(BeNil())
				res.AssertOutputsEmpty()
			})

			It("is successful", func() {
				loggerMiddleware.MiddlewareFunc(hndlr)(res, req)
			})

			It("does nothing if the handler is nil", func() {
				loggerMiddleware.MiddlewareFunc(nil)(res, req)
			})

			It("does nothing if the response is nil", func() {
				loggerMiddleware.MiddlewareFunc(hndlr)(nil, req)
			})

			It("does nothing if the request is nil", func() {
				loggerMiddleware.MiddlewareFunc(hndlr)(res, nil)
			})
		})
	})
})
