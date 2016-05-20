package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Version", func() {
	Context("NewReporter", func() {
		It("returns an error if base is missing", func() {
			reporter, err := version.NewReporter("", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(MatchError("version: base is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if shortCommit is missing", func() {
			reporter, err := version.NewReporter("1.2.3", "", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(MatchError("version: shortCommit is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if fullCommit is missing", func() {
			reporter, err := version.NewReporter("1.2.3", "4567890", "")
			Expect(err).To(MatchError("version: fullCommit is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns successfully", func() {
			reporter, err := version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(Succeed())
			Expect(reporter).ToNot(BeNil())
		})
	})

	Context("Reporter", func() {
		var reporter version.Reporter

		BeforeEach(func() {
			var err error
			reporter, err = version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(Succeed())
			Expect(reporter).ToNot(BeNil())
		})

		It("returns the expected base", func() {
			Expect(reporter.Base()).To(Equal("1.2.3"))
		})

		It("returns the expected short commit", func() {
			Expect(reporter.ShortCommit()).To(Equal("4567890"))
		})

		It("returns the expected full commit", func() {
			Expect(reporter.FullCommit()).To(Equal("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn"))
		})

		It("returns the expected short form", func() {
			Expect(reporter.Short()).To(Equal("1.2.3+4567890"))
		})

		It("returns the expected long form", func() {
			Expect(reporter.Long()).To(Equal("1.2.3+ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn"))
		})
	})
})
