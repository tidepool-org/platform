package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/context"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Context", func() {
	Context("with request and response", func() {
		var details request.Details
		var res *testRest.ResponseWriter
		var req *rest.Request

		BeforeEach(func() {
			details = request.NewDetails(request.MethodSessionToken, id.New(), testAuth.NewSessionToken())
			res = testRest.NewResponseWriter()
			req = testRest.NewRequest()
			req.Env["AUTH-DETAILS"] = details
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
