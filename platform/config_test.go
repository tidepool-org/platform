package platform_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strconv"
	"time"

	testConfig "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/platform"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully", func() {
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
		})

		It("returns default values", func() {
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			Expect(config.Config.Address).To(Equal(""))
			Expect(config.Config.UserAgent).To(Equal(""))
			Expect(config.Timeout).To(Equal(60 * time.Second))
		})
	})

	Context("with new config", func() {
		var address string
		var userAgent string
		var timeout int
		var config *platform.Config

		BeforeEach(func() {
			address = testHTTP.NewAddress()
			userAgent = testHTTP.NewUserAgent()
			timeout = testHTTP.NewTimeout()
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *testConfig.Reporter

			BeforeEach(func() {
				configReporter = testConfig.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
				configReporter.Config["timeout"] = strconv.Itoa(timeout)
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses default address if not set", func() {
				delete(configReporter.Config, "address")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Config).ToNot(BeNil())
				Expect(config.Config.Address).To(Equal(""))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHTTP.NewUserAgent()
				config.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Config).ToNot(BeNil())
				Expect(config.Config.UserAgent).To(Equal(existingUserAgent))
			})

			It("uses default timeout if not set", func() {
				delete(configReporter.Config, "timeout")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns an error if the timeout cannot be parsed to an integer", func() {
				configReporter.Config["timeout"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("timeout is invalid"))
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Config).ToNot(BeNil())
				Expect(config.Config.Address).To(Equal(address))
				Expect(config.Config.UserAgent).To(Equal(userAgent))
				Expect(config.Timeout).To(Equal(time.Duration(timeout) * time.Second))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Config.Address = address
				config.Config.UserAgent = userAgent
				config.Timeout = time.Duration(timeout) * time.Second
			})

			Context("Validate", func() {
				It("returns an error if the address is missing", func() {
					config.Config.Address = ""
					Expect(config.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if the address is not a parseable URL", func() {
					config.Config.Address = "Not%Parseable"
					Expect(config.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the user agent is missing", func() {
					config.Config.UserAgent = ""
					Expect(config.Validate()).To(MatchError("user agent is missing"))
				})

				It("returns an error if the timeout is invalid", func() {
					config.Timeout = 0
					Expect(config.Validate()).To(MatchError("timeout is invalid"))
				})

				It("returns success", func() {
					Expect(config.Validate()).To(Succeed())
					Expect(config.Config).ToNot(BeNil())
					Expect(config.Config.Address).To(Equal(address))
					Expect(config.Config.UserAgent).To(Equal(userAgent))
					Expect(config.Timeout).To(Equal(time.Duration(timeout) * time.Second))
				})
			})
		})
	})
})
