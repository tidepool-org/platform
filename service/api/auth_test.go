package api_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/ant0ine/go-json-rest/rest"

// 	testAuth "github.com/tidepool-org/platform/auth/test"
// 	"github.com/tidepool-org/platform/service"
// 	"github.com/tidepool-org/platform/service/api"
// 	"github.com/tidepool-org/platform/service/context"
// 	testRest "github.com/tidepool-org/platform/test/rest"
// )

// var _ = Describe("Auth", func() {
// 	var authDetails *testAuth.Details
// 	var response *testRest.ResponseWriter
// 	var request *rest.Request
// 	var handlerFunc rest.HandlerFunc

// 	BeforeEach(func() {
// 		authDetails = testAuth.NewDetails()
// 		response = testRest.NewResponseWriter()
// 		request = testRest.NewRequest()
// 		request.Env["AUTH-DETAILS"] = authDetails
// 		handlerFunc = func(response rest.ResponseWriter, request *rest.Request) {
// 			Expect(response).ToNot(BeNil())
// 			Expect(request).ToNot(BeNil())
// 			response.WriteHeader(0)
// 		}
// 	})

// 	AfterEach(func() {
// 		Expect(response.UnusedOutputsCount()).To(Equal(0))
// 		Expect(authDetails.UnusedOutputsCount()).To(Equal(0))
// 	})

// 	Context("AuthServer", func() {
// 		It("returns successfully", func() {
// 			Expect(api.AuthServer(handlerFunc)).ToNot(BeNil())
// 		})

// 		It("does nothing if handlerFunc is nil", func() {
// 			authHandlerFunc := api.AuthServer(nil)
// 			Expect(authHandlerFunc).ToNot(BeNil())
// 			authHandlerFunc(response, request)
// 			Expect(response.WriteHeaderInputs).To(BeEmpty())
// 		})

// 		Context("with new auth handler func", func() {
// 			var authHandlerFunc rest.HandlerFunc

// 			BeforeEach(func() {
// 				authHandlerFunc = api.AuthServer(handlerFunc)
// 				Expect(authHandlerFunc).ToNot(BeNil())
// 			})

// 			It("does nothing if response is nil", func() {
// 				authHandlerFunc(nil, request)
// 				Expect(response.WriteHeaderInputs).To(BeEmpty())
// 			})

// 			It("does nothing if request is nil", func() {
// 				authHandlerFunc(response, nil)
// 				Expect(response.WriteHeaderInputs).To(BeEmpty())
// 			})

// 			It("responds with unauthenticated error if auth details is missing", func() {
// 				response.WriteJsonOutputs = []error{nil}
// 				delete(request.Env, "AUTH-DETAILS")
// 				authHandlerFunc(response, request)
// 				Expect(response.WriteHeaderInputs).To(Equal([]int{401}))
// 				Expect(response.WriteJsonInputs).To(HaveLen(1))
// 				Expect(response.WriteJsonInputs[0].(*context.JSONResponse).Errors).To(Equal([]*service.Error{service.ErrorUnauthenticated()}))
// 			})

// 			It("responds with unauthenticated error if auth details is not server", func() {
// 				response.WriteJsonOutputs = []error{nil}
// 				authDetails.IsServerOutputs = []bool{false}
// 				authHandlerFunc(response, request)
// 				Expect(response.WriteHeaderInputs).To(Equal([]int{401}))
// 				Expect(response.WriteJsonInputs).To(HaveLen(1))
// 				Expect(response.WriteJsonInputs[0].(*context.JSONResponse).Errors).To(Equal([]*service.Error{service.ErrorUnauthenticated()}))
// 			})

// 			It("responds successfully if auth details is server", func() {
// 				authDetails.IsServerOutputs = []bool{true}
// 				authHandlerFunc(response, request)
// 				Expect(response.WriteHeaderInputs).To(Equal([]int{0}))
// 			})
// 		})
// 	})

// 	Context("AuthUser", func() {
// 		It("returns successfully", func() {
// 			Expect(api.AuthUser(handlerFunc)).ToNot(BeNil())
// 		})

// 		It("does nothing if handlerFunc is nil", func() {
// 			authHandlerFunc := api.AuthUser(nil)
// 			Expect(authHandlerFunc).ToNot(BeNil())
// 			authHandlerFunc(response, request)
// 			Expect(response.WriteHeaderInputs).To(BeEmpty())
// 		})

// 		Context("with new auth handler func", func() {
// 			var authHandlerFunc rest.HandlerFunc

// 			BeforeEach(func() {
// 				authHandlerFunc = api.AuthUser(handlerFunc)
// 				Expect(authHandlerFunc).ToNot(BeNil())
// 			})

// 			It("does nothing if response is nil", func() {
// 				authHandlerFunc(nil, request)
// 				Expect(response.WriteHeaderInputs).To(BeEmpty())
// 			})

// 			It("does nothing if request is nil", func() {
// 				authHandlerFunc(response, nil)
// 				Expect(response.WriteHeaderInputs).To(BeEmpty())
// 			})

// 			It("responds with unauthenticated error if auth details is missing", func() {
// 				response.WriteJsonOutputs = []error{nil}
// 				delete(request.Env, "AUTH-DETAILS")
// 				authHandlerFunc(response, request)
// 				Expect(response.WriteHeaderInputs).To(Equal([]int{401}))
// 				Expect(response.WriteJsonInputs).To(HaveLen(1))
// 				Expect(response.WriteJsonInputs[0].(*context.JSONResponse).Errors).To(Equal([]*service.Error{service.ErrorUnauthenticated()}))
// 			})

// 			It("responds successfully if auth details", func() {
// 				authHandlerFunc(response, request)
// 				Expect(response.WriteHeaderInputs).To(Equal([]int{0}))
// 			})
// 		})
// 	})
// })
