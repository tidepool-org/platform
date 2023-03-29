package platform_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully", func() {
			cfg := platform.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Config).ToNot(BeNil())
		})

		It("returns default values", func() {
			cfg := platform.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Config).ToNot(BeNil())
			Expect(cfg.Address).To(BeEmpty())
			Expect(cfg.UserAgent).To(BeEmpty())
			Expect(cfg.ServiceSecret).To(BeEmpty())
		})
	})

	Context("with new config", func() {
		var address string
		var userAgent string
		var serviceSecret string
		var cfg *platform.Config

		BeforeEach(func() {
			address = testHttp.NewAddress()
			userAgent = testHttp.NewUserAgent()
			serviceSecret = test.RandomStringFromRangeAndCharset(1, 64, test.CharsetText)
			cfg = platform.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
				configReporter.Config["service_secret"] = serviceSecret
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
				Expect(cfg.ServiceSecret).To(Equal(serviceSecret))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Config).ToNot(BeNil())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
				Expect(cfg.ServiceSecret).To(Equal(serviceSecret))
			})

			It("uses existing service secret if not set", func() {
				existingServiceSecret := test.RandomStringFromRangeAndCharset(1, 64, test.CharsetText)
				cfg.ServiceSecret = existingServiceSecret
				delete(configReporter.Config, "service_secret")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ServiceSecret).To(Equal(existingServiceSecret))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Config).ToNot(BeNil())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ServiceSecret).To(Equal(serviceSecret))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				cfg.Address = address
				cfg.UserAgent = userAgent
				cfg.ServiceSecret = serviceSecret
			})

			Context("Validate", func() {
				It("returns an error if the address is missing", func() {
					cfg.Address = ""
					Expect(cfg.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if the address is not a parseable URL", func() {
					cfg.Address = "Not%Parseable"
					Expect(cfg.Validate()).To(MatchError("address is invalid: parse \"Not%Parseable\": invalid URL escape \"%Pa\""))
				})

				It("returns an error if the user agent is missing", func() {
					cfg.UserAgent = ""
					Expect(cfg.Validate()).To(MatchError("user agent is missing"))
				})

				It("returns success", func() {
					Expect(cfg.Validate()).To(Succeed())
					Expect(cfg.Config).ToNot(BeNil())
					Expect(cfg.Address).To(Equal(address))
					Expect(cfg.UserAgent).To(Equal(userAgent))
					Expect(cfg.ServiceSecret).To(Equal(serviceSecret))
				})
			})
		})
	})
})
