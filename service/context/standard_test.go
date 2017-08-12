package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service/context"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Standard", func() {
	Context("with request and response", func() {
		var request *rest.Request
		var response *testRest.ResponseWriter

		BeforeEach(func() {
			request = testRest.NewRequest()
			response = testRest.NewResponseWriter()
		})

		Context("NewStandard", func() {
			It("returns an error if the response is missing", func() {
				responder, err := context.NewStandard(nil, request)
				Expect(err).To(MatchError("context: response is missing"))
				Expect(responder).To(BeNil())
			})

			It("returns an error if the request is missing", func() {
				responder, err := context.NewStandard(response, nil)
				Expect(err).To(MatchError("context: request is missing"))
				Expect(responder).To(BeNil())
			})

			It("is successful", func() {
				Expect(context.NewStandard(response, request)).ToNot(BeNil())
			})
		})
	})
})
