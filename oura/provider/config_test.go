package provider_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	ouraProvider "github.com/tidepool-org/platform/oura/provider"
	ouraProviderTest "github.com/tidepool-org/platform/oura/provider/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("config", func() {
	var expectedConfig *ouraProvider.Config
	var configReporter *configTest.Reporter

	BeforeEach(func() {
		expectedConfig = ouraProviderTest.RandomConfig()
		expectedConfig.Provider.AuthStyleInParams = true
		expectedConfig.Provider.CookieDisabled = true
		configReporter = configTest.NewReporter()
		configReporter.Config = ouraProviderTest.NewObjectFromConfig(expectedConfig, test.ObjectFormatConfig)
	})

	Context("NewConfigWithConfigReporter", func() {
		It("returns an error if the config reporter is missing", func() {
			cfg, err := ouraProvider.NewConfigWithConfigReporter(nil)
			Expect(err).To(MatchError("config reporter is missing"))
			Expect(cfg).To(BeNil())
		})

		It("returns successfully", func() {
			cfg, err := ouraProvider.NewConfigWithConfigReporter(configReporter)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg).To(Equal(expectedConfig))
		})
	})

	Context("NewConfig", func() {
		It("returns successfully", func() {
			cfg := ouraProvider.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Config).ToNot(BeNil())
		})
	})

	Context("with config", func() {
		var cfg *ouraProvider.Config

		BeforeEach(func() {
			cfg = ouraProvider.NewConfig()
		})

		Context("LoadFromConfigReporter", func() {
			It("returns an error if the config reporter is missing", func() {
				err := cfg.LoadFromConfigReporter(nil)
				Expect(err).To(MatchError("config reporter is missing"))
				Expect(cfg).To(Equal(ouraProvider.NewConfig()))
			})

			It("returns successfully", func() {
				err := cfg.LoadFromConfigReporter(configReporter)
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).To(Equal(expectedConfig))
			})
		})

		Context("configured", func() {
			BeforeEach(func() {
				Expect(cfg.LoadFromConfigReporter(configReporter)).ToNot(HaveOccurred())
				Expect(cfg).To(Equal(expectedConfig))
			})

			Context("Validate", func() {
				It("returns an error if the config is missing", func() {
					cfg.Config = nil
					Expect(cfg.Validate()).To(MatchError("config is missing"))
				})

				It("returns an error if the config is invalid", func() {
					cfg.Provider.ClientID = ""
					Expect(cfg.Validate()).To(MatchError("config is invalid; provider is invalid; client id is empty"))
				})

				It("returns an error if the provider accept url is missing", func() {
					cfg.Provider.AcceptURL = nil
					Expect(cfg.Validate()).To(MatchError("config is invalid; provider is invalid; accept url is missing"))
				})

				It("returns an error if the provider revoke url is missing", func() {
					cfg.Provider.RevokeURL = nil
					Expect(cfg.Validate()).To(MatchError("config is invalid; provider is invalid; revoke url is missing"))
				})

				It("returns an error if the partner url is missing", func() {
					cfg.PartnerURL = ""
					Expect(cfg.Validate()).To(MatchError("partner url is missing"))
				})

				It("returns an error if the partner url is invalid", func() {
					cfg.PartnerURL = ":::"
					Expect(cfg.Validate()).To(MatchError("partner url is invalid"))
				})

				It("returns an error if the partner secret is missing", func() {
					cfg.PartnerSecret = ""
					Expect(cfg.Validate()).To(MatchError("partner secret is missing"))
				})

				It("returns successfully", func() {
					Expect(cfg.Validate()).ToNot(HaveOccurred())
				})
			})
		})
	})
})
