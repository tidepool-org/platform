package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/store/mongo"
)

var _ = Describe("Config", func() {
	Context("Validate", func() {
		var username string
		var password string
		var timeout time.Duration
		var config *mongo.Config

		BeforeEach(func() {
			username = "username"
			password = "password"
			timeout = 5 * time.Second
			config = &mongo.Config{
				Addresses:  "1.2.3.4, 5.6.7.8",
				Database:   "database",
				Collection: "collection",
				Username:   &username,
				Password:   &password,
				Timeout:    &timeout,
				SSL:        true,
			}
		})

		It("return success if all are valid", func() {
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the addresses is missing", func() {
			config.Addresses = ""
			Expect(config.Validate()).To(MatchError("mongo: addresses is missing"))
		})

		It("returns an error if the addresses has no non-whitespace entries", func() {
			config.Addresses = "  ,   ,  "
			Expect(config.Validate()).To(MatchError("mongo: addresses is missing"))
		})

		It("returns an error if the database is missing", func() {
			config.Database = ""
			Expect(config.Validate()).To(MatchError("mongo: database is missing"))
		})

		It("returns an error if the collection is missing", func() {
			config.Collection = ""
			Expect(config.Validate()).To(MatchError("mongo: collection is missing"))
		})

		It("returns success if the username is not specified", func() {
			config.Username = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the username is empty", func() {
			username = ""
			Expect(config.Validate()).To(MatchError("mongo: username is empty"))
		})

		It("returns success if the password is not specified", func() {
			config.Password = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the password is empty", func() {
			password = ""
			Expect(config.Validate()).To(MatchError("mongo: password is empty"))
		})

		It("returns success if the timeout is not specified", func() {
			config.Timeout = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the timeout is invalid", func() {
			timeout = -1
			Expect(config.Validate()).To(MatchError("mongo: timeout is invalid"))
		})
	})
})
