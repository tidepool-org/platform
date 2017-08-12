package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/client"
	testConfig "github.com/tidepool-org/platform/config/test"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Address).To(Equal(""))
			Expect(config.Timeout).To(Equal(60 * time.Second))
		})
	})

	Context("with new config", func() {
		var config *client.Config

		BeforeEach(func() {
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *testConfig.Reporter

			BeforeEach(func() {
				configReporter = testConfig.NewReporter()
				configReporter.Config["address"] = "https://1.2.3.4:5678"
				configReporter.Config["timeout"] = "120"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("client: config reporter is missing"))
			})

			It("uses default address if not set", func() {
				delete(configReporter.Config, "address")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal(""))
			})

			It("uses default timeout if not set", func() {
				delete(configReporter.Config, "timeout")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns an error if the timeout cannot be parsed to an integer", func() {
				configReporter.Config["timeout"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("client: timeout is invalid"))
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal("https://1.2.3.4:5678"))
				Expect(config.Timeout).To(Equal(120 * time.Second))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Address = "http://localhost:1234"
				config.Timeout = 30 * time.Second
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

				It("returns an error if the timeout is invalid", func() {
					config.Timeout = 0
					Expect(config.Validate()).To(MatchError("client: timeout is invalid"))
				})

				It("returns success", func() {
					Expect(config.Validate()).To(Succeed())
					Expect(config.Address).To(Equal("http://localhost:1234"))
					Expect(config.Timeout).To(Equal(30 * time.Second))
				})
			})
		})
	})
})
