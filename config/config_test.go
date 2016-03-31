package config_test

import (
	"os"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/store"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("FromJson", func() {

		It("loads the given config file", func() {
			var mongoConfig store.MongoConfig
			config.FromJSON(&mongoConfig, "mongo.json")
			Expect(mongoConfig).To(Not(BeNil()))
			Expect(mongoConfig.URL).To(Not(BeEmpty()))
		})

		It("returns error if the config doen't exist", func() {
			var random interface{}
			err := config.FromJSON(&random, "random.json")
			Expect(random).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("random.json: no such file or directory"))
		})
	})

	Describe("FromEnv", func() {

		It("loads the given config value from env", func() {
			const platformKey, platformValue = "CONFIG_TEST", "yay I exist!"
			os.Setenv(platformKey, platformValue)

			platfromValue, _ := config.FromEnv(platformKey)
			Expect(platfromValue).To(Equal(platformValue))

			os.Unsetenv(platformKey)
		})

		It("returns error if the value doesn't exist", func() {
			const otherKey = "OTHER"
			os.Unsetenv(otherKey)

			_, err := config.FromEnv(otherKey)

			Expect(err).To(MatchError("$OTHER must be set"))
		})
	})

})
