package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/context"
	testAuth "github.com/tidepool-org/platform/auth/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Context", func() {
	var authClient *testAuth.Client
	var authDetails *testAuth.Details
	var request *rest.Request
	var response *testRest.ResponseWriter

	BeforeEach(func() {
		authClient = testAuth.NewClient()
		authDetails = testAuth.NewDetails()
		request = testRest.NewRequest()
		request.Env["AUTH-DETAILS"] = authDetails
		response = testRest.NewResponseWriter()
	})

	Context("New", func() {
		It("returns an error if the response is missing", func() {
			ctx, err := context.New(nil, request, authClient)
			Expect(err).To(MatchError("context: response is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if the request is missing", func() {
			ctx, err := context.New(response, nil, authClient)
			Expect(err).To(MatchError("context: request is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if the auth client is missing", func() {
			ctx, err := context.New(response, request, nil)
			Expect(err).To(MatchError("context: auth client is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(context.New(response, request, authClient)).ToNot(BeNil())
		})
	})

	Context("with new context", func() {
		var ctx *context.Context

		BeforeEach(func() {
			var err error
			ctx, err = context.New(response, request, authClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(ctx).ToNot(BeNil())
		})

		Context("AuthClient", func() {
			It("returns auth client", func() {
				Expect(ctx.AuthClient()).To(Equal(authClient))
			})
		})

		Context("AuthDetails", func() {
			It("returns auth details", func() {
				Expect(ctx.AuthDetails()).To(Equal(authDetails))
			})

			It("returns nil if none set", func() {
				delete(request.Env, "AUTH-DETAILS")
				Expect(ctx.AuthDetails()).To(BeNil())
			})
		})
	})
})
