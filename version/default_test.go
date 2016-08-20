package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Default", func() {
	Context("NewDefaultReporter", func() {
		It("returns successfully", func() {
			version.Base = "1.2.3"
			version.ShortCommit = "4567890"
			version.FullCommit = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn"
			Expect(version.NewDefaultReporter()).ToNot(BeNil())
		})
	})
})
