package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/store/mongo"
)

var _ = Describe("Config", func() {
	var config *mongo.Config

	BeforeEach(func() {
		config = &mongo.Config{
			Addresses:  "1.2.3.4, 5.6.7.8",
			Database:   "database",
			Collection: "collection",
			Username:   app.StringAsPointer("username"),
			Password:   app.StringAsPointer("password"),
			Timeout:    app.DurationAsPointer(5 * time.Second),
			SSL:        true,
		}
	})

	Context("Validate", func() {
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
			config.Username = app.StringAsPointer("")
			Expect(config.Validate()).To(MatchError("mongo: username is empty"))
		})

		It("returns success if the password is not specified", func() {
			config.Password = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the password is empty", func() {
			config.Password = app.StringAsPointer("")
			Expect(config.Validate()).To(MatchError("mongo: password is empty"))
		})

		It("returns success if the timeout is not specified", func() {
			config.Timeout = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the timeout is invalid", func() {
			config.Timeout = app.DurationAsPointer(-1)
			Expect(config.Validate()).To(MatchError("mongo: timeout is invalid"))
		})
	})

	Context("Clone", func() {
		It("returns successfully", func() {
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Addresses).To(Equal(config.Addresses))
			Expect(clone.Database).To(Equal(config.Database))
			Expect(clone.Collection).To(Equal(config.Collection))
			Expect(clone.Username).ToNot(BeIdenticalTo(config.Username))
			Expect(*clone.Username).To(Equal(*config.Username))
			Expect(clone.Password).ToNot(BeIdenticalTo(config.Password))
			Expect(*clone.Password).To(Equal(*config.Password))
			Expect(clone.Timeout).ToNot(BeIdenticalTo(config.Timeout))
			Expect(*clone.Timeout).To(Equal(*config.Timeout))
			Expect(clone.SSL).To(Equal(config.SSL))
		})

		It("returns successfully if username is nil", func() {
			config.Username = nil
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Addresses).To(Equal(config.Addresses))
			Expect(clone.Database).To(Equal(config.Database))
			Expect(clone.Collection).To(Equal(config.Collection))
			Expect(clone.Username).To(BeNil())
			Expect(clone.Password).ToNot(BeIdenticalTo(config.Password))
			Expect(*clone.Password).To(Equal(*config.Password))
			Expect(clone.Timeout).ToNot(BeIdenticalTo(config.Timeout))
			Expect(*clone.Timeout).To(Equal(*config.Timeout))
			Expect(clone.SSL).To(Equal(config.SSL))
		})

		It("returns successfully if password is nil", func() {
			config.Password = nil
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Addresses).To(Equal(config.Addresses))
			Expect(clone.Database).To(Equal(config.Database))
			Expect(clone.Collection).To(Equal(config.Collection))
			Expect(clone.Username).ToNot(BeIdenticalTo(config.Username))
			Expect(*clone.Username).To(Equal(*config.Username))
			Expect(clone.Password).To(BeNil())
			Expect(clone.Timeout).ToNot(BeIdenticalTo(config.Timeout))
			Expect(*clone.Timeout).To(Equal(*config.Timeout))
			Expect(clone.SSL).To(Equal(config.SSL))
		})

		It("returns successfully if timeout is nil", func() {
			config.Timeout = nil
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Addresses).To(Equal(config.Addresses))
			Expect(clone.Database).To(Equal(config.Database))
			Expect(clone.Collection).To(Equal(config.Collection))
			Expect(clone.Username).ToNot(BeIdenticalTo(config.Username))
			Expect(*clone.Username).To(Equal(*config.Username))
			Expect(clone.Password).ToNot(BeIdenticalTo(config.Password))
			Expect(*clone.Password).To(Equal(*config.Password))
			Expect(clone.Timeout).To(BeNil())
			Expect(clone.SSL).To(Equal(config.SSL))
		})
	})
})
