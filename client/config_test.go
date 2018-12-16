package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	configTest "github.com/tidepool-org/platform/config/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully", func() {
			cfg := client.NewConfig()
			Expect(cfg).ToNot(BeNil())
		})

		It("returns default values", func() {
			cfg := client.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Address).To(BeEmpty())
			Expect(cfg.UserAgent).To(BeEmpty())
		})
	})

	Context("with new config", func() {
		var address string
		var userAgent string
		var cfg *client.Config

		BeforeEach(func() {
			address = testHttp.NewAddress()
			userAgent = testHttp.NewUserAgent()
			cfg = client.NewConfig()
			Expect(cfg).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
			})

			It("returns an error if config reporter is missing", func() {
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses existing address if not set", func() {
				existingAddress := testHttp.NewAddress()
				cfg.Address = existingAddress
				delete(configReporter.Config, "address")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(existingAddress))
				Expect(cfg.UserAgent).To(Equal(userAgent))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				cfg.Address = address
				cfg.UserAgent = userAgent
			})

			Context("Validate", func() {
				It("returns an error if the address is missing", func() {
					cfg.Address = ""
					Expect(cfg.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if the address is not a parseable URL", func() {
					cfg.Address = "Not%Parseable"
					Expect(cfg.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the user agent is missing", func() {
					cfg.UserAgent = ""
					Expect(cfg.Validate()).To(MatchError("user agent is missing"))
				})

				It("returns success", func() {
					Expect(cfg.Validate()).To(Succeed())
					Expect(cfg.Address).To(Equal(address))
					Expect(cfg.UserAgent).To(Equal(userAgent))
				})
			})
		})
	})
})
