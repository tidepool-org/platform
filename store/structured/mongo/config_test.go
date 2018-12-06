package mongo_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			cfg := storeStructuredMongo.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Addresses).To(BeNil())
			Expect(cfg.TLS).To(BeTrue())
			Expect(cfg.Database).To(BeEmpty())
			Expect(cfg.CollectionPrefix).To(BeEmpty())
			Expect(cfg.Username).To(BeNil())
			Expect(cfg.Password).To(BeNil())
			Expect(cfg.Timeout).To(Equal(60 * time.Second))
		})
	})

	Context("with new config", func() {
		var cfg *storeStructuredMongo.Config

		BeforeEach(func() {
			cfg = storeStructuredMongo.NewConfig()
			Expect(cfg).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["addresses"] = "https://1.2.3.4:5678, http://a.b.c.d:9999"
				configReporter.Config["tls"] = "false"
				configReporter.Config["database"] = "database"
				configReporter.Config["collection_prefix"] = "collection_prefix"
				configReporter.Config["username"] = "username"
				configReporter.Config["password"] = "password"
				configReporter.Config["timeout"] = "120"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses default addresses if not set", func() {
				delete(configReporter.Config, "addresses")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Addresses).To(BeEmpty())
			})

			It("uses default tls if not set", func() {
				delete(configReporter.Config, "tls")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.TLS).To(BeTrue())
			})

			It("returns an error if the tls cannot be parsed to a boolean", func() {
				configReporter.Config["tls"] = "abc"
				Expect(cfg.Load(configReporter)).To(MatchError("tls is invalid"))
				Expect(cfg.TLS).To(BeTrue())
			})

			It("uses default database if not set", func() {
				delete(configReporter.Config, "database")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Database).To(BeEmpty())
			})

			It("uses default collection prefix if not set", func() {
				delete(configReporter.Config, "collection_prefix")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.CollectionPrefix).To(BeEmpty())
			})

			It("uses default username if not set", func() {
				delete(configReporter.Config, "username")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Username).To(BeNil())
			})

			It("uses default password if not set", func() {
				delete(configReporter.Config, "password")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Password).To(BeNil())
			})

			It("uses default timeout if not set", func() {
				delete(configReporter.Config, "timeout")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Timeout).To(Equal(60 * time.Second))
			})

			It("returns an error if the timeout cannot be parsed to an integer", func() {
				configReporter.Config["timeout"] = "abc"
				Expect(cfg.Load(configReporter)).To(MatchError("timeout is invalid"))
				Expect(cfg.Timeout).To(Equal(60 * time.Second))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Addresses).To(Equal([]string{"https://1.2.3.4:5678", "http://a.b.c.d:9999"}))
				Expect(cfg.TLS).To(BeFalse())
				Expect(cfg.Database).To(Equal("database"))
				Expect(cfg.CollectionPrefix).To(Equal("collection_prefix"))
				Expect(cfg.Username).ToNot(BeNil())
				Expect(*cfg.Username).To(Equal("username"))
				Expect(cfg.Password).ToNot(BeNil())
				Expect(*cfg.Password).To(Equal("password"))
				Expect(cfg.Timeout).To(Equal(120 * time.Second))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				cfg.Addresses = []string{"1.2.3.4", "5.6.7.8"}
				cfg.TLS = false
				cfg.Database = "database"
				cfg.CollectionPrefix = "collection_prefix"
				cfg.Username = pointer.FromString("username")
				cfg.Password = pointer.FromString("password")
				cfg.Timeout = 5 * time.Second
			})

			Context("Validate", func() {
				It("return success if all are valid", func() {
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns an error if the addresses is nil", func() {
					cfg.Addresses = nil
					Expect(cfg.Validate()).To(MatchError("addresses is missing"))
				})

				It("returns an error if the addresses is empty", func() {
					cfg.Addresses = []string{}
					Expect(cfg.Validate()).To(MatchError("addresses is missing"))
				})

				It("returns an error if one of the addresses is missing", func() {
					cfg.Addresses = []string{""}
					Expect(cfg.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if one of the addresses is not a parseable URL", func() {
					cfg.Addresses = []string{"Not%Parseable"}
					Expect(cfg.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the database is missing", func() {
					cfg.Database = ""
					Expect(cfg.Validate()).To(MatchError("database is missing"))
				})

				It("returns success if the username is not specified", func() {
					cfg.Username = nil
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns success if the password is not specified", func() {
					cfg.Password = nil
					Expect(cfg.Validate()).To(Succeed())
				})

				It("returns an error if the timeout is invalid", func() {
					cfg.Timeout = 0
					Expect(cfg.Validate()).To(MatchError("timeout is invalid"))
				})
			})
		})
	})

	Context("SplitAddresses", func() {
		DescribeTable("returns expected addresses when",
			func(addressesString string, expectedAddresses []string) {
				Expect(storeStructuredMongo.SplitAddresses(addressesString)).To(Equal(expectedAddresses))
			},
			Entry("has empty addresses string with no separator", "", []string{}),
			Entry("has whitespace-only addresses string with no separator", "   ", []string{}),
			Entry("has addresses string with only separators", ",,,", []string{}),
			Entry("has whitespace-only addresses string with separators", "  ,,   ,, ", []string{}),
			Entry("has non-whitespace addresses string with no separator", "alpha", []string{"alpha"}),
			Entry("has addresses string with whitespace no separator", "  alpha  ", []string{"alpha"}),
			Entry("has addresses string with separators", "alpha,beta,charlie", []string{"alpha", "beta", "charlie"}),
			Entry("has addresses string with whitespace and separators", "  alpha   ,  beta, charlie    ", []string{"alpha", "beta", "charlie"}),
		)
	})
})
