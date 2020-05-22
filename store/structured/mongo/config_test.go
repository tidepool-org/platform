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
			datum := storeStructuredMongo.NewConfig()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Scheme).To(BeEmpty())
			Expect(datum.Addresses).To(BeNil())
			Expect(datum.TLS).To(BeTrue())
			Expect(datum.Database).To(BeEmpty())
			Expect(datum.CollectionPrefix).To(BeEmpty())
			Expect(datum.Username).To(BeNil())
			Expect(datum.Password).To(BeNil())
			Expect(datum.Timeout).To(Equal(60 * time.Second))
		})
	})

	Context("with new config", func() {
		var config *storeStructuredMongo.Config

		BeforeEach(func() {
			config = storeStructuredMongo.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["addresses"] = "https://1.2.3.4:5678, http://a.b.c.d:9999"
				configReporter.Config["tls"] = "false"
				configReporter.Config["scheme"] = "mongodb+srv"
				configReporter.Config["database"] = "database"
				configReporter.Config["collection_prefix"] = "collection_prefix"
				configReporter.Config["username"] = "username"
				configReporter.Config["password"] = "password"
				configReporter.Config["timeout"] = "120"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses default addresses if not set", func() {
				delete(configReporter.Config, "addresses")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Addresses).To(BeEmpty())
			})

			It("uses default tls if not set", func() {
				delete(configReporter.Config, "tls")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.TLS).To(BeTrue())
			})

			It("returns an error if the tls cannot be parsed to a boolean", func() {
				configReporter.Config["tls"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("tls is invalid"))
				Expect(config.TLS).To(BeTrue())
			})

			It("uses default database if not set", func() {
				delete(configReporter.Config, "database")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Database).To(BeEmpty())
			})

			It("uses default collection prefix if not set", func() {
				delete(configReporter.Config, "collection_prefix")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.CollectionPrefix).To(BeEmpty())
			})

			It("uses default username if not set", func() {
				delete(configReporter.Config, "username")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Username).To(BeNil())
			})

			It("uses default password if not set", func() {
				delete(configReporter.Config, "password")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Password).To(BeNil())
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
				Expect(config.Scheme).To(Equal("mongodb+srv"))
				Expect(config.Addresses).To(Equal([]string{"https://1.2.3.4:5678", "http://a.b.c.d:9999"}))
				Expect(config.TLS).To(BeFalse())
				Expect(config.Database).To(Equal("database"))
				Expect(config.CollectionPrefix).To(Equal("collection_prefix"))
				Expect(config.Username).ToNot(BeNil())
				Expect(*config.Username).To(Equal("username"))
				Expect(config.Password).ToNot(BeNil())
				Expect(*config.Password).To(Equal("password"))
				Expect(config.Timeout).To(Equal(120 * time.Second))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Addresses = []string{"1.2.3.4", "5.6.7.8"}
				config.TLS = false
				config.Database = "database"
				config.CollectionPrefix = "collection_prefix"
				config.Username = pointer.FromString("username")
				config.Password = pointer.FromString("password")
				config.Timeout = 5 * time.Second
			})

			Context("Validate", func() {
				It("return success if all are valid", func() {
					Expect(config.Validate()).To(Succeed())
				})

				It("returns an error if the addresses is nil", func() {
					config.Addresses = nil
					Expect(config.Validate()).To(MatchError("addresses is missing"))
				})

				It("returns an error if the addresses is empty", func() {
					config.Addresses = []string{}
					Expect(config.Validate()).To(MatchError("addresses is missing"))
				})

				It("returns an error if one of the addresses is missing", func() {
					config.Addresses = []string{""}
					Expect(config.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if one of the addresses is not a parseable URL", func() {
					config.Addresses = []string{"Not%Parseable"}
					Expect(config.Validate()).To(MatchError("address is invalid"))
				})

				It("returns an error if the database is missing", func() {
					config.Database = ""
					Expect(config.Validate()).To(MatchError("database is missing"))
				})

				It("returns success if the username is not specified", func() {
					config.Username = nil
					Expect(config.Validate()).To(Succeed())
				})

				It("returns success if the password is not specified", func() {
					config.Password = nil
					Expect(config.Validate()).To(Succeed())
				})

				It("returns an error if the timeout is invalid", func() {
					config.Timeout = 0
					Expect(config.Validate()).To(MatchError("timeout is invalid"))
				})
			})
		})
	})

	Context("AsConnectionString", func() {
		var config *storeStructuredMongo.Config

		BeforeEach(func() {
			config = storeStructuredMongo.NewConfig()
			config.Scheme = "mongodb"
			config.Addresses = []string{"1.2.3.4:1234", "5.6.7.8:5678"}
			config.TLS = true
			config.Database = "database"
			config.CollectionPrefix = "collection_prefix"
			config.Username = pointer.FromString("username")
			config.Password = pointer.FromString("password")
			config.Timeout = 5 * time.Second
			config.OptParams = pointer.FromString("w=majority")
		})

		It("generates correct connection string", func() {
			expected := "mongodb://username:password@1.2.3.4:1234,5.6.7.8:5678/database?ssl=true&w=majority"
			Expect(config.AsConnectionString()).To(Equal(expected))
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
