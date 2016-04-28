package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Version", func() {
	It("uses default values", func() {
		Expect(version.Base()).To(Equal("0.0.0"))
		Expect(version.Commit()).To(Equal("0000000000000000000000000000000000000000"))
		Expect(version.ShortCommit()).To(Equal("00000000"))
		Expect(version.Short()).To(Equal("0.0.0+00000000"))
		Expect(version.Long()).To(Equal("0.0.0+0000000000000000000000000000000000000000"))
	})

	DescribeTable("properly constructs the Short and Long alternatives",
		func(base string, commit string, short string, long string) {
			version.BaseInitial = base
			version.CommitInitial = commit
			Expect(version.Short()).To(Equal(short))
			Expect(version.Long()).To(Equal(long))
		},
		Entry("returns the major base with default commit", "1", "", "1+00000000", "1+0000000000000000000000000000000000000000"),
		Entry("returns the major.minor base with default commit", "1.2", "", "1.2+00000000", "1.2+0000000000000000000000000000000000000000"),
		Entry("returns the major.minor.patch base with default commit", "1.2.3", "", "1.2.3+00000000", "1.2.3+0000000000000000000000000000000000000000"),
		Entry("returns the major.minor.patch base with default commit", "1.2.3", "", "1.2.3+00000000", "1.2.3+0000000000000000000000000000000000000000"),
		Entry("returns the default base with 1-character commit", "", "1", "0.0.0+1", "0.0.0+1"),
		Entry("returns the default base with 8-character commit", "", "12345678", "0.0.0+12345678", "0.0.0+12345678"),
		Entry("returns the default base with 9-character commit", "", "123456789", "0.0.0+12345678", "0.0.0+123456789"),
		Entry("returns the default base with full commit", "", "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd", "0.0.0+12345678", "0.0.0+1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd"),
	)
})
