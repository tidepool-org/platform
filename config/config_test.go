package config_test

import (
	"os"

	. "github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/store"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	var _ = Describe("FromJson", func() {

		It("should load the given config file", func() {
			var mgoConfig store.MongoConfig
			FromJson(&mgoConfig, "mongo.json")
			Expect(mgoConfig).To(Not(BeNil()))
			Expect(mgoConfig.Url).To(Not(BeEmpty()))
		})
		It("should error if the config doen't exist", func() {
			var random interface{}
			err := FromJson(&random, "random.json")
			Expect(random).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	var _ = Describe("FromEnv", func() {

		It("should load the given config value from env", func() {
			const platform_key, platform_val = "CONFIG_TEST", "yay I exist!"
			os.Setenv(platform_key, platform_val)

			platfromValue, _ := FromEnv(platform_key)
			Expect(platfromValue).To(Equal(platform_val))

			os.Unsetenv(platform_key)
		})

		It("should error if the value doesn't exist", func() {
			const other_key = "OTHER"
			os.Unsetenv(other_key) // make sure it doesn't exist

			_, err := FromEnv("OTHER")

			Expect(err).To(MatchError("$OTHER must be set"))
		})
	})

})
