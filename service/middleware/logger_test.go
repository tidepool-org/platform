package middleware_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/ant0ine/go-json-rest/rest"

// 	"github.com/tidepool-org/platform/log"
// 	"github.com/tidepool-org/platform/log/null"
// 	"github.com/tidepool-org/platform/service/middleware"
// )

// var _ = Describe("Logger", func() {
// 	Context("with logger", func() {
// 		var logger log.Logger

// 		BeforeEach(func() {
// 			logger = null.NewLogger()
// 		})

// 		Context("NewLogger", func() {
// 			It("returns successfully", func() {
// 				Expect(middleware.NewLogger(logger)).ToNot(BeNil())
// 			})

// 			It("returns an error if the logger is missing", func() {
// 				loggerMiddleware, err := middleware.NewLogger(nil)
// 				Expect(err).To(MatchError("logger is missing"))
// 				Expect(loggerMiddleware).To(BeNil())
// 			})
// 		})

// 		Context("with logger middleware, handler, request, and response", func() {
// 			var loggerMiddleware *middleware.Logger
// 			var handler rest.HandlerFunc
// 			var request *rest.Request
// 			var response *TestResponseWriter

// 			BeforeEach(func() {
// 				var err error
// 				loggerMiddleware, err = middleware.NewLogger(logger)
// 				Expect(err).ToNot(HaveOccurred())
// 				Expect(loggerMiddleware).ToNot(BeNil())
// 				handler = func(response rest.ResponseWriter, request *rest.Request) {
// 					Expect(request.Env["LOGGER"]).To(Equal(logger))
// 				}
// 				request = NewTestRequest()
// 				response = NewTestResponseWriter()
// 			})

// 			It("is successful", func() {
// 				loggerMiddleware.MiddlewareFunc(handler)(response, request)
// 				Expect(request.Env["LOGGER"]).To(BeNil())
// 			})

// 			It("does nothing if the handler is nil", func() {
// 				loggerMiddleware.MiddlewareFunc(nil)(response, request)
// 				Expect(request.Env["LOGGER"]).To(BeNil())
// 			})

// 			It("does nothing if the response is nil", func() {
// 				loggerMiddleware.MiddlewareFunc(handler)(nil, request)
// 				Expect(request.Env["LOGGER"]).To(BeNil())
// 			})

// 			It("does nothing if the request is nil", func() {
// 				loggerMiddleware.MiddlewareFunc(nil)(response, nil)
// 				Expect(request.Env["LOGGER"]).To(BeNil())
// 			})
// 		})
// 	})
// })
