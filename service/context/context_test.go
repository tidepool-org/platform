package context_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/ant0ine/go-json-rest/rest"

// 	testAuth "github.com/tidepool-org/platform/auth/test"
// 	"github.com/tidepool-org/platform/service/context"
// 	testRest "github.com/tidepool-org/platform/test/rest"
// )

// var _ = Describe("Context", func() {
// 	Context("with request and response", func() {
// 		var authDetails *testAuth.Details
// 		var response *testRest.ResponseWriter
// 		var request *rest.Request

// 		BeforeEach(func() {
// 			authDetails = testAuth.NewDetails()
// 			response = testRest.NewResponseWriter()
// 			request = testRest.NewRequest()
// 			request.Env["AUTH-DETAILS"] = authDetails
// 		})

// 		Context("New", func() {
// 			It("returns an error if it fails", func() {
// 				ctx, err := context.New(nil, nil)
// 				Expect(err).To(HaveOccurred())
// 				Expect(ctx).To(BeNil())
// 			})

// 			It("is successful", func() {
// 				Expect(context.New(response, request)).ToNot(BeNil())
// 			})
// 		})

// 		Context("with new context", func() {
// 			var ctx *context.Context

// 			BeforeEach(func() {
// 				var err error
// 				ctx, err = context.New(response, request)
// 				Expect(err).ToNot(HaveOccurred())
// 				Expect(ctx).ToNot(BeNil())
// 			})
// 		})
// 	})
// })
