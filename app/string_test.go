package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("String", func() {
	Context("FirstStringNotEmpty", func() {
		It("returns an empty string with no arguments", func() {
			Expect(app.FirstStringNotEmpty()).To(Equal(""))
		})

		It("returns an empty string with all empty arguments", func() {
			Expect(app.FirstStringNotEmpty("", "", "")).To(Equal(""))
		})

		It("returns the first empty string with only one argument", func() {
			Expect(app.FirstStringNotEmpty("pixie")).To(Equal("pixie"))
		})

		It("returns the first empty string with multiple arguments", func() {
			Expect(app.FirstStringNotEmpty("", "", "goblin", "")).To(Equal("goblin"))
		})
	})

	Context("SplitStringAndRemoveWhitespace", func() {
		DescribeTable("returns expected value when",
			func(sourceString string, stringSeperator string, expectedStringArray []string) {
				Expect(app.SplitStringAndRemoveWhitespace(sourceString, stringSeperator)).To(Equal(expectedStringArray))
			},
			Entry("empty source string with no comma separator", "", ",", []string{}),
			Entry("whitespace-only source string with no comma separator", "   ", ",", []string{}),
			Entry("source string with only comma separators", ",,,", ",", []string{}),
			Entry("whitespace-only source string with comma separators", "  ,,   ,, ", ",", []string{}),
			Entry("non-whitespace source string with no comma separator", "alpha", ",", []string{"alpha"}),
			Entry("source string with whitespace no comma separator", "  alpha  ", ",", []string{"alpha"}),
			Entry("source strings with comma separators", "alpha,beta,charlie", ",", []string{"alpha", "beta", "charlie"}),
			Entry("source strings with whitespace and comma separators", "  alpha   ,  beta, charlie    ", ",", []string{"alpha", "beta", "charlie"}),
			Entry("source string with whitespace and whitespace separator", "  alpha    beta   charlie", " ", []string{"alpha", "beta", "charlie"}),
			Entry("source string with whitespace and empty separator", "  alpha    beta   charlie", "", []string{"a", "l", "p", "h", "a", "b", "e", "t", "a", "c", "h", "a", "r", "l", "i", "e"}),
		)
	})
})
