package service_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Context", func() {
	Context("Trace struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				trace := &service.Trace{}
				Expect(json.Marshal(trace)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				trace := &service.Trace{
					Request: "test-request",
					Session: "test-session",
				}
				Expect(json.Marshal(trace)).To(MatchJSON("{" +
					"\"request\":\"test-request\"," +
					"\"session\":\"test-session\"" +
					"}"))
			})
		})
	})

	Context("Meta struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				meta := &service.Meta{}
				Expect(json.Marshal(meta)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				meta := &service.Meta{
					Trace: &service.Trace{
						Request: "test-request",
						Session: "test-session",
					},
				}
				Expect(json.Marshal(meta)).To(MatchJSON("{" +
					"\"trace\":{" +
					"\"request\":\"test-request\"," +
					"\"session\":\"test-session\"" +
					"}}"))
			})
		})
	})

	Context("JSONResponse struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				jsonResponse := &service.JSONResponse{}
				Expect(json.Marshal(jsonResponse)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				jsonResponse := &service.JSONResponse{
					Errors: []*service.Error{
						{
							Code:   "test-code",
							Detail: "test-detail",
							Status: 400,
							Title:  "test-title",
						},
					},
					Meta: &service.Meta{
						Trace: &service.Trace{
							Request: "test-request",
							Session: "test-session",
						},
					},
				}
				Expect(json.Marshal(jsonResponse)).To(MatchJSON("{" +
					"\"errors\":[{" +
					"\"code\":\"test-code\"," +
					"\"detail\":\"test-detail\"," +
					"\"status\":\"400\"," +
					"\"title\":\"test-title\"}]," +
					"\"meta\":{" +
					"\"trace\":{" +
					"\"request\":\"test-request\"," +
					"\"session\":\"test-session\"" +
					"}}}"))
			})
		})
	})

	Context("InternalServerError", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorInternalServerFailure()).To(Equal(
				&service.Error{
					Code:   "internal-server-failure",
					Status: 500,
					Title:  "internal server failure",
					Detail: "Internal server failure",
				}))
		})
	})

	PContext("NewContext", func() {})

	PContext("Context", func() {})
})
