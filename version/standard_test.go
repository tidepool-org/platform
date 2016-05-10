package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Standard", func() {

	It("has default values", func() {
		standard := version.NewStandard("", "", "")
		Expect(standard.Base()).To(Equal("0.0.0"))
		Expect(standard.ShortCommit()).To(Equal("0000000"))
		Expect(standard.FullCommit()).To(Equal("0000000000000000000000000000000000000000"))
		Expect(standard.Short()).To(Equal("0.0.0+0000000"))
		Expect(standard.Long()).To(Equal("0.0.0+0000000000000000000000000000000000000000"))
	})

	DescribeTable("properly constructs the Short and Long alternatives",
		func(base string, shortCommit string, fullCommit string, short string, long string) {
			standard := version.NewStandard(base, shortCommit, fullCommit)
			Expect(standard.Short()).To(Equal(short))
			Expect(standard.Long()).To(Equal(long))
		},
		Entry("returns the default base with default short commit and default full commit", "", "", "", "0.0.0+0000000", "0.0.0+0000000000000000000000000000000000000000"),
		Entry("returns the major base with default short commit and default full commit", "1", "", "", "1+0000000", "1+0000000000000000000000000000000000000000"),
		Entry("returns the major.minor base with default short commit and default full commit", "1.2", "", "", "1.2+0000000", "1.2+0000000000000000000000000000000000000000"),
		Entry("returns the major.minor.patch base with default short commit and default full commit", "1.2.3", "", "", "1.2.3+0000000", "1.2.3+0000000000000000000000000000000000000000"),
		Entry("returns the default base with 1-character short commit and default full commit", "", "1", "", "0.0.0+1", "0.0.0+0000000000000000000000000000000000000000"),
		Entry("returns the default base with 7-character short commit and default full commit", "", "1234567", "", "0.0.0+1234567", "0.0.0+0000000000000000000000000000000000000000"),
		Entry("returns the default base with 8-character short commit and default full commit", "", "12345678", "", "0.0.0+12345678", "0.0.0+0000000000000000000000000000000000000000"),
		Entry("returns the default base and short commit with 1-character full commit", "", "", "1", "0.0.0+0000000", "0.0.0+1"),
		Entry("returns the default base and short commit with 7-character full commit", "", "", "1234567", "0.0.0+0000000", "0.0.0+1234567"),
		Entry("returns the default base and short commit with 40-character full commit", "", "", "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd", "0.0.0+0000000", "0.0.0+1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd"),
		Entry("returns the major.minor.patch base and short commit with 40-character fullcommit", "1.2.3", "1234567", "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd", "1.2.3+1234567", "1.2.3+1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd"),
	)
})
