package environment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/environment"
)

var _ = Describe("Default", func() {
	Context("NewDefaultReporter", func() {
		It("returns successfully with prefix", func() {
			prefix := app.NewID()
			name := app.NewID()
			os.Setenv(fmt.Sprintf("%s_ENV", prefix), name)
			environment, err := environment.NewDefaultReporter(prefix)
			Expect(err).ToNot(HaveOccurred())
			Expect(environment).ToNot(BeNil())
			Expect(environment.Name()).To(Equal(name))
			Expect(environment.Prefix()).To(Equal(prefix))
		})

		It("returns successfully without a prefix", func() {
			name := app.NewID()
			os.Setenv("ENV", name)
			environment, err := environment.NewDefaultReporter("")
			Expect(err).ToNot(HaveOccurred())
			Expect(environment).ToNot(BeNil())
			Expect(environment.Name()).To(Equal(name))
			Expect(environment.Prefix()).To(Equal(""))
		})

		It("returns an error if ENV with prefix not defined", func() {
			environment, err := environment.NewDefaultReporter(app.NewID())
			Expect(err).To(MatchError("environment: name is missing"))
			Expect(environment).To(BeNil())
		})
	})
})
