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
			Expect(cfg.ClientTimeout).To(BeZero())
			Expect(cfg.ResponseTimeout).To(BeZero())
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
		var clientTimeoutSeconds int
		var clientTimeout time.Duration
		var responseTimeoutSeconds int
		var responseTimeout time.Duration
		var cfg *client.Config

		BeforeEach(func() {
			address = testHttp.NewAddress()
			userAgent = testHttp.NewUserAgent()
			clientTimeoutSeconds = test.RandomIntFromRange(1, 3600)
			clientTimeout = time.Duration(clientTimeoutSeconds) * time.Second
			responseTimeoutSeconds = test.RandomIntFromRange(1, 3600)
			responseTimeout = time.Duration(responseTimeoutSeconds) * time.Second
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
				configReporter.Config["client_timeout"] = strconv.Itoa(clientTimeoutSeconds)
				configReporter.Config["response_timeout"] = strconv.Itoa(responseTimeoutSeconds)
				loader = client.NewConfigReporterLoader(configReporter)
			})

			It("uses existing address if not set", func() {
				existingAddress := testHttp.NewAddress()
				cfg.Address = existingAddress
				delete(configReporter.Config, "address")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(existingAddress))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing client timeout if not set", func() {
				existingClientTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.ClientTimeout = existingClientTimeout
				delete(configReporter.Config, "client_timeout")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(existingClientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing response timeout if not set", func() {
				existingResponseTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.ResponseTimeout = existingResponseTimeout
				delete(configReporter.Config, "response_timeout")
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(existingResponseTimeout))
			})

			It("interprets a client timeout without units as seconds", func() {
				configReporter.Config["client_timeout"] = "1.5"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ClientTimeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a response timeout without units as seconds", func() {
				configReporter.Config["response_timeout"] = "1.5"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ResponseTimeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a client timeout with units as a duration", func() {
				configReporter.Config["client_timeout"] = "1m30s"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ClientTimeout).To(Equal(90 * time.Second))
			})

			It("interprets a response timeout with units as a duration", func() {
				configReporter.Config["response_timeout"] = "1m30s"
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ResponseTimeout).To(Equal(90 * time.Second))
			})

			It("returns an error if the client timeout is not a number or a duration", func() {
				configReporter.Config["client_timeout"] = "invalid"
				Expect(cfg.Load(loader)).To(MatchError("client timeout is invalid"))
			})

			It("returns an error if the client timeout is empty", func() {
				configReporter.Config["client_timeout"] = ""
				Expect(cfg.Load(loader)).To(MatchError("client timeout is invalid"))
			})

			It("returns an error if the response timeout is not a number or a duration", func() {
				configReporter.Config["response_timeout"] = "invalid"
				Expect(cfg.Load(loader)).To(MatchError("response timeout is invalid"))
			})

			It("returns an error if the response timeout is empty", func() {
				configReporter.Config["response_timeout"] = ""
				Expect(cfg.Load(loader)).To(MatchError("response timeout is invalid"))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})
		})

		Context("Load with envconfig loader", func() {
			var loader client.ConfigLoader

			BeforeEach(func() {
				loader = client.NewEnvconfigLoader()
			})

			It("returns successfully and uses the client timeout from the environment", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_CLIENT_TIMEOUT", clientTimeout.String())
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(BeZero())
			})

			It("returns successfully and uses the response timeout from the environment", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_RESPONSE_TIMEOUT", responseTimeout.String())
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ClientTimeout).To(BeZero())
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("ignores the untagged environment variables", func() {
				GinkgoT().Setenv("CLIENT_TIMEOUT", clientTimeout.String())
				GinkgoT().Setenv("RESPONSE_TIMEOUT", responseTimeout.String())
				Expect(cfg.Load(loader)).To(Succeed())
				Expect(cfg.ClientTimeout).To(BeZero())
				Expect(cfg.ResponseTimeout).To(BeZero())
			})

			// Unlike the config reporter loaders, envconfig requires units
			It("returns an error if the client timeout has no units", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_CLIENT_TIMEOUT", strconv.Itoa(clientTimeoutSeconds))
				Expect(cfg.Load(loader)).To(MatchError(ContainSubstring("assigning TIDEPOOL_CLIENT_CLIENT_TIMEOUT to ClientTimeout")))
			})

			// Unlike the config reporter loaders, envconfig requires units
			It("returns an error if the response timeout has no units", func() {
				GinkgoT().Setenv("TIDEPOOL_CLIENT_RESPONSE_TIMEOUT", strconv.Itoa(responseTimeoutSeconds))
				Expect(cfg.Load(loader)).To(MatchError(ContainSubstring("assigning TIDEPOOL_CLIENT_RESPONSE_TIMEOUT to ResponseTimeout")))
			})
		})

		Context("LoadFromConfigReporter", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["address"] = address
				configReporter.Config["user_agent"] = userAgent
				configReporter.Config["client_timeout"] = strconv.Itoa(clientTimeoutSeconds)
				configReporter.Config["response_timeout"] = strconv.Itoa(responseTimeoutSeconds)
			})

			It("uses existing address if not set", func() {
				existingAddress := testHttp.NewAddress()
				cfg.Address = existingAddress
				delete(configReporter.Config, "address")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(existingAddress))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing user agent if not set", func() {
				existingUserAgent := testHttp.NewUserAgent()
				cfg.UserAgent = existingUserAgent
				delete(configReporter.Config, "user_agent")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(existingUserAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing client timeout if not set", func() {
				existingClientTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.ClientTimeout = existingClientTimeout
				delete(configReporter.Config, "client_timeout")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(existingClientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})

			It("uses existing response timeout if not set", func() {
				existingResponseTimeout := time.Duration(test.RandomIntFromRange(1, 3600)) * time.Second
				cfg.ResponseTimeout = existingResponseTimeout
				delete(configReporter.Config, "response_timeout")
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(existingResponseTimeout))
			})

			It("interprets a client timeout without units as seconds", func() {
				configReporter.Config["client_timeout"] = "1.5"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.ClientTimeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a response timeout without units as seconds", func() {
				configReporter.Config["response_timeout"] = "1.5"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.ResponseTimeout).To(Equal(1500 * time.Millisecond))
			})

			It("interprets a client timeout with units as a duration", func() {
				configReporter.Config["client_timeout"] = "1m30s"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.ClientTimeout).To(Equal(90 * time.Second))
			})

			It("interprets a response timeout with units as a duration", func() {
				configReporter.Config["response_timeout"] = "1m30s"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.ResponseTimeout).To(Equal(90 * time.Second))
			})

			It("returns an error if the client timeout is not a number or a duration", func() {
				configReporter.Config["client_timeout"] = "invalid"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("client timeout is invalid"))
			})

			It("returns an error if the client timeout is empty", func() {
				configReporter.Config["client_timeout"] = ""
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("client timeout is invalid"))
			})

			It("returns an error if the response timeout is not a number or a duration", func() {
				configReporter.Config["response_timeout"] = "invalid"
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("response timeout is invalid"))
			})

			It("returns an error if the response timeout is empty", func() {
				configReporter.Config["response_timeout"] = ""
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(MatchError("response timeout is invalid"))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.LoadFromConfigReporter(configReporter)).To(Succeed())
				Expect(cfg.Address).To(Equal(address))
				Expect(cfg.UserAgent).To(Equal(userAgent))
				Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
				Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				cfg.Address = address
				cfg.UserAgent = userAgent
				cfg.ClientTimeout = clientTimeout
				cfg.ResponseTimeout = responseTimeout
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

				It("returns an error if the client timeout is negative", func() {
					cfg.ClientTimeout = -clientTimeout
					Expect(cfg.Validate()).To(MatchError("client timeout is invalid"))
				})

				It("returns success if the client timeout is zero", func() {
					cfg.ClientTimeout = 0
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns an error if the response timeout is negative", func() {
					cfg.ResponseTimeout = -responseTimeout
					Expect(cfg.Validate()).To(MatchError("response timeout is invalid"))
				})

				It("returns success if the response timeout is zero", func() {
					cfg.ResponseTimeout = 0
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns success", func() {
					Expect(cfg.Validate()).To(Succeed())
					Expect(cfg.Address).To(Equal(address))
					Expect(cfg.UserAgent).To(Equal(userAgent))
					Expect(cfg.ClientTimeout).To(Equal(clientTimeout))
					Expect(cfg.ResponseTimeout).To(Equal(responseTimeout))
				})
			})
		})
	})
})
