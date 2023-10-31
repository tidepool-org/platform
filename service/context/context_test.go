package context_test

import (
	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/context"
	serviceTest "github.com/tidepool-org/platform/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Context", func() {
	Context("with request and response", func() {
		var details request.Details
		var res *testRest.ResponseWriter
		var req *rest.Request

		BeforeEach(func() {
			details = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken(), "patient")
			res = testRest.NewResponseWriter()
			req = testRest.NewRequest()
			req.Env["AUTH-DETAILS"] = details
		})

		AfterEach(func() {
			res.AssertOutputsEmpty()
		})

		Context("New", func() {
			It("returns an error if it fails", func() {
				ctx, err := context.New(nil, nil)
				Expect(err).To(MatchError("response is missing"))
				Expect(ctx).To(BeNil())
			})

			It("is successful", func() {
				Expect(context.New(res, req)).ToNot(BeNil())
			})
		})
	})
})
