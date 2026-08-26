package provider_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully with the default client timeout", func() {
			config := oauthProvider.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.ClientTimeout).To(Equal(oauthProvider.ClientTimeoutDefault))
		})
	})

	Context("LoadFromConfigReporter", func() {
		var config *oauthProvider.Config
		var configReporter *configTest.Reporter

		BeforeEach(func() {
			config = oauthProvider.NewConfig()
			configReporter = configTest.NewReporter()
		})

		It("sets the client timeout from a duration string", func() {
			clientTimeout := oauthProviderTest.RandomClientTimeout()
			configReporter.Config["client_timeout"] = clientTimeout.String()
			Expect(config.LoadFromConfigReporter(configReporter)).To(Succeed())
			Expect(config.ClientTimeout).To(Equal(clientTimeout))
		})

		It("sets the client timeout from a bare number as seconds", func() {
			clientTimeoutSeconds := test.RandomIntFromRange(1, 3600)
			configReporter.Config["client_timeout"] = strconv.Itoa(clientTimeoutSeconds)
			Expect(config.LoadFromConfigReporter(configReporter)).To(Succeed())
			Expect(config.ClientTimeout).To(Equal(time.Duration(clientTimeoutSeconds) * time.Second))
		})

		It("keeps the default client timeout when the key is not present", func() {
			Expect(config.LoadFromConfigReporter(configReporter)).To(Succeed())
			Expect(config.ClientTimeout).To(Equal(oauthProvider.ClientTimeoutDefault))
		})

		It("keeps the existing client timeout when the key is not present", func() {
			clientTimeout := oauthProviderTest.RandomClientTimeout()
			config.ClientTimeout = clientTimeout
			Expect(config.LoadFromConfigReporter(configReporter)).To(Succeed())
			Expect(config.ClientTimeout).To(Equal(clientTimeout))
		})

		It("returns an error when the client timeout is invalid", func() {
			configReporter.Config["client_timeout"] = "invalid"
			Expect(config.LoadFromConfigReporter(configReporter)).To(MatchError("client timeout is invalid"))
		})
	})

	Context("Validate", func() {
		var config *oauthProvider.Config

		BeforeEach(func() {
			config = oauthProviderTest.RandomConfig()
		})

		It("returns successfully when the client timeout is positive", func() {
			Expect(config.Validate()).To(Succeed())
		})

		It("returns successfully when the client timeout is the default", func() {
			config.ClientTimeout = oauthProvider.ClientTimeoutDefault
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error when the client timeout is zero", func() {
			config.ClientTimeout = 0
			Expect(config.Validate()).To(MatchError("client timeout is invalid"))
		})

		It("returns an error when the client timeout is negative", func() {
			config.ClientTimeout = -oauthProviderTest.RandomClientTimeout()
			Expect(config.Validate()).To(MatchError("client timeout is invalid"))
		})
	})
})
