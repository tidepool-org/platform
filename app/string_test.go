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

	DescribeTable("StringsContainsString",
		func(sourceStrings []string, searchString string, expectedResult bool) {
			Expect(app.StringsContainsString(sourceStrings, searchString)).To(Equal(expectedResult))
		},
		Entry("is nil source strings with empty search string", nil, "", false),
		Entry("is nil source strings with valid search string", nil, "one", false),
		Entry("is empty source strings with empty search string", []string{}, "", false),
		Entry("is empty source strings with valid search string", []string{}, "one", false),
		Entry("is single source strings with empty search string", []string{"two"}, "", false),
		Entry("is single source strings with non-matching search string", []string{"two"}, "one", false),
		Entry("is single source strings with matching search string", []string{"two"}, "two", true),
		Entry("is multiple source strings with empty search string", []string{"zero", "two", "four"}, "", false),
		Entry("is multiple source strings with non-matching search string", []string{"zero", "two", "four"}, "one", false),
		Entry("is multiple source strings with matching search string", []string{"zero", "two", "four"}, "two", true),
	)

	DescribeTable("StringsContainsAnyStrings",
		func(sourceStrings []string, searchStrings []string, expectedResult bool) {
			Expect(app.StringsContainsAnyStrings(sourceStrings, searchStrings)).To(Equal(expectedResult))
		},
		Entry("is nil source strings with nil search strings", nil, nil, false),
		Entry("is nil source strings with empty search strings", nil, []string{}, false),
		Entry("is nil source strings with single invalid search strings", nil, []string{"one"}, false),
		Entry("is nil source strings with multiple invalid search strings", nil, []string{"one", "three"}, false),
		Entry("is empty source strings with nil search strings", []string{}, nil, false),
		Entry("is empty source strings with empty search strings", []string{}, []string{}, false),
		Entry("is empty source strings with single invalid search strings", []string{}, []string{"one"}, false),
		Entry("is empty source strings with multiple invalid search strings", []string{}, []string{"one", "three"}, false),
		Entry("is single source strings with nil search strings", []string{"two"}, nil, false),
		Entry("is single source strings with single search strings", []string{"two"}, []string{}, false),
		Entry("is single source strings with single invalid search strings", []string{"two"}, []string{"one"}, false),
		Entry("is single source strings with single valid search strings", []string{"two"}, []string{"two"}, true),
		Entry("is single source strings with multiple invalid search strings", []string{"two"}, []string{"one", "three"}, false),
		Entry("is single source strings with multiple invalid and valid search strings", []string{"two"}, []string{"one", "two", "three", "four"}, true),
		Entry("is multiple source strings with nil search strings", []string{"two", "four"}, nil, false),
		Entry("is multiple source strings with single search strings", []string{"two", "four"}, []string{}, false),
		Entry("is multiple source strings with single invalid search strings", []string{"two", "four"}, []string{"one"}, false),
		Entry("is multiple source strings with single valid search strings", []string{"two", "four"}, []string{"two"}, true),
		Entry("is multiple source strings with multiple invalid search strings", []string{"two", "four"}, []string{"one", "three"}, false),
		Entry("is multiple source strings with multiple valid search strings", []string{"two", "four"}, []string{"two", "four"}, true),
		Entry("is multiple source strings with multiple invalid and valid search strings", []string{"two", "four"}, []string{"one", "two", "three", "four"}, true),
	)
})
