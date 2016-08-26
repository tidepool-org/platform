package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/environment"
)

type TestConfig struct {
	String  string
	Integer int
	Float   float64
}

var _ = Describe("Config", func() {
	var environmentReporter environment.Reporter

	BeforeEach(func() {
		var err error
		environmentReporter, err = environment.NewReporter("test", "TIDEPOOL_TEST")
		Expect(err).ToNot(HaveOccurred())
		Expect(environmentReporter).ToNot(BeNil())
	})

	Context("NewLoader", func() {
		It("returns an error if environment reporter is missing", func() {
			loader, err := config.NewLoader(nil, "_fixtures/config", "TIDEPOOL_TEST")
			Expect(err).To(MatchError("config: environment reporter is missing"))
			Expect(loader).To(BeNil())
		})

		It("returns an error if directory is missing", func() {
			loader, err := config.NewLoader(environmentReporter, "", "TIDEPOOL_TEST")
			Expect(err).To(MatchError("config: directory is missing"))
			Expect(loader).To(BeNil())
		})

		It("returns an error if prefix is missing", func() {
			loader, err := config.NewLoader(environmentReporter, "_fixtures/config", "")
			Expect(err).To(MatchError("config: prefix is missing"))
			Expect(loader).To(BeNil())
		})

		It("returns an error if directory does not exist", func() {
			loader, err := config.NewLoader(environmentReporter, "_fixtures/config/missing", "TIDEPOOL_TEST")
			Expect(err).To(MatchError("config: directory does not exist"))
			Expect(loader).To(BeNil())
		})

		It("returns an error if directory is a file", func() {
			loader, err := config.NewLoader(environmentReporter, "_fixtures/config/directory.json/file", "TIDEPOOL_TEST")
			Expect(err).To(MatchError("config: directory is not a directory"))
			Expect(loader).To(BeNil())
		})

		It("returns a new object if name is specified", func() {
			Expect(config.NewLoader(environmentReporter, "_fixtures/config", "TIDEPOOL_TEST")).ToNot(BeNil())
		})
	})

	Context("Load", func() {
		var loader config.Loader
		var testConfig *TestConfig

		BeforeEach(func() {
			var err error
			loader, err = config.NewLoader(environmentReporter, "_fixtures/config", "TIDEPOOL_TEST")
			Expect(err).ToNot(HaveOccurred())
			Expect(loader).ToNot(BeNil())
			testConfig = &TestConfig{}
		})

		It("returns an error if name is missing", func() {
			Expect(loader.Load("", testConfig)).To(MatchError("config: name is missing"))
		})

		It("returns an error if config is missing", func() {
			Expect(loader.Load("basic", nil)).To(MatchError("config: config is missing"))
		})

		It("returns an error if the file is a directory", func() {
			Expect(loader.Load("directory", testConfig)).To(MatchError("config: file is a directory"))
		})

		It("returns an error if the file does not contain JSON", func() {
			err := loader.Load("empty", testConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("config: unable to load config;"))
		})

		It("returns an error if the file contains malformed JSON", func() {
			err := loader.Load("malformed", testConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("config: unable to load config;"))
		})

		It("returns an error if the file contains a JSON array", func() {
			err := loader.Load("array", testConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("config: unable to load config;"))
		})

		It("successfully initialized the config without a file", func() {
			Expect(loader.Load("missing", testConfig)).To(Succeed())
			Expect(testConfig.String).To(Equal(""))
			Expect(testConfig.Integer).To(Equal(0))
			Expect(testConfig.Float).To(Equal(0.0))
		})

		It("successfully reads a single file config", func() {
			Expect(loader.Load("single", testConfig)).To(Succeed())
			Expect(testConfig.String).To(Equal("single"))
			Expect(testConfig.Integer).To(Equal(123))
			Expect(testConfig.Float).To(Equal(1.23))
		})

		It("successfully reads a multiple file config", func() {
			Expect(loader.Load("multiple", testConfig)).To(Succeed())
			Expect(testConfig.String).To(Equal("multiple"))
			Expect(testConfig.Integer).To(Equal(456))
			Expect(testConfig.Float).To(Equal(4.56))
		})
	})
})
