package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/permission/store/mongo"
	baseConfig "github.com/tidepool-org/platform/store/mongo"
)

var _ = Describe("Config", func() {
	var config *mongo.Config

	BeforeEach(func() {
		config = &mongo.Config{
			Config: &baseConfig.Config{
				Addresses:  "1.2.3.4, 5.6.7.8",
				Database:   "database",
				Collection: "collection",
				Username:   app.StringAsPointer("username"),
				Password:   app.StringAsPointer("password"),
				Timeout:    app.DurationAsPointer(5 * time.Second),
				SSL:        true,
			},
			Secret: "secret",
		}
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
			config.Config.Addresses = ""
			Expect(config.Validate()).To(MatchError("mongo: addresses is missing"))
		})

		It("returns an error if the secret is missing", func() {
			config.Secret = ""
			Expect(config.Validate()).To(MatchError("mongo: secret is missing"))
		})
	})

	Context("Clone", func() {
		It("returns successfully", func() {
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Config).ToNot(BeIdenticalTo(config.Config))
			Expect(clone.Config).To(Equal(config.Config))
			Expect(clone.Secret).To(Equal(config.Secret))
		})
	})
})
