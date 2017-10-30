package version_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Reporter", func() {
	Context("NewReporter", func() {
		It("returns an error if base is missing", func() {
			reporter, err := version.NewReporter("", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(MatchError("base is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if short commit is missing", func() {
			reporter, err := version.NewReporter("1.2.3", "", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).To(MatchError("short commit is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if full commit is missing", func() {
			reporter, err := version.NewReporter("1.2.3", "4567890", "")
			Expect(err).To(MatchError("full commit is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")).ToNot(BeNil())
		})
	})

	Context("Reporter", func() {
		var reporter version.Reporter

		BeforeEach(func() {
			var err error
			reporter, err = version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
			Expect(err).ToNot(HaveOccurred())
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

		It("returns the expected short", func() {
			Expect(reporter.Short()).To(Equal("1.2.3+4567890"))
		})

		It("returns the expected long", func() {
			Expect(reporter.Long()).To(Equal("1.2.3+ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn"))
		})
	})
})
