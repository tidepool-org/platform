package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/task/service/api"
	testService "github.com/tidepool-org/platform/task/service/test"
)

var _ = Describe("Router", func() {
	var svc *testService.Service

	BeforeEach(func() {
		svc = testService.NewService()
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			rtr, err := api.NewRouter(nil)
			Expect(err).To(MatchError("api: service is missing"))
			Expect(rtr).To(BeNil())
		})

		It("returns successfully", func() {
			rtr, err := api.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var rtr *api.Router

		BeforeEach(func() {
			var err error
			rtr, err = api.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(rtr.Routes()).ToNot(BeEmpty())
			})
		})
	})
})
