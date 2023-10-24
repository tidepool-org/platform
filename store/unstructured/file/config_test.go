package file_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	storeUnstructuredFile "github.com/tidepool-org/platform/store/unstructured/file"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully with default values", func() {
			Expect(storeUnstructuredFile.NewConfig()).To(Equal(&storeUnstructuredFile.Config{}))
		})
	})

	Context("with new config", func() {
		var directory string
		var config *storeUnstructuredFile.Config

		BeforeEach(func() {
			directory = test.RandomTemporaryDirectory()
			config = storeUnstructuredFile.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		AfterEach(func() {
			if directory != "" {
				Expect(os.Remove(directory)).To(Succeed())
			}
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
				configReporter.Config["directory"] = directory
			})

			It("returns an error if the config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns successfully and does not set the directory", func() {
				delete(configReporter.Config, "directory")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Directory).To(BeEmpty())
			})

			It("returns successfully and sets the directory", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Directory).To(Equal(directory))
			})
		})

		Context("Validate", func() {
			BeforeEach(func() {
				config.Directory = directory
			})

			It("returns an error if the directory is missing", func() {
				config.Directory = ""
				Expect(config.Validate()).To(MatchError("directory is missing"))
			})

			It("returns successfully", func() {
				Expect(config.Validate()).To(Succeed())
			})
		})
	})
})
