package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
)

var _ = Describe("Standard", func() {
	Context("Trace struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				trace := &context.Trace{}
				Expect(json.Marshal(trace)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				trace := &context.Trace{
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
				meta := &context.Meta{}
				Expect(json.Marshal(meta)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				meta := &context.Meta{
					Trace: &context.Trace{
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
				jsonResponse := &context.JSONResponse{}
				Expect(json.Marshal(jsonResponse)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				jsonResponse := &context.JSONResponse{
					Errors: []*service.Error{
						{
							Code:   "test-code",
							Detail: "test-detail",
							Status: 400,
							Title:  "test-title",
						},
					},
					Meta: &context.Meta{
						Trace: &context.Trace{
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
			Expect(context.ErrorInternalServerFailure()).To(Equal(
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
