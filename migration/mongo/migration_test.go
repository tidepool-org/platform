package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application/version"
	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/migration/mongo"
)

var _ = Describe("Migration", func() {
	Context("New", func() {
		It("returns an error if the prefix is missing", func() {
			migration, err := mongo.NewMigration("")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(migration).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(mongo.NewMigration("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with new migration", func() {
		var migration *mongo.Migration

		BeforeEach(func() {
			var err error
			migration, err = mongo.NewMigration("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(migration).ToNot(BeNil())
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
					Expect(migration.Initialize()).To(MatchError("application: unable to create version reporter; version: base is missing"))
				})
			})

			It("returns successfully", func() {
				Expect(migration.Initialize()).To(Succeed())
			})
		})

		Context("Terminate", func() {
			It("returns without panic", func() {
				migration.Terminate()
			})
		})

		Context("initialized", func() {
			BeforeEach(func() {
				Expect(migration.Initialize()).To(Succeed())
			})

			AfterEach(func() {
				migration.Terminate()
			})

			Context("DryRun", func() {
				It("returns false", func() {
					Expect(migration.DryRun()).To(BeFalse())
				})
			})
		})
	})
})
