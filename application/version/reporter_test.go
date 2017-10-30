package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application/version"
)

var _ = Describe("Reporter", func() {
	Context("NewReporter", func() {
		It("returns successfully", func() {
			version.Base = "1.2.3"
			version.ShortCommit = "4567890"
			version.FullCommit = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn"
			Expect(version.NewReporter()).ToNot(BeNil())
		})
	})
})
