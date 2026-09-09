package provider_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Config", func() {
	Context("LoadFromConfigReporter", func() {
		It("does not load pkce enabled from the config reporter", func() {
			reporter := configTest.NewReporter()
			reporter.Set("pkce_enabled", "true")
			cfg := oauthProvider.NewConfig()
			Expect(cfg.LoadFromConfigReporter(reporter)).To(Succeed())
			Expect(cfg.PKCEEnabled).To(BeFalse())
		})
	})

	Context("Validate", func() {
		var cfg *oauthProvider.Config

		BeforeEach(func() {
			cfg = oauthProviderTest.RandomConfig()
			cfg.CookieDisabled = true
			cfg.StateSalt = nil
			cfg.PKCEEnabled = true
		})

		It("returns an error when pkce is enabled and the state salt is missing", func() {
			Expect(cfg.Validate()).To(MatchError("state salt is missing"))
		})

		It("returns an error when pkce is enabled and the state salt is empty", func() {
			cfg.StateSalt = pointer.FromString("")
			Expect(cfg.Validate()).To(MatchError("state salt is empty"))
		})

		It("returns successfully when pkce is enabled and the state salt is present", func() {
			cfg.StateSalt = pointer.FromString(oauthProviderTest.RandomStateSalt())
			Expect(cfg.Validate()).To(Succeed())
		})

		It("returns successfully when pkce and cookies are disabled without a state salt", func() {
			cfg.PKCEEnabled = false
			Expect(cfg.Validate()).To(Succeed())
		})
	})
})
