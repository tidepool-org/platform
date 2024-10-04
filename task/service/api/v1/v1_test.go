package v1_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	taskServiceApiV1 "github.com/tidepool-org/platform/task/service/api/v1"
	"github.com/tidepool-org/platform/task/service/taskservicetest"
)

var _ = Describe("V1", func() {
	var service *taskservicetest.Service

	BeforeEach(func() {
		service = taskservicetest.NewService()
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			router, err := taskServiceApiV1.NewRouter(nil)
			Expect(err).To(MatchError("service is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			router, err := taskServiceApiV1.NewRouter(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *taskServiceApiV1.Router

		BeforeEach(func() {
			var err error
			router, err = taskServiceApiV1.NewRouter(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/tasks")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/tasks")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/tasks/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPut), "PathExp": Equal("/v1/tasks/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/tasks/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/metrics")})),
				))
			})
		})
	})
})
