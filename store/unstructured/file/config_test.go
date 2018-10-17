package file_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	storeUnstructuredFile "github.com/tidepool-org/platform/store/unstructured/file"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns successfully with default values", func() {
			cfg := storeUnstructuredFile.NewConfig()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Directory).To(BeEmpty())
		})
	})

	Context("with new config", func() {
		var directory string
		var cfg *storeUnstructuredFile.Config

		BeforeEach(func() {
			directory = test.RandomTemporaryDirectory()
			cfg = storeUnstructuredFile.NewConfig()
			Expect(cfg).ToNot(BeNil())
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
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns successfully and does not set the directory", func() {
				delete(configReporter.Config, "directory")
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Directory).To(BeEmpty())
			})

			It("returns successfully and sets the directory", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Directory).To(Equal(directory))
			})
		})

		Context("Validate", func() {
			BeforeEach(func() {
				cfg.Directory = directory
			})

			It("returns an error if the directory is missing", func() {
				cfg.Directory = ""
				Expect(cfg.Validate()).To(MatchError("directory is missing"))
			})

			It("returns successfully", func() {
				Expect(cfg.Validate()).To(Succeed())
			})
		})
	})
})
