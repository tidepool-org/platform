package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	applicationTest "github.com/tidepool-org/platform/application/test"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
)

var _ = Describe("Migration", func() {
	Context("NewMigration", func() {
		It("returns successfully", func() {
			Expect(migrationMongo.NewMigration()).ToNot(BeNil())
		})
	})

	Context("with new migration", func() {
		var provider *applicationTest.Provider
		var migration *migrationMongo.Migration

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()
			migration = migrationMongo.NewMigration()
			Expect(migration).ToNot(BeNil())
		})

		AfterEach(func() {
			provider.AssertOutputsEmpty()
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				migration.Terminate()
			})

			Context("Initialize", func() {
				It("returns error when provider is missing", func() {
					Expect(migration.Initialize(nil)).To(MatchError("provider is missing"))
				})

				It("returns successfully", func() {
					Expect(migration.Initialize(provider)).To(Succeed())
				})
			})

			Context("with Initialize before", func() {
				BeforeEach(func() {
					Expect(migration.Initialize(provider)).To(Succeed())
				})

				Context("Terminate", func() {
					It("returns successfully", func() {
						migration.Terminate()
					})
				})

				Context("DryRun", func() {
					It("returns false", func() {
						Expect(migration.DryRun()).To(BeFalse())
					})
				})
			})
		})
	})
})
