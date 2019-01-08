package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	notificationServiceApiV1 "github.com/tidepool-org/platform/notification/service/api/v1"
	testService "github.com/tidepool-org/platform/notification/service/test"
)

var _ = Describe("Router", func() {
	var svc *testService.Service

	BeforeEach(func() {
		svc = testService.NewService()
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			rtr, err := notificationServiceApiV1.NewRouter(nil)
			Expect(err).To(MatchError("service is missing"))
			Expect(rtr).To(BeNil())
		})

		It("returns successfully", func() {
			rtr, err := notificationServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var rtr *notificationServiceApiV1.Router

		BeforeEach(func() {
			var err error
			rtr, err = notificationServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(rtr.Routes()).To(BeEmpty())
			})
		})
	})
})
