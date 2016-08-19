package environment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/tidepool-org/platform/environment"
)

var _ = Describe("Default", func() {
	Context("NewDefaultReporter", func() {
		It("returns successfully", func() {
			os.Setenv("ENV", "test")
			Expect(environment.NewDefaultReporter()).ToNot(BeNil())
		})
	})
})
