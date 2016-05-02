package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("BaseCommit", func() {

	It("has default values", func() {
		baseCommit := version.NewBaseCommit("", "")
		Expect(baseCommit.Base()).To(Equal("0.0.0"))
		Expect(baseCommit.Commit()).To(Equal("0000000000000000000000000000000000000000"))
		Expect(baseCommit.ShortCommit()).To(Equal("00000000"))
		Expect(baseCommit.Short()).To(Equal("0.0.0+00000000"))
		Expect(baseCommit.Long()).To(Equal("0.0.0+0000000000000000000000000000000000000000"))
	})

	DescribeTable("properly constructs the Short and Long alternatives",
		func(base string, commit string, short string, long string) {
			baseCommit := version.NewBaseCommit(base, commit)
			Expect(baseCommit.Short()).To(Equal(short))
			Expect(baseCommit.Long()).To(Equal(long))
		},
		Entry("returns the default base with default commit", "", "", "0.0.0+00000000", "0.0.0+0000000000000000000000000000000000000000"),
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
