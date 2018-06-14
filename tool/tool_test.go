package tool_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	applicationTest "github.com/tidepool-org/platform/application/test"
	"github.com/tidepool-org/platform/tool"
)

var _ = Describe("Tool", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(tool.New()).ToNot(BeNil())
		})
	})

	Context("with new tool", func() {
		var provider *applicationTest.Provider
		var tuel *tool.Tool

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()
			tuel = tool.New()
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

				Context("CLI", func() {
					It("returns not nil", func() {
						Expect(tuel.CLI()).ToNot(BeNil())
					})
				})

				Context("Args", func() {
					It("returns nil", func() {
						Expect(tuel.Args()).To(BeNil())
					})
				})
			})
		})
	})
})
