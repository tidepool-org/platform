package mongoofficial_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/store/structured/mongoofficial"
)

var _ = Describe("Config", func() {
	scheme := "mongodb+srv"
	addresses := []string{"https://1.2.3.4:5678", "http://a.b.c.d:9999"}
	tls := false
	database := "tp_database"
	collectionPrefix := "tp_collection_prefix"
	username := "tp_username"
	password := "tp_password"
	timeout := time.Duration(120) * time.Second
	optParams := "safe=1"

	Describe("Load", func() {
		var config *mongoofficial.Config

		BeforeEach(func() {
			Expect(os.Setenv("TIDEPOOL_STORE_SCHEME", scheme)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_TLS", fmt.Sprintf("%v", tls))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_DATABASE", database)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_ADDRESSES", strings.Join(addresses, ","))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_COLLECTION_PREFIX", collectionPrefix)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_USERNAME", username)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_PASSWORD", password)).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_TIMEOUT", fmt.Sprintf("%vs", int(timeout.Seconds())))).To(Succeed())
			Expect(os.Setenv("TIDEPOOL_STORE_OPT_PARAMS", optParams)).To(Succeed())

			config = &mongoofficial.Config{}
			Expect(config.Load()).To(Succeed())
		})

		AfterEach(func() {
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
			Expect(config.TLS).To(Equal(false))
		})

		It("sets tls to 'true' if not found in env", func() {
			Expect(os.Unsetenv("TIDEPOOL_STORE_TLS")).To(Succeed())
			config = &mongoofficial.Config{}
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
			config = &mongoofficial.Config{}
			Expect(config.Load()).To(Succeed())
			Expect(config.Timeout).To(Equal(time.Second * time.Duration(60)))
		})

		It("loads optional params from environment", func() {
			Expect(config.OptParams).ToNot(BeNil())
			Expect(*config.OptParams).To(Equal(optParams))
		})
	})

	Context("Validate", func() {
		var config *mongoofficial.Config

		BeforeEach(func() {
			config = &mongoofficial.Config{
				Addresses:        []string{"www.mongo.com:4321"},
				TLS:              tls,
				Database:         database,
				CollectionPrefix: collectionPrefix,
				Username:         pointer.FromString(username),
				Password:         pointer.FromString(password),
				Timeout:          timeout,
				OptParams:        nil,
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
