package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/application/version"
	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
)

var _ = Describe("Application", func() {
	Context("New", func() {
		It("returns an error if the name is missing", func() {
			app, err := application.New("", "TIDEPOOL")
			Expect(err).To(MatchError("application: name is missing"))
			Expect(app).To(BeNil())
		})

		It("returns an error if the prefix is missing", func() {
			app, err := application.New("test", "")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(app).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(application.New("test", "TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with new application", func() {
		var app *application.Application

		BeforeEach(func() {
			var err error
			app, err = application.New("test", "TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(app).ToNot(BeNil())
		})

		Context("Initialize", func() {
			Context("with incorrectly specified version", func() {
				var versionBase string

				BeforeEach(func() {
					versionBase = version.Base
					version.Base = ""
				})

				AfterEach(func() {
					version.Base = versionBase
				})

				It("returns an error if the version is not specified correctly", func() {
					Expect(app.Initialize()).To(MatchError("application: unable to create version reporter; version: base is missing"))
				})
			})

			Context("with invalid level", func() {
				var configReporter config.Reporter
				var level string

				BeforeEach(func() {
					var err error
					configReporter, err = env.NewReporter("TIDEPOOL")
					Expect(err).ToNot(HaveOccurred())
					Expect(configReporter).ToNot(BeNil())
					configReporter = configReporter.WithScopes("test", "logger")
					level = configReporter.GetWithDefault("level", "warn")
					configReporter.Set("level", "invalid")
				})

				AfterEach(func() {
					configReporter.Set("level", level)
				})

				It("returns an error if the logger level is invalid", func() {
					Expect(app.Initialize()).To(MatchError("application: unable to create logger; log: level not found"))
				})
			})

			It("returns successfully", func() {
				Expect(app.Initialize()).To(Succeed())
			})
		})

		Context("Terminate", func() {
			It("returns without panic", func() {
				app.Terminate()
			})
		})

		Context("Name", func() {
			It("returns the name", func() {
				Expect(app.Name()).To(Equal("test"))
			})
		})

		Context("Prefix", func() {
			It("returns the prefix", func() {
				Expect(app.Prefix()).To(Equal("TIDEPOOL"))
			})
		})

		Context("initialized", func() {
			BeforeEach(func() {
				Expect(app.Initialize()).To(Succeed())
			})

			AfterEach(func() {
				app.Terminate()
			})

			Context("VersionReporter", func() {
				It("returns not nil", func() {
					Expect(app.VersionReporter()).ToNot(BeNil())
				})
			})

			Context("ConfigReporter", func() {
				It("returns not nil", func() {
					Expect(app.ConfigReporter()).ToNot(BeNil())
				})
			})

			Context("Logger", func() {
				It("returns not nil", func() {
					Expect(app.Logger()).ToNot(BeNil())
				})
			})
		})
	})
})
