package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/application/version"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"

	_ "github.com/tidepool-org/platform/application/version/test"
)

var _ = Describe("Application", func() {
	Context("New", func() {
		It("returns an error if the prefix is missing", func() {
			app, err := application.New("")
			Expect(err).To(MatchError("prefix is missing"))
			Expect(app).To(BeNil())
		})

		It("returns successfully with no scopes", func() {
			Expect(application.New("TIDEPOOL")).ToNot(BeNil())
		})

		It("returns successfully with one scope", func() {
			Expect(application.New("TIDEPOOL", "alpha")).ToNot(BeNil())
		})

		It("returns successfully with multiple scopes", func() {
			Expect(application.New("TIDEPOOL", "alpha", "bravo")).ToNot(BeNil())
		})
	})

	Context("with new application", func() {
		var app *application.Application

		BeforeEach(func() {
			var err error
			app, err = application.New("TIDEPOOL", "alpha", "bravo")
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
					Expect(app.Initialize()).To(MatchError("unable to create version reporter; base is missing"))
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
					configReporter = configReporter.WithScopes("application.test", "alpha", "bravo", "logger")
					level = configReporter.GetWithDefault("level", "warn")
					configReporter.Set("level", "invalid")
				})

				AfterEach(func() {
					configReporter.Set("level", level)
				})

				It("returns an error if the logger level is invalid", func() {
					Expect(app.Initialize()).To(MatchError("unable to create logger; level not found"))
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
				Expect(app.Name()).To(Equal("application.test"))
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
