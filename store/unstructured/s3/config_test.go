package s3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	storeUnstructuredS3 "github.com/tidepool-org/platform/store/unstructured/s3"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully with default values", func() {
			Expect(storeUnstructuredS3.NewConfig()).To(Equal(&storeUnstructuredS3.Config{}))
		})
	})

	Context("with new config", func() {
		var bucket string
		var prefix string
		var config *storeUnstructuredS3.Config

		BeforeEach(func() {
			bucket = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
			prefix = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
			config = storeUnstructuredS3.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["bucket"] = bucket
				configReporter.Config["prefix"] = prefix
			})

			It("returns an error if the config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns successfully and does not set the bucket or prefix", func() {
				delete(configReporter.Config, "bucket")
				delete(configReporter.Config, "prefix")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Bucket).To(BeEmpty())
				Expect(config.Prefix).To(BeEmpty())
			})

			It("returns successfully and does not set the bucket", func() {
				delete(configReporter.Config, "bucket")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Bucket).To(BeEmpty())
				Expect(config.Prefix).To(Equal(prefix))
			})

			It("returns successfully and does not set the prefix", func() {
				delete(configReporter.Config, "prefix")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Bucket).To(Equal(bucket))
				Expect(config.Prefix).To(BeEmpty())
			})

			It("returns successfully and sets the bucket and prefix", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Bucket).To(Equal(bucket))
				Expect(config.Prefix).To(Equal(prefix))
			})
		})

		Context("Validate", func() {
			BeforeEach(func() {
				config.Bucket = bucket
				config.Prefix = prefix
			})

			It("returns an error if the bucket is missing", func() {
				config.Bucket = ""
				Expect(config.Validate()).To(MatchError("bucket is missing"))
			})

			It("returns an error if the prefix is invalid", func() {
				config.Prefix = ""
				Expect(config.Validate()).To(MatchError("prefix is invalid"))
			})

			It("returns successfully", func() {
				Expect(config.Validate()).To(Succeed())
			})
		})
	})
})
