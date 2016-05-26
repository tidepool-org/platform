package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/userservices/client"
)

var _ = Describe("Config", func() {
	var config *client.Config

	BeforeEach(func() {
		config = &client.Config{
			Address:            "http://localhost:1234",
			RequestTimeout:     30,
			ServerTokenSecret:  " I Have A Better Secret! ",
			ServerTokenTimeout: 1800,
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

		It("returns an error if server token secret is missing", func() {
			config.ServerTokenSecret = ""
			Expect(config.Validate()).To(MatchError("client: server token secret is missing"))
		})

		It("returns an error if the server token timeout is less than zero", func() {
			config.ServerTokenTimeout = -1
			Expect(config.Validate()).To(MatchError("client: server token timeout is invalid"))
		})

		It("returns success and leaves all valid (non-zero) fields as-is", func() {
			Expect(config.Validate()).To(Succeed())
			Expect(config.Address).To(Equal("http://localhost:1234"))
			Expect(config.RequestTimeout).To(Equal(30))
			Expect(config.ServerTokenSecret).To(Equal(" I Have A Better Secret! "))
			Expect(config.ServerTokenTimeout).To(Equal(1800))
		})

		It("sets the request timeout to a default of 60", func() {
			config.RequestTimeout = 0
			Expect(config.Validate()).To(Succeed())
			Expect(config.RequestTimeout).To(Equal(60))
		})

		It("sets the server token timeout to a default of 3600", func() {
			config.ServerTokenTimeout = 0
			Expect(config.Validate()).To(Succeed())
			Expect(config.ServerTokenTimeout).To(Equal(3600))
		})
	})
})
