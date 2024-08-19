package service_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/service"
)

var _ = Describe("Route", func() {
	Context("MakeRoute", func() {
		It("returns a route with missing parameters", func() {
			route := service.MakeRoute("", "", nil)
			Expect(route.Method).To(BeEmpty())
			Expect(route.Path).To(BeEmpty())
			Expect(route.Handler).To(BeNil())
		})

		It("returns a route matching valid parameters", func() {
			route := service.MakeRoute("GET", "/path/:to/resource", func(context service.Context) {})
			Expect(route.Method).To(Equal("GET"))
			Expect(route.Path).To(Equal("/path/:to/resource"))
		})
	})
})
