package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application/version"
	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/tool/mongo"
)

var _ = Describe("Tool", func() {
	Context("New", func() {
		It("returns an error if the prefix is missing", func() {
			tuel, err := mongo.NewTool("")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(tuel).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(mongo.NewTool("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with new tool", func() {
		var tuel *mongo.Tool

		BeforeEach(func() {
			var err error
			tuel, err = mongo.NewTool("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(tuel).ToNot(BeNil())
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
					Expect(tuel.Initialize()).To(MatchError("application: unable to create version reporter; version: base is missing"))
				})
			})

			Context("with invalid store tls", func() {
				var configReporter config.Reporter
				var tls string

				BeforeEach(func() {
					var err error
					configReporter, err = env.NewReporter("TIDEPOOL")
					Expect(err).ToNot(HaveOccurred())
					Expect(configReporter).ToNot(BeNil())
					configReporter = configReporter.WithScopes("mongo.test", "store")
					tls = configReporter.GetWithDefault("tls", "false")
					configReporter.Set("tls", "invalid")
				})

				AfterEach(func() {
					configReporter.Set("tls", tls)
				})

				It("returns an error if the store tls is invalid", func() {
					Expect(tuel.Initialize()).To(MatchError("mongo: unable to load store config; mongo: tls is invalid"))
				})
			})

			It("returns successfully", func() {
				Expect(tuel.Initialize()).To(Succeed())
			})
		})

		Context("Terminate", func() {
			It("returns without panic", func() {
				tuel.Terminate()
			})
		})

		Context("initialized", func() {
			BeforeEach(func() {
				Expect(tuel.Initialize()).To(Succeed())
			})

			AfterEach(func() {
				tuel.Terminate()
			})

			Context("NewMongoConfig", func() {
				It("returns not nil", func() {
					Expect(tuel.NewMongoConfig()).ToNot(BeNil())
				})

				It("returns a new config each time", func() {
					Expect(tuel.NewMongoConfig()).ToNot(BeIdenticalTo(tuel.NewMongoConfig()))
				})
			})
		})
	})
})
