package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/permission/store/mongo"
	"github.com/tidepool-org/platform/pointer"
	baseConfig "github.com/tidepool-org/platform/store/mongo"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			config := mongo.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Secret).To(Equal(""))
		})
	})

	Context("with new config", func() {
		var config *mongo.Config

		BeforeEach(func() {
			config = mongo.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *test.Reporter

			BeforeEach(func() {
				configReporter = test.NewReporter()
				configReporter.Config["secret"] = "super"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("mongo: config reporter is missing"))
			})

			It("returns an error if base config is missing", func() {
				config.Config = nil
				Expect(config.Load(configReporter)).To(MatchError("mongo: config is missing"))
			})

			It("returns an error if base config returns an error", func() {
				configReporter.Config["tls"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("mongo: tls is invalid"))
			})

			It("uses default secret if not set", func() {
				delete(configReporter.Config, "secret")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Secret).To(BeEmpty())
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Secret).To(Equal("super"))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Config = baseConfig.NewConfig()
				config.Config.Addresses = []string{"1.2.3.4", "5.6.7.8"}
				config.Config.TLS = false
				config.Config.Database = "database"
				config.Config.Collection = "collection"
				config.Config.Username = pointer.String("username")
				config.Config.Password = pointer.String("password")
				config.Config.Timeout = 5 * time.Second
				config.Secret = "super"
			})

			Context("Validate", func() {
				It("return success if all are valid", func() {
					Expect(config.Validate()).To(Succeed())
				})

				It("returns an error if the base config is missing", func() {
					config.Config = nil
					Expect(config.Validate()).To(MatchError("mongo: config is missing"))
				})

				It("returns an error if the base config is not valid", func() {
					config.Config.Addresses = nil
					Expect(config.Validate()).To(MatchError("mongo: addresses is missing"))
				})

				It("returns an error if the secret is missing", func() {
					config.Secret = ""
					Expect(config.Validate()).To(MatchError("mongo: secret is missing"))
				})
			})
		})
	})
})
