package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("VersionReporter", func() {
	Context("NewVersionReporter", func() {
		var base string
		var commit string

		It("returns successfully", func() {
			base = netTest.RandomSemanticVersion()
			commit = test.RandomStringFromRangeAndCharset(40, 40, test.CharsetHexidecimalLowercase)
			application.VersionBase = base
			application.VersionShortCommit = commit[:8]
			application.VersionFullCommit = commit
			versionReporter, err := application.NewVersionReporter()
			Expect(err).ToNot(HaveOccurred())
			Expect(versionReporter).ToNot(BeNil())
			Expect(versionReporter.Base()).To(Equal(base))
			Expect(versionReporter.ShortCommit()).To(Equal(commit[:8]))
			Expect(versionReporter.FullCommit()).To(Equal(commit))
		})
	})
})
