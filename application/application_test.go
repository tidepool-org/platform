package application_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application"
	applicationTest "github.com/tidepool-org/platform/application/test"
)

var _ = Describe("Application", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(application.New()).ToNot(BeNil())
		})
	})

	Context("with new application", func() {
		var provider *applicationTest.Provider
		var app *application.Application

		BeforeEach(func() {
			provider = applicationTest.NewProvider()
			app = application.New()
			Expect(app).ToNot(BeNil())
		})

		AfterEach(func() {
			provider.AssertOutputsEmpty()
		})

		Context("Initialize", func() {
			It("returns an error when the provider is missing", func() {
				Expect(app.Initialize(nil)).To(MatchError("provider is missing"))
			})

			It("returns successfully", func() {
				Expect(app.Initialize(provider)).To(Succeed())
			})
		})

		Context("Terminate", func() {
			It("returns successfully", func() {
				app.Terminate()
			})
		})
	})
})
