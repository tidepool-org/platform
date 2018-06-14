package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	applicationTest "github.com/tidepool-org/platform/application/test"
	toolMongo "github.com/tidepool-org/platform/tool/mongo"
)

var _ = Describe("Tool", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(toolMongo.NewTool()).ToNot(BeNil())
		})
	})

	Context("with new tool", func() {
		var provider *applicationTest.Provider
		var tuel *toolMongo.Tool

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()
			tuel = toolMongo.NewTool()
			Expect(tuel).ToNot(BeNil())
		})

		AfterEach(func() {
			provider.AssertOutputsEmpty()
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				tuel.Terminate()
			})

			Context("Initialize", func() {
				It("returns an error when the provider is missing", func() {
					Expect(tuel.Initialize(nil)).To(MatchError("provider is missing"))
				})

				It("returns successfully", func() {
					Expect(tuel.Initialize(provider)).To(Succeed())
				})
			})

			Context("Terminate", func() {
				It("returns successfully", func() {
					tuel.Terminate()
				})
			})

			Context("with Initialize before", func() {
				BeforeEach(func() {
					Expect(tuel.Initialize(provider)).To(Succeed())
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
})
