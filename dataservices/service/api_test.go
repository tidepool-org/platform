package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/service"
)

var _ = Describe("API", func() {
	Context("MakeRoute", func() {
		It("returns a route with missing parameters", func() {
			route := service.MakeRoute("", "", nil)
			Expect(route.Method).To(Equal(""))
			Expect(route.Path).To(Equal(""))
			Expect(route.Handler).To(BeNil())
		})

		It("returns a route matching valid parameters", func() {
			route := service.MakeRoute("GET", "/path/:to/resource", func(context service.Context) {})
			Expect(route.Method).To(Equal("GET"))
			Expect(route.Path).To(Equal("/path/:to/resource"))
		})
	})
})
