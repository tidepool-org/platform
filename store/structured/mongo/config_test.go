package mongo_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	platformConfig "github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/store/structured/mongo"
)

var _ = Describe("Config", func() {
	scheme := "mongodb+srv"
	addresses := []string{"https://1.2.3.4:5678", "http://a.b.c.d:9999"}
	tls := true
	database := "tp_database"
	altDatabase := "tp_alt_database"
	collectionPrefix := "tp_collection_prefix"
	username := "tp_username"
	password := "tp_password"
	timeout := time.Duration(120) * time.Second
	optParams := "replicaSet=Cluster0-shard-0&authSource=admin&w=majority"

	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			datum := mongo.NewConfig()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Scheme).To(BeEmpty())
			Expect(datum.Addresses).To(BeNil())
			Expect(datum.TLS).To(BeTrue())
			Expect(datum.Database).To(BeEmpty())
			Expect(datum.CollectionPrefix).To(BeEmpty())
			Expect(datum.Username).To(BeNil())
			Expect(datum.Password).To(BeNil())
			Expect(datum.Timeout).To(Equal(30 * time.Second))
		})
	})

	Context("Load", func() {
		var config *mongo.Config
		var variables = []string{
			"TIDEPOOL_STORE_SCHEME",
			"TIDEPOOL_STORE_TLS",
			"TIDEPOOL_STORE_DATABASE",
			"TIDEPOOL_STORE_ADDRESSES",
			"TIDEPOOL_STORE_COLLECTION_PREFIX",
			"TIDEPOOL_STORE_USERNAME",
			"TIDEPOOL_STORE_PASSWORD",
			"TIDEPOOL_STORE_TIMEOUT",
			"TIDEPOOL_STORE_OPT_PARAMS",
		}
		var existingEnvVars map[string]string

		BeforeEach(func() {
			existingEnvVars = make(map[string]string)
			for _, v := range variables {
				existingEnvVars[v] = os.Getenv(v)
			}

			Expect(os.Setenv("TIDEPOOL_STORE_SCHEME", scheme)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_TLS", fmt.Sprintf("%v", tls))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_DATABASE", database)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_ADDRESSES", strings.Join(addresses, ","))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_COLLECTION_PREFIX", collectionPrefix)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_USERNAME", username)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_PASSWORD", password)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_TIMEOUT", fmt.Sprintf("%vs", int(timeout.Seconds())))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_OPT_PARAMS", optParams)).To(Succeed())

			config = &mongo.Config{}
			Expect(config.Load()).To(Succeed())
		})

		AfterEach(func() {
			existingEnvVars = make(map[string]string)
			for _, v := range variables {
				_ = os.Setenv(v, existingEnvVars[v])
			}

			_ = os.Unsetenv("TIDEPOOL_STORE_SCHEME")
			_ = os.Unsetenv("TIDEPOOL_STORE_ADDRESSES")
			_ = os.Unsetenv("TIDEPOOL_STORE_TLS")
			_ = os.Unsetenv("TIDEPOOL_STORE_DATABASE")
			_ = os.Unsetenv("TIDEPOOL_STORE_COLLECTION_PREFIX")
			_ = os.Unsetenv("TIDEPOOL_STORE_USERNAME")
			_ = os.Unsetenv("TIDEPOOL_STORE_PASSWORD")
			_ = os.Unsetenv("TIDEPOOL_STORE_TIMEOUT")
			_ = os.Unsetenv("TIDEPOOL_STORE_OPT_PARAMS")
		})

		It("loads scheme from environment", func() {
			Expect(config.Scheme).To(Equal(scheme))
		})

		It("loads addresses from environment", func() {
			Expect(config.Addresses).To(ConsistOf(addresses))
		})

		It("loads tls from environment", func() {
			Expect(config.TLS).To(Equal(tls))
		})

		It("sets tls to 'true' if not found in env", func() {
			Expect(os.Unsetenv("TIDEPOOL_STORE_TLS")).To(Succeed())
			config = &mongo.Config{}
			Expect(config.Load()).To(Succeed())
			Expect(config.TLS).To(Equal(true))
		})

		It("loads database from environment", func() {
			Expect(config.Database).To(Equal(database))
		})

		It("loads collection prefix from environment", func() {
			Expect(config.CollectionPrefix).To(Equal(collectionPrefix))
		})

		It("loads username from environment", func() {
			Expect(config.Username).ToNot(BeNil())
			Expect(*config.Username).To(Equal(username))
		})

		It("loads password from environment", func() {
			Expect(config.Password).ToNot(BeNil())
			Expect(*config.Password).To(Equal(password))
		})

		It("loads timeout from environment", func() {
			Expect(config.Timeout).To(Equal(timeout))
		})

		It("uses default timeout of 60 seconds if timeout not found in env", func() {
			Expect(os.Unsetenv("TIDEPOOL_STORE_TIMEOUT")).To(Succeed())
			config = &mongo.Config{}
			Expect(config.Load()).To(Succeed())
			Expect(config.Timeout).To(Equal(time.Second * time.Duration(60)))
		})

		It("loads optional params from environment", func() {
			Expect(config.OptParams).ToNot(BeNil())
			Expect(*config.OptParams).To(Equal(optParams))
		})
	})

	Context("SetDatabaseFromReporter", func() {
		var config *mongo.Config
		var reporter platformConfig.Reporter

		BeforeEach(func() {
			config = &mongo.Config{}
			var err error
			reporter, err = env.NewDefaultReporter()
			Expect(err).ToNot(HaveOccurred())
			reporter = reporter.WithScopes("alt", "store")
			Expect(err).ToNot(HaveOccurred())
		})

		It("loads database from environment", func() {
			Expect(os.Setenv("TIDEPOOL_ALT_STORE_DATABASE", altDatabase)).To(Succeed())
			Expect(config.SetDatabaseFromReporter(reporter)).To(Succeed())
			Expect(config.Database).To(Equal(altDatabase))
			_ = os.Unsetenv("TIDEPOOL_ALT_STORE_DATABASE")
		})

		It("errors if database not set in environment", func() {
			reporter := reporter.WithScopes("empty")
			Expect(config.SetDatabaseFromReporter(reporter)).To(MatchError("key \"TIDEPOOL_ALT_STORE_EMPTY_DATABASE\" not found"))
			Expect(config.Database).To(Equal(""))
		})
	})

	Context("Validate", func() {
		var config *mongo.Config

		BeforeEach(func() {
			config = &mongo.Config{
				Addresses:        []string{"www.mongo.com:4321"},
				TLS:              tls,
				Database:         database,
				CollectionPrefix: collectionPrefix,
				Username:         pointer.FromString(username),
				Password:         pointer.FromString(password),
				Timeout:          timeout,
				OptParams:        pointer.FromString("w=majority"),
			}
		})

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

		It("returns an error if one of the addresses is not a parsable URL", func() {
			config.Addresses = []string{"Not%Parsable"}
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

	Context("AsConnectionString", func() {
		var config *mongo.Config

		BeforeEach(func() {
			config = &mongo.Config{
				Scheme:           "mongodb",
				Addresses:        []string{"1.2.3.4:1234", "5.6.7.8:5678"},
				TLS:              true,
				Database:         "database",
				CollectionPrefix: "collection_prefix",
				Username:         pointer.FromString("username"),
				Password:         pointer.FromString("password"),
				Timeout:          5 * time.Second,
				OptParams:        pointer.FromString("w=majority"),
			}
		})

		It("generates correct connection string", func() {
			expected := "mongodb://username:password@1.2.3.4:1234,5.6.7.8:5678/database?ssl=true&w=majority"
			Expect(config.AsConnectionString()).To(Equal(expected))
		})
	})
})
