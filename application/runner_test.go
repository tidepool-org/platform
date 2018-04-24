package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/tidepool-org/platform/application"
	testApplication "github.com/tidepool-org/platform/application/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
)

var _ = Describe("Runner", func() {
	Context("Run", func() {
		var rnnr *testApplication.Runner
		var oldStderr *os.File

		BeforeEach(func() {
			rnnr = testApplication.NewRunner()
			oldStderr = os.Stderr
			newStderr, err := os.Open(os.DevNull)
			Expect(err).ToNot(HaveOccurred())
			Expect(newStderr).ToNot(BeNil())
			os.Stderr = newStderr
		})

		AfterEach(func() {
			if oldStderr != nil {
				os.Stderr = oldStderr
			}
			rnnr.Expectations()
		})

		It("returns failure if err is not nil", func() {
			Expect(application.Run(rnnr, testErrors.NewError())).To(Equal(application.Failure))
		})

		It("returns failure if runner is nil", func() {
			Expect(application.Run(nil, nil)).To(Equal(application.Failure))
		})

		It("returns failure if Initialize returns error", func() {
			rnnr.InitializeOutputs = []error{testErrors.NewError()}
			Expect(application.Run(rnnr, nil)).To(Equal(application.Failure))
		})

		Context("with successful Initialize", func() {
			BeforeEach(func() {
				rnnr.InitializeOutputs = []error{nil}
			})

			It("returns failure if Run returns error", func() {
				rnnr.RunOutputs = []error{testErrors.NewError()}
				Expect(application.Run(rnnr, nil)).To(Equal(application.Failure))
			})

			Context("with successful Run", func() {
				BeforeEach(func() {
					rnnr.RunOutputs = []error{nil}
				})

				It("returns successfully", func() {
					Expect(application.Run(rnnr, nil)).To(Equal(application.Success))
				})
			})
		})
	})
})
