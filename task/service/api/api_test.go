package api_test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	taskService "github.com/tidepool-org/platform/task/service"
	taskServiceApi "github.com/tidepool-org/platform/task/service/api"
	"github.com/tidepool-org/platform/task/service/taskservicetest"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("API", func() {
	var service *taskservicetest.Service

	BeforeEach(func() {
		service = taskservicetest.NewService()
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			router, err := taskServiceApi.NewRouter(nil)
			Expect(err).To(MatchError("service is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			router, err := taskServiceApi.NewRouter(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *taskServiceApi.Router

		BeforeEach(func() {
			var err error
			router, err = taskServiceApi.NewRouter(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).ToNot(BeEmpty())
			})
		})

		Context("StatusGet", func() {
			var response *testRest.ResponseWriter
			var request *rest.Request

			BeforeEach(func() {
				response = testRest.NewResponseWriter()
				request = testRest.NewRequest()
				request.Request = request.WithContext(log.NewContextWithLogger(request.Context(), logTest.NewLogger()))
				service = taskservicetest.NewService()
				var err error
				router, err = taskServiceApi.NewRouter(service)
				Expect(err).ToNot(HaveOccurred())
				Expect(router).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(service.UnusedOutputsCount()).To(Equal(0))
				response.AssertOutputsEmpty()
			})

			Context("StatusGet", func() {
				It("panics if response is missing", func() {
					Expect(func() { router.StatusGet(nil, request) }).To(Panic())
				})

				It("panics if request is missing", func() {
					Expect(func() { router.StatusGet(response, nil) }).To(Panic())
				})

				Context("with service status", func() {
					var status *taskService.Status

					BeforeEach(func() {
						status = &taskService.Status{}
						service.StatusOutputs = []*taskService.Status{status}
						response.HeaderOutput = &http.Header{}
						response.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
					})

					It("returns successfully", func() {
						router.StatusGet(response, request)
						Expect(response.WriteInputs).To(HaveLen(1))
						Expect(response.WriteInputs[0]).To(MatchJSON(`{"Version": "", "Server": null, "TaskStore": null}`))
					})
				})
			})
		})
	})
})
