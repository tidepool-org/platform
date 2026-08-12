package client_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/test"
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
			Expect(cfg.Timeout).To(BeZero())
		})
	})

	Context("NewConfigReporterLoader", func() {
		It("returns successfully", func() {
			Expect(client.NewConfigReporterLoader(configTest.NewReporter())).ToNot(BeNil())
		})
	})

	Context("NewEnvconfigLoader", func() {
		It("returns successfully", func() {
			Expect(client.NewEnvconfigLoader()).ToNot(BeNil())
		})
	})

	Context("with new config", func() {
		var address string
		var userAgent string
		var timeoutSeconds int
		var timeout time.Duration
		var cfg *client.Config

		BeforeEach(func() {
			address = testHttp.NewAddress()
			userAgent = testHttp.NewUserAgent()
			timeoutSeconds = test.RandomIntFromRange(1, 3600)
			timeout = time.Duration(timeoutSeconds) * time.Second
			cfg = client.NewConfig()
			Expect(cfg).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter
			var loader client.ConfigLoader

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
				configReporter.Config["timeout"] = strconv.Itoa(timeoutSeconds)
				loader = client.NewConfigReporterLoader(configReporter)
			})

			It("uses existing address if not set", func() {
				existingAddress := testHttp.NewAddress()
				cfg.Address = existingAddress
				delete(configReporter.Config, "address")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(existingAddress))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})

			It("uses existing timeout if not set", func() {
				existingTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.Timeout = existingTimeout
				delete(configReporter.Config, "timeout")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(existingTimeout))
			})

			It("interprets a timeout without units as seconds", func() {
				configReporter.Config["timeout"] = "1.5"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a timeout with units as a duration", func() {
				configReporter.Config["timeout"] = "1m30s"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(90 * time.Second))
			})

			It("returns an error if the timeout is not a number or a duration", func() {
				configReporter.Config["timeout"] = "invalid"
				Expect(cfg.Load(loader)).To(MatchError("timeout is invalid"))
			})

			It("returns an error if the timeout is empty", func() {
				configReporter.Config["timeout"] = ""
				Expect(cfg.Load(loader)).To(MatchError("timeout is invalid"))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})
		})

		Context("Load with envconfig loader", func() {
			var loader client.ConfigLoader

			BeforeEach(func() {
				loader = client.NewEnvconfigLoader()
			})

			It("returns successfully and uses the timeout from the environment", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_TIMEOUT", timeout.String())
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(timeout))
			})

			It("ignores the untagged TIMEOUT environment variable", func() {
				GinkgoT().Setenv("TIMEOUT", timeout.String())
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Timeout).To(BeZero())
			})

			// Unlike the config reporter loaders, envconfig requires units - see CLIENT-016.
			It("returns an error if the timeout has no units", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_TIMEOUT", strconv.Itoa(timeoutSeconds))
				Expect(cfg.Load(loader)).To(MatchError(ContainSubstring("assigning TIDEPOOL_CLIENT_TIMEOUT to Timeout")))
			})
		})

		Context("LoadFromConfigReporter", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
				configReporter.Config["timeout"] = strconv.Itoa(timeoutSeconds)
			})

			It("uses existing address if not set", func() {
				existingAddress := testHttp.NewAddress()
				cfg.Address = existingAddress
				delete(configReporter.Config, "address")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(existingAddress))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})

			It("uses existing timeout if not set", func() {
				existingTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.Timeout = existingTimeout
				delete(configReporter.Config, "timeout")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(existingTimeout))
			})

			It("interprets a timeout without units as seconds", func() {
				configReporter.Config["timeout"] = "1.5"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a timeout with units as a duration", func() {
				configReporter.Config["timeout"] = "1m30s"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(90 * time.Second))
			})

			It("returns an error if the timeout is not a number or a duration", func() {
				configReporter.Config["timeout"] = "invalid"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("timeout is invalid"))
			})

			It("returns an error if the timeout is empty", func() {
				configReporter.Config["timeout"] = ""
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("timeout is invalid"))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.Timeout).To(Equal(timeout))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				cfg.Address = address
				cfg.UserAgent = userAgent
				cfg.Timeout = timeout
			})

			Context("Validate", func() {
				It("returns an error if the address is missing", func() {
					cfg.Address = ""
					Expect(cfg.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if the address is not a parsable URL", func() {
					cfg.Address = "Not%Parsable"
					Expect(cfg.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the timeout is negative", func() {
					cfg.Timeout = -timeout
					Expect(cfg.Validate()).To(MatchError("timeout is invalid"))
				})

				It("returns success if the timeout is zero", func() {
					cfg.Timeout = 0
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns success", func() {
					Expect(cfg.Validate()).To(Succeed())
					Expect(cfg.Address).To(Equal(address))
					Expect(cfg.UserAgent).To(Equal(userAgent))
					Expect(cfg.Timeout).To(Equal(timeout))
				})
			})
		})
	})
})
