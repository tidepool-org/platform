package api_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("API", func() {
	// 	var svc *testService.Service

	// 	BeforeEach(func() {
	// 		svc = testService.NewService()
	// 	})

	// 	Context("NewRouter", func() {
	// 		It("returns an error if context is missing", func() {
	// 			rtr, err := api.NewRouter(nil)
	// 			Expect(err).To(MatchError("service is missing"))
	// 			Expect(rtr).To(BeNil())
	// 		})

	// 		It("returns successfully", func() {
	// 			rtr, err := api.NewRouter(svc)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(rtr).ToNot(BeNil())
	// 		})
	// 	})

	// 	Context("with new router", func() {
	// 		var rtr *api.Router

	// 		BeforeEach(func() {
	// 			var err error
	// 			rtr, err = api.NewRouter(svc)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(rtr).ToNot(BeNil())
	// 		})

	// 		Context("Routes", func() {
	// 			It("returns the expected routes", func() {
	// 				Expect(rtr.Routes()).ToNot(BeEmpty())
	// 			})
	// 		})

	// 		var _ = Describe("StatusGet", func() {
	// 			var response *testRest.ResponseWriter
	// 			var request *rest.Request
	// 			var svc *testService.Service
	// 			var rtr *api.Router

	// 			BeforeEach(func() {
	// 				response = testRest.NewResponseWriter()
	// 				request = testRest.NewRequest()
	// 				svc = testService.NewService()
	// 				var err error
	// 				rtr, err = api.NewRouter(svc)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(rtr).ToNot(BeNil())
	// 			})

	// 			AfterEach(func() {
	// 				Expect(svc.UnusedOutputsCount()).To(Equal(0))
	// 				Expect(response.UnusedOutputsCount()).To(Equal(0))
	// 			})

	// 			Context("StatusGet", func() {
	// 				It("panics if response is missing", func() {
	// 					Expect(func() { rtr.StatusGet(nil, request) }).To(Panic())
	// 				})

	// 				It("panics if request is missing", func() {
	// 					Expect(func() { rtr.StatusGet(response, nil) }).To(Panic())
	// 				})

	// 				Context("with service status", func() {
	// 					var sts *service.Status

	// 					BeforeEach(func() {
	// 						sts = &service.Status{}
	// 						svc.StatusOutputs = []*service.Status{sts}
	// 						response.WriteJsonOutputs = []error{nil}
	// 					})

	// 					It("returns successfully", func() {
	// 						rtr.StatusGet(response, request)
	// 						Expect(response.WriteJsonInputs).To(HaveLen(1))
	// 						Expect(response.WriteJsonInputs[0].(*serviceContext.JSONResponse).Data).To(Equal(sts))
	// 					})
	// 				})
	// 			})
	// 		})
	// 	})
})
