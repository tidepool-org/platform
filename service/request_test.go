package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
)

func NewTestRequest() *rest.Request {
	baseRequest, err := http.NewRequest("GET", "http://127.0.0.1/", nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(baseRequest).ToNot(BeNil())
	return &rest.Request{
		Request:    baseRequest,
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}

var _ = Describe("Request", func() {
	Context("with request", func() {
		var errors []*service.Error
		var logger log.Logger
		var authDetails *testAuth.Details
		var request *rest.Request

		BeforeEach(func() {
			errors = []*service.Error{
				{
					Code:   "test-error-code",
					Status: 400,
					Title:  "test-error-title",
					Detail: "test-error-detail",
				},
			}
			logger = null.NewLogger()
			authDetails = testAuth.NewDetails()
			request = NewTestRequest()
			request.Env["ERRORS"] = errors
			request.Env["LOGGER"] = logger
			request.Env["AUTH-DETAILS"] = authDetails
			request.Env["TRACE-REQUEST"] = "request-raven"
			request.Env["TRACE-SESSION"] = "session-starling"
		})

		Context("GetRequestErrors", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestErrors(request)).To(Equal(errors))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestErrors(nil)).To(BeNil())
			})

			It("returns nil if the errors is not of the correct type", func() {
				request.Env["ERRORS"] = 0
				Expect(service.GetRequestErrors(request)).To(BeNil())
			})

			It("returns nil if the errors is missing", func() {
				request.Env["ERRORS"] = nil
				Expect(service.GetRequestErrors(request)).To(BeNil())
			})
		})

		Context("SetRequestErrors", func() {
			var newErrors []*service.Error

			BeforeEach(func() {
				newErrors = []*service.Error{
					{
						Code:   "test-error-code-new",
						Status: 400,
						Title:  "test-error-title-new",
						Detail: "test-error-detail-new",
					},
				}
			})

			It("successfully sets the errors", func() {
				service.SetRequestErrors(request, newErrors)
				Expect(request.Env["ERRORS"]).To(Equal(newErrors))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestErrors(nil, newErrors)
				Expect(request.Env["ERRORS"]).To(Equal(errors))
			})

			It("deletes the errors if the errors is missing", func() {
				service.SetRequestErrors(request, nil)
				Expect(request.Env["ERRORS"]).To(BeNil())
			})
		})

		Context("GetRequestLogger", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestLogger(request)).To(Equal(logger))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestLogger(nil)).To(BeNil())
			})

			It("returns nil if the logger is not of the correct type", func() {
				request.Env["LOGGER"] = 0
				Expect(service.GetRequestLogger(request)).To(BeNil())
			})

			It("returns nil if the logger is missing", func() {
				request.Env["LOGGER"] = nil
				Expect(service.GetRequestLogger(request)).To(BeNil())
			})
		})

		Context("SetRequestLogger", func() {
			var newLogger log.Logger

			BeforeEach(func() {
				newLogger = null.NewLogger()
			})

			It("successfully sets the logger", func() {
				service.SetRequestLogger(request, newLogger)
				Expect(request.Env["LOGGER"]).To(Equal(newLogger))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestLogger(nil, newLogger)
				Expect(request.Env["LOGGER"]).To(Equal(logger))
			})

			It("deletes the logger if the logger is missing", func() {
				service.SetRequestLogger(request, nil)
				Expect(request.Env["LOGGER"]).To(BeNil())
			})
		})

		Context("GetRequestAuthDetails", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestAuthDetails(request)).To(Equal(authDetails))
			})

			It("returns nil if the request is missing", func() {
				Expect(service.GetRequestAuthDetails(nil)).To(BeNil())
			})

			It("returns nil if the auth details is not of the correct type", func() {
				request.Env["AUTH-DETAILS"] = 0
				Expect(service.GetRequestAuthDetails(request)).To(BeNil())
			})

			It("returns nil if the auth details is missing", func() {
				request.Env["AUTH-DETAILS"] = nil
				Expect(service.GetRequestAuthDetails(request)).To(BeNil())
			})
		})

		Context("SetRequestAuthDetails", func() {
			var newAuthDetails *testAuth.Details

			BeforeEach(func() {
				newAuthDetails = testAuth.NewDetails()
			})

			It("successfully sets the auth details", func() {
				service.SetRequestAuthDetails(request, newAuthDetails)
				Expect(request.Env["AUTH-DETAILS"]).To(Equal(newAuthDetails))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestAuthDetails(nil, newAuthDetails)
				Expect(request.Env["AUTH-DETAILS"]).To(Equal(authDetails))
			})

			It("deletes the auth details if the auth details is missing", func() {
				service.SetRequestAuthDetails(request, nil)
				Expect(request.Env["AUTH-DETAILS"]).To(BeNil())
			})
		})

		Context("GetRequestTraceRequest", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestTraceRequest(request)).To(Equal("request-raven"))
			})

			It("returns empty string if the request is missing", func() {
				Expect(service.GetRequestTraceRequest(nil)).To(BeEmpty())
			})

			It("returns empty string if the request trace is not of the correct type", func() {
				request.Env["TRACE-REQUEST"] = 0
				Expect(service.GetRequestTraceRequest(request)).To(BeEmpty())
			})

			It("returns empty string if the request trace is missing", func() {
				delete(request.Env, "TRACE-REQUEST")
				Expect(service.GetRequestTraceRequest(request)).To(BeEmpty())
			})
		})

		Context("SetRequestTraceRequest", func() {
			It("successfully sets the request trace", func() {
				service.SetRequestTraceRequest(request, "request-raven-new")
				Expect(request.Env["TRACE-REQUEST"]).To(Equal("request-raven-new"))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestTraceRequest(nil, "request-raven-new")
				Expect(request.Env["TRACE-REQUEST"]).To(Equal("request-raven"))
			})

			It("deletes the request trace if the request trace is missing", func() {
				service.SetRequestTraceRequest(request, "")
				Expect(request.Env["TRACE-REQUEST"]).To(BeNil())
			})
		})

		Context("GetRequestTraceSession", func() {
			It("returns successfully", func() {
				Expect(service.GetRequestTraceSession(request)).To(Equal("session-starling"))
			})

			It("returns empty string if the request is missing", func() {
				Expect(service.GetRequestTraceSession(nil)).To(BeEmpty())
			})

			It("returns empty string if the session trace is not of the correct type", func() {
				request.Env["TRACE-SESSION"] = 0
				Expect(service.GetRequestTraceSession(request)).To(BeEmpty())
			})

			It("returns empty string if the session trace is missing", func() {
				delete(request.Env, "TRACE-SESSION")
				Expect(service.GetRequestTraceSession(request)).To(BeEmpty())
			})
		})

		Context("SetRequestTraceSession", func() {
			It("successfully sets the session trace", func() {
				service.SetRequestTraceSession(request, "session-starling-new")
				Expect(request.Env["TRACE-SESSION"]).To(Equal("session-starling-new"))
			})

			It("does nothing if the request is missing", func() {
				service.SetRequestTraceSession(nil, "session-starling-new")
				Expect(request.Env["TRACE-SESSION"]).To(Equal("session-starling"))
			})

			It("deletes the session trace if the session trace is missing", func() {
				service.SetRequestTraceSession(request, "")
				Expect(request.Env["TRACE-SESSION"]).To(BeNil())
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
				Expect(service.CopyRequestTrace(request, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Request"]).To(ConsistOf("request-raven"))
				Expect(destinationRequest.Header["X-Tidepool-Trace-Session"]).To(ConsistOf("session-starling"))
			})

			It("returns an error if the source request is missing", func() {
				Expect(service.CopyRequestTrace(nil, destinationRequest)).To(MatchError("service: source request is missing"))
			})

			It("returns an error if the destination request is missing", func() {
				Expect(service.CopyRequestTrace(request, nil)).To(MatchError("service: destination request is missing"))
			})

			It("is successful even if request trace not set", func() {
				delete(request.Env, "TRACE-REQUEST")
				Expect(service.CopyRequestTrace(request, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Request"]).To(BeEmpty())
			})

			It("is successful even if session trace not set", func() {
				delete(request.Env, "TRACE-SESSION")
				Expect(service.CopyRequestTrace(request, destinationRequest)).To(Succeed())
				Expect(destinationRequest.Header["X-Tidepool-Trace-Session"]).To(BeEmpty())
			})
		})
	})
})
