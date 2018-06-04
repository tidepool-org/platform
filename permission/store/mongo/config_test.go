package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/permission/store/mongo"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			config := mongo.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Secret).To(BeEmpty())
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
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns an error if base config is missing", func() {
				config.Config = nil
				Expect(config.Load(configReporter)).To(MatchError("config is missing"))
			})

			It("returns an error if base config returns an error", func() {
				configReporter.Config["tls"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("tls is invalid"))
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
				config.Config = storeStructuredMongo.NewConfig()
				config.Addresses = []string{"1.2.3.4", "5.6.7.8"}
				config.TLS = false
				config.Database = "database"
				config.CollectionPrefix = "collection_prefix"
				config.Username = pointer.FromString("username")
				config.Password = pointer.FromString("password")
				config.Timeout = 5 * time.Second
				config.Secret = "super"
			})

			Context("Validate", func() {
				It("return success if all are valid", func() {
					Expect(config.Validate()).To(Succeed())
				})

				It("returns an error if the base config is missing", func() {
					config.Config = nil
					Expect(config.Validate()).To(MatchError("config is missing"))
				})

				It("returns an error if the base config is not valid", func() {
					config.Addresses = nil
					Expect(config.Validate()).To(MatchError("addresses is missing"))
				})

				It("returns an error if the secret is missing", func() {
					config.Secret = ""
					Expect(config.Validate()).To(MatchError("secret is missing"))
				})
			})
		})
	})
})
