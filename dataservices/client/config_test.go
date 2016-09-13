package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/client"
)

var _ = Describe("Config", func() {
	var config *client.Config

	BeforeEach(func() {
		config = &client.Config{
			Address:        "http://localhost:1234",
			RequestTimeout: 30,
		}
	})

	Context("Validate", func() {
		It("returns an error if the address is missing", func() {
			config.Address = ""
			Expect(config.Validate()).To(MatchError("client: address is missing"))
		})

		It("returns an error if the address is not a parseable URL", func() {
			config.Address = "Not%Parseable"
			Expect(config.Validate()).To(MatchError("client: address is invalid"))
		})

		It("returns an error if the request timeout is less than zero", func() {
			config.RequestTimeout = -1
			Expect(config.Validate()).To(MatchError("client: request timeout is invalid"))
		})

		It("returns success and leaves all valid (non-zero) fields as-is", func() {
			Expect(config.Validate()).To(Succeed())
			Expect(config.Address).To(Equal("http://localhost:1234"))
			Expect(config.RequestTimeout).To(Equal(30))
		})

		It("sets the request timeout to a default of 60", func() {
			config.RequestTimeout = 0
			Expect(config.Validate()).To(Succeed())
			Expect(config.RequestTimeout).To(Equal(60))
		})
	})

	Context("Clone", func() {
		It("returns successfully", func() {
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Address).To(Equal(config.Address))
			Expect(clone.RequestTimeout).To(Equal(config.RequestTimeout))
		})
	})
})
