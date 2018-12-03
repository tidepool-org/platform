package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/application"
	applicationTest "github.com/tidepool-org/platform/application/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
)

var _ = Describe("Runner", func() {
	Context("RunAndExit", func() {
		// NOTE: Cannot be tested due to embedded os.Exit
	})

	Context("Run", func() {
		var runner *applicationTest.Runner
		var provider *applicationTest.Provider

		BeforeEach(func() {
			runner = applicationTest.NewRunner()
			provider = applicationTest.NewProvider()
		})

		AfterEach(func() {
			provider.AssertOutputsEmpty()
			runner.AssertOutputsEmpty()
		})

		It("returns error when runner is missing", func() {
			Expect(application.Run(nil, provider)).To(MatchError("runner is missing"))
		})

		It("returns error when provider is missing", func() {
			Expect(application.Run(runner, nil)).To(MatchError("provider is missing"))
		})

		When("Initialize is invoked", func() {
			AfterEach(func() {
				Expect(runner.InitializeInputs).To(Equal([]application.Provider{provider}))
			})

			It("returns error when Initialize returns error", func() {
				err := errorsTest.RandomError()
				runner.InitializeOutputs = []error{err}
				Expect(application.Run(runner, provider)).To(MatchError(fmt.Sprintf("unable to initialize runner; %s", err)))
			})

			When("Initialize returns successfully", func() {
				BeforeEach(func() {
					runner.InitializeOutputs = []error{nil}
				})

				It("returns error when Run returns error", func() {
					err := errorsTest.RandomError()
					runner.RunOutputs = []error{err}
					Expect(application.Run(runner, provider)).To(MatchError(fmt.Sprintf("unable to run runner; %s", err)))
				})

				When("Run returns successfully", func() {
					BeforeEach(func() {
						runner.RunOutputs = []error{nil}
					})

					It("returns successfully", func() {
						Expect(application.Run(runner, provider)).To(Succeed())
					})
				})
			})
		})
	})
})
