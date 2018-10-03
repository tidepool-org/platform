package factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	awsTest "github.com/tidepool-org/platform/aws/test"
	configTest "github.com/tidepool-org/platform/config/test"
	storeUnstructuredFactory "github.com/tidepool-org/platform/store/unstructured/factory"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Factory", func() {
	var configReporter *configTest.Reporter
	var awsAPI *awsTest.API

	BeforeEach(func() {
		configReporter = configTest.NewReporter()
		awsAPI = awsTest.NewAPI()
	})

	Context("NewStore", func() {
		It("returns an error if the config reporter is missing", func() {
			str, err := storeUnstructuredFactory.NewStore(nil, awsAPI)
			Expect(err).To(MatchError("config reporter is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the aws api is missing", func() {
			str, err := storeUnstructuredFactory.NewStore(configReporter, nil)
			Expect(err).To(MatchError("aws api is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the type is missing", func() {
			str, err := storeUnstructuredFactory.NewStore(configReporter, awsAPI)
			Expect(err).To(MatchError("type is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the type is empty", func() {
			configReporter.Set("type", "")
			str, err := storeUnstructuredFactory.NewStore(configReporter, awsAPI)
			Expect(err).To(MatchError("type is empty"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the type is invalid", func() {
			configReporter.Set("type", "invalid")
			str, err := storeUnstructuredFactory.NewStore(configReporter, awsAPI)
			Expect(err).To(MatchError("type is invalid"))
			Expect(str).To(BeNil())
		})

		Context("with type file", func() {
			var directory string

			BeforeEach(func() {
				directory = test.RandomTemporaryDirectory()
				configReporter.Config["type"] = "file"
				configReporter.Config["file"] = map[string]interface{}{
					"directory": directory,
				}
			})

			AfterEach(func() {
				if directory != "" {
					Expect(os.Remove(directory)).To(Succeed())
				}
			})

			It("returns an error if the config is invalid", func() {
				delete(configReporter.Config, "file")
				str, err := storeUnstructuredFactory.NewStore(configReporter, awsAPI)
				Expect(err).To(MatchError("config is invalid; directory is missing"))
				Expect(str).To(BeNil())
			})

			It("returns an error if the aws api is invalid", func() {
				str, err := storeUnstructuredFactory.NewStore(configReporter, nil)
				Expect(err).To(MatchError("aws api is missing"))
				Expect(str).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredFactory.NewStore(configReporter, awsAPI)).ToNot(BeNil())
			})
		})

		Context("with type s3", func() {
			BeforeEach(func() {
				configReporter.Config["type"] = "s3"
				configReporter.Config["s3"] = map[string]interface{}{
					"bucket": test.NewVariableString(1, 64, test.CharsetAlphaNumeric),
					"prefix": test.NewVariableString(1, 64, test.CharsetAlphaNumeric),
				}
			})

			It("returns an error if the config is invalid", func() {
				delete(configReporter.Config, "s3")
				str, err := storeUnstructuredFactory.NewStore(configReporter, awsAPI)
				Expect(err).To(MatchError("config is invalid; bucket is missing"))
				Expect(str).To(BeNil())
			})

			It("returns an error if the aws api is invalid", func() {
				str, err := storeUnstructuredFactory.NewStore(configReporter, nil)
				Expect(err).To(MatchError("aws api is missing"))
				Expect(str).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(storeUnstructuredFactory.NewStore(configReporter, awsAPI)).ToNot(BeNil())
			})
		})
	})

	Context("NewFileStore", func() {
		var directory string

		BeforeEach(func() {
			directory = test.RandomTemporaryDirectory()
			configReporter.Config["directory"] = directory
		})

		AfterEach(func() {
			if directory != "" {
				Expect(os.Remove(directory)).To(Succeed())
			}
		})

		It("returns an error if the config reporter is missing", func() {
			str, err := storeUnstructuredFactory.NewFileStore(nil)
			Expect(err).To(MatchError("unable to load config; config reporter is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the config is invalid", func() {
			delete(configReporter.Config, "directory")
			str, err := storeUnstructuredFactory.NewFileStore(configReporter)
			Expect(err).To(MatchError("config is invalid; directory is missing"))
			Expect(str).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(storeUnstructuredFactory.NewFileStore(configReporter)).ToNot(BeNil())
		})
	})

	Context("NewS3Store", func() {
		BeforeEach(func() {
			configReporter.Config["bucket"] = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
			configReporter.Config["prefix"] = test.NewVariableString(1, 64, test.CharsetAlphaNumeric)
		})

		It("returns an error if the config reporter is missing", func() {
			str, err := storeUnstructuredFactory.NewS3Store(nil, awsAPI)
			Expect(err).To(MatchError("unable to load config; config reporter is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the aws api is missing", func() {
			str, err := storeUnstructuredFactory.NewS3Store(configReporter, nil)
			Expect(err).To(MatchError("aws api is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the config is invalid", func() {
			delete(configReporter.Config, "bucket")
			str, err := storeUnstructuredFactory.NewS3Store(configReporter, awsAPI)
			Expect(err).To(MatchError("config is invalid; bucket is missing"))
			Expect(str).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(storeUnstructuredFactory.NewS3Store(configReporter, awsAPI)).ToNot(BeNil())
		})
	})
})
