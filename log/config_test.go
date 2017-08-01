package log_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			config := log.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Level).To(Equal("warn"))
		})
	})

	Context("with new config", func() {
		var config *log.Config

		BeforeEach(func() {
			config = log.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *test.Reporter

			BeforeEach(func() {
				configReporter = test.NewReporter()
				configReporter.Config["level"] = "debug"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("log: config reporter is missing"))
			})

			It("uses default level if not set", func() {
				delete(configReporter.Config, "level")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Level).To(Equal("warn"))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Level).To(Equal("debug"))
			})
		})

		Context("Validate", func() {
			It("returns success for the debug level", func() {
				config.Level = "debug"
				Expect(config.Validate()).To(Succeed())
			})

			It("returns success for the info level", func() {
				config.Level = "info"
				Expect(config.Validate()).To(Succeed())
			})

			It("returns success for the warn level", func() {
				config.Level = "warn"
				Expect(config.Validate()).To(Succeed())
			})

			It("returns success for the error level", func() {
				config.Level = "error"
				Expect(config.Validate()).To(Succeed())
			})

			It("returns success for the fatal level", func() {
				config.Level = "fatal"
				Expect(config.Validate()).To(Succeed())
			})

			It("returns an error for the panic level", func() {
				config.Level = "panic"
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("log: level is invalid"))
			})

			It("returns an error for any other invalid level", func() {
				config.Level = "invalid"
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("log: level is invalid"))
			})

			It("returns an error for missing level", func() {
				config.Level = ""
				err := config.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("log: level is invalid"))
			})
		})

		Context("Clone", func() {
			It("returns successfully", func() {
				config.Level = "debug"
				clone := config.Clone()
				Expect(clone).ToNot(BeIdenticalTo(config))
				Expect(clone.Level).To(Equal(config.Level))
			})
		})
	})
})
