package service_test

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	serviceTest "github.com/tidepool-org/platform/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Request", func() {
	Context("with request", func() {
		var errs []*service.Error
		var logger log.Logger
		var details request.Details
		var req *rest.Request

		BeforeEach(func() {
			errs = []*service.Error{
				{
					Code:   "test-error-code",
					Status: 400,
					Title:  "test-error-title",
					Detail: "test-error-detail",
				},
			}
			logger = logNull.NewLogger()
			details = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken())
			req = testRest.NewRequest()
			req.Env["ERRORS"] = errs
			req.Env["LOGGER"] = logger
			req.Env["AUTH-DETAILS"] = details
			req.Env["TRACE-REQUEST"] = "request-raven"
			req.Env["TRACE-SESSION"] = "session-starling"
		})

		Context("GetRequestErrors", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestErrors(req)).To(Equal(errs))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestErrors(nil)).To(BeNil())
			})

			It("returns nil if the errors is not of the correct type", func() {
				req.Env["ERRORS"] = 0
				Expect(service.GetRequestErrors(req)).To(BeNil())
			})

			It("returns nil if the errors is missing", func() {
				delete(req.Env, "ERRORS")
				Expect(service.GetRequestErrors(req)).To(BeNil())
			})
		})

		Context("SetRequestErrors", func() {
			var newErrs []*service.Error

			BeforeEach(func() {
				newErrs = []*service.Error{
					{
						Code:   "test-error-code-new",
						Status: 400,
						Title:  "test-error-title-new",
						Detail: "test-error-detail-new",
					},
				}
			})

			It("successfully sets the errors", func() {
				service.SetRequestErrors(req, newErrs)
				Expect(req.Env["ERRORS"]).To(Equal(newErrs))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestErrors(nil, newErrs)
				Expect(req.Env["ERRORS"]).To(Equal(errs))
			})

			It("deletes the errors if the errors is missing", func() {
				service.SetRequestErrors(req, nil)
				Expect(req.Env["ERRORS"]).To(BeNil())
			})
		})

		Context("GetRequestLogger", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestLogger(req)).To(Equal(logger))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestLogger(nil)).To(BeNil())
			})

			It("returns nil if the logger is not of the correct type", func() {
				req.Env["LOGGER"] = 0
				Expect(service.GetRequestLogger(req)).To(BeNil())
			})

			It("returns nil if the logger is missing", func() {
				delete(req.Env, "LOGGER")
				Expect(service.GetRequestLogger(req)).To(BeNil())
			})
		})

		Context("SetRequestLogger", func() {
			var newLogger log.Logger

			BeforeEach(func() {
				newLogger = logNull.NewLogger()
			})

			It("successfully sets the logger", func() {
				service.SetRequestLogger(req, newLogger)
				Expect(req.Env["LOGGER"]).To(Equal(newLogger))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestLogger(nil, newLogger)
				Expect(req.Env["LOGGER"]).To(Equal(logger))
			})

			It("deletes the logger if the logger is missing", func() {
				service.SetRequestLogger(req, nil)
				Expect(req.Env["LOGGER"]).To(BeNil())
			})
		})

		Context("GetRequestAuthDetails", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestAuthDetails(nil)).To(BeNil())
			})

			It("returns nil if the auth details is not of the correct type", func() {
				req.Env["AUTH-DETAILS"] = 0
				Expect(service.GetRequestAuthDetails(req)).To(BeNil())
			})

			It("returns nil if the auth details is missing", func() {
				delete(req.Env, "AUTH-DETAILS")
				Expect(service.GetRequestAuthDetails(req)).To(BeNil())
			})
		})

		Context("SetRequestAuthDetails", func() {
			var newDetails request.Details

			BeforeEach(func() {
				newDetails = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken())
			})

			It("successfully sets the auth details", func() {
				service.SetRequestAuthDetails(req, newDetails)
				Expect(req.Env["AUTH-DETAILS"]).To(Equal(newDetails))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestAuthDetails(nil, newDetails)
				Expect(req.Env["AUTH-DETAILS"]).To(Equal(details))
			})

			It("deletes the auth details if the auth details is missing", func() {
				service.SetRequestAuthDetails(req, nil)
				Expect(req.Env["AUTH-DETAILS"]).To(BeNil())
			})
		})

		Context("GetRequestTraceRequest", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestTraceRequest(req)).To(Equal("request-raven"))
			})

			It("returns empty string if the request is missing", func() {
				Expect(service.GetRequestTraceRequest(nil)).To(BeEmpty())
			})

			It("returns empty string if the request trace is not of the correct type", func() {
				req.Env["TRACE-REQUEST"] = 0
				Expect(service.GetRequestTraceRequest(req)).To(BeEmpty())
			})

			It("returns empty string if the request trace is missing", func() {
				delete(req.Env, "TRACE-REQUEST")
				Expect(service.GetRequestTraceRequest(req)).To(BeEmpty())
			})
		})

		Context("SetRequestTraceRequest", func() {
			It("successfully sets the request trace", func() {
				service.SetRequestTraceRequest(req, "request-raven-new")
				Expect(req.Env["TRACE-REQUEST"]).To(Equal("request-raven-new"))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestTraceRequest(nil, "request-raven-new")
				Expect(req.Env["TRACE-REQUEST"]).To(Equal("request-raven"))
			})

			It("deletes the request trace if the request trace is missing", func() {
				service.SetRequestTraceRequest(req, "")
				Expect(req.Env["TRACE-REQUEST"]).To(BeNil())
			})
		})

		Context("GetRequestTraceSession", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestTraceSession(req)).To(Equal("session-starling"))
			})

			It("returns empty string if the request is missing", func() {
				Expect(service.GetRequestTraceSession(nil)).To(BeEmpty())
			})

			It("returns empty string if the session trace is not of the correct type", func() {
				req.Env["TRACE-SESSION"] = 0
				Expect(service.GetRequestTraceSession(req)).To(BeEmpty())
			})

			It("returns empty string if the session trace is missing", func() {
				delete(req.Env, "TRACE-SESSION")
				Expect(service.GetRequestTraceSession(req)).To(BeEmpty())
			})
		})

		Context("SetRequestTraceSession", func() {
			It("successfully sets the session trace", func() {
				service.SetRequestTraceSession(req, "session-starling-new")
				Expect(req.Env["TRACE-SESSION"]).To(Equal("session-starling-new"))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestTraceSession(nil, "session-starling-new")
				Expect(req.Env["TRACE-SESSION"]).To(Equal("session-starling"))
			})

			It("deletes the session trace if the session trace is missing", func() {
				service.SetRequestTraceSession(req, "")
				Expect(req.Env["TRACE-SESSION"]).To(BeNil())
			})
		})

		Context("CopyRequestTrace", func() {
			var destinationRequest *http.Request

			BeforeEach(func() {
				var err error
				destinationRequest, err = http.NewRequest("GET", "http://127.0.0.1/", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(destinationRequest).ToNot(BeNil())
			})

			It("is successful", func() {
				Expect(service.CopyRequestTrace(req, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Request"]).To(ConsistOf("request-raven"))
				Expect(destinationRequest.Header["X-Tidepool-Trace-Session"]).To(ConsistOf("session-starling"))
			})

			It("returns an error if the source request is missing", func() {
				Expect(service.CopyRequestTrace(nil, destinationRequest)).To(MatchError("source request is missing"))
			})

			It("returns an error if the destination request is missing", func() {
				Expect(service.CopyRequestTrace(req, nil)).To(MatchError("destination request is missing"))
			})

			It("is successful even if request trace not set", func() {
				delete(req.Env, "TRACE-REQUEST")
				Expect(service.CopyRequestTrace(req, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Request"]).To(BeEmpty())
			})

			It("is successful even if session trace not set", func() {
				delete(req.Env, "TRACE-SESSION")
				Expect(service.CopyRequestTrace(req, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})
		})
	})
})
