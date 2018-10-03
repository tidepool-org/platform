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
			cfg := storeUnstructuredS3.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Bucket).To(BeEmpty())
			Expect(cfg.Prefix).To(BeEmpty())
		})
	})

	Context("with new config", func() {
		var bucket string
		var prefix string
		var cfg *storeUnstructuredS3.Config

		BeforeEach(func() {
			bucket = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
			prefix = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
			cfg = storeUnstructuredS3.NewConfig()
			Expect(cfg).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["bucket"] = bucket
				configReporter.Config["prefix"] = prefix
			})

			It("returns an error if the config reporter is missing", func() {
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns successfully and does not set the bucket or prefix", func() {
				delete(configReporter.Config, "bucket")
				delete(configReporter.Config, "prefix")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Bucket).To(BeEmpty())
				Expect(cfg.Prefix).To(BeEmpty())
			})

			It("returns successfully and does not set the bucket", func() {
				delete(configReporter.Config, "bucket")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Bucket).To(BeEmpty())
				Expect(cfg.Prefix).To(Equal(prefix))
			})

			It("returns successfully and does not set the prefix", func() {
				delete(configReporter.Config, "prefix")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Bucket).To(Equal(bucket))
				Expect(cfg.Prefix).To(BeEmpty())
			})

			It("returns successfully and sets the bucket and prefix", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Bucket).To(Equal(bucket))
				Expect(cfg.Prefix).To(Equal(prefix))
			})
		})

		Context("Validate", func() {
			BeforeEach(func() {
				cfg.Bucket = bucket
				cfg.Prefix = prefix
			})

			It("returns an error if the bucket is missing", func() {
				cfg.Bucket = ""
				Expect(cfg.Validate()).To(MatchError("bucket is missing"))
			})

			It("returns an error if the prefix is invalid", func() {
				cfg.Prefix = ""
				Expect(cfg.Validate()).To(MatchError("prefix is invalid"))
			})

			It("returns successfully", func() {
				Expect(cfg.Validate()).To(Succeed())
			})
		})
	})
})
