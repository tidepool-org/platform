package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	testConfig "github.com/tidepool-org/platform/config/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully", func() {
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		It("returns default values", func() {
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Address).To(Equal(""))
			Expect(config.UserAgent).To(Equal(""))
		})
	})

	Context("with new config", func() {
		var address string
		var userAgent string
		var config *client.Config

		BeforeEach(func() {
			address = testHTTP.NewAddress()
			userAgent = testHTTP.NewUserAgent()
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *testConfig.Reporter

			BeforeEach(func() {
				configReporter = testConfig.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses default address if not set", func() {
				delete(configReporter.Config, "address")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal(""))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHTTP.NewUserAgent()
				config.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.UserAgent).To(Equal(existingUserAgent))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal(address))
				Expect(config.UserAgent).To(Equal(userAgent))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Address = address
				config.UserAgent = userAgent
			})

			Context("Validate", func() {
				It("returns an error if the address is missing", func() {
					config.Address = ""
					Expect(config.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if the address is not a parseable URL", func() {
					config.Address = "Not%Parseable"
					Expect(config.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the user agent is missing", func() {
					config.UserAgent = ""
					Expect(config.Validate()).To(MatchError("user agent is missing"))
				})

				It("returns success", func() {
					Expect(config.Validate()).To(Succeed())
					Expect(config.Address).To(Equal(address))
					Expect(config.UserAgent).To(Equal(userAgent))
				})
			})
		})
	})
})
