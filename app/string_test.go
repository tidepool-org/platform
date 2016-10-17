package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"errors"

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
			Entry("has empty source string with no comma separator", "", ",", []string{}),
			Entry("has whitespace-only source string with no comma separator", "   ", ",", []string{}),
			Entry("has source string with only comma separators", ",,,", ",", []string{}),
			Entry("has whitespace-only source string with comma separators", "  ,,   ,, ", ",", []string{}),
			Entry("has non-whitespace source string with no comma separator", "alpha", ",", []string{"alpha"}),
			Entry("has source string with whitespace no comma separator", "  alpha  ", ",", []string{"alpha"}),
			Entry("has source strings with comma separators", "alpha,beta,charlie", ",", []string{"alpha", "beta", "charlie"}),
			Entry("has source strings with whitespace and comma separators", "  alpha   ,  beta, charlie    ", ",", []string{"alpha", "beta", "charlie"}),
			Entry("has source string with whitespace and whitespace separator", "  alpha    beta   charlie", " ", []string{"alpha", "beta", "charlie"}),
			Entry("has source string with whitespace and empty separator", "  alpha    beta   charlie", "", []string{"a", "l", "p", "h", "a", "b", "e", "t", "a", "c", "h", "a", "r", "l", "i", "e"}),
		)
	})

	Context("QuoteIfString", func() {
		It("returns nil when the interface value is nil", func() {
			Expect(app.QuoteIfString(nil)).To(BeNil())
		})

		DescribeTable("returns expected value when",
			func(interfaceValue interface{}, expectedValue interface{}) {
				Expect(app.QuoteIfString(interfaceValue)).To(Equal(expectedValue))
			},
			Entry("is a string", "a string", `"a string"`),
			Entry("is an empty string", "", `""`),
			Entry("is an error", errors.New("error"), errors.New("error")),
			Entry("is an integer", 1, 1),
			Entry("is a float", 1.23, 1.23),
			Entry("is an array", []string{"a"}, []string{"a"}),
			Entry("is a map", map[string]string{"a": "b"}, map[string]string{"a": "b"}),
		)
	})

	DescribeTable("StringArrayContains",
		func(sourceStrings []string, searchString string, expectedResult bool) {
			Expect(app.StringArrayContains(sourceStrings, searchString)).To(Equal(expectedResult))
		},
		Entry("is an empty source array with empty search string", []string{}, "", false),
		Entry("is an empty source array with valid search string", []string{}, "one", false),
		Entry("is an single source array with empty search string", []string{"two"}, "", false),
		Entry("is an single source array with non-matching search string", []string{"two"}, "one", false),
		Entry("is an single source array with matching search string", []string{"two"}, "two", true),
		Entry("is an multiple source array with empty search string", []string{"zero", "two", "four"}, "", false),
		Entry("is an multiple source array with non-matching search string", []string{"zero", "two", "four"}, "one", false),
		Entry("is an multiple source array with matching search string", []string{"zero", "two", "four"}, "two", true),
	)
})
