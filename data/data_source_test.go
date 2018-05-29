package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DataSource", func() {
	Context("NewSourceID", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(data.NewSourceID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(data.NewSourceID()).ToNot(Equal(data.NewSourceID()))
		})
	})

	Context("IsValidSourceID, SourceIDValidator, and ValidateSourceID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(data.IsValidSourceID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				data.SourceIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(data.ValidateSourceID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is first version with string length out of range (lower)", "upid_0123456789a", data.ErrorValueStringAsSourceIDNotValid("upid_0123456789a")),
			Entry("is first version with string length in range", "upid_"+test.RandomStringFromRangeAndCharset(12, 12, test.CharsetHexidecimalLowercase)),
			Entry("is first version with uppercase characters", "upid_0123456789AB", data.ErrorValueStringAsSourceIDNotValid("upid_0123456789AB")),
			Entry("is second version with string length in range", "upid_"+test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("is second version with uppercase characters", "upid_0123456789ABCDEF0123456789ABCDEF", data.ErrorValueStringAsSourceIDNotValid("upid_0123456789ABCDEF0123456789ABCDEF")),
			Entry("is second version with string length out of range (upper)", "upid_0123456789abcdef0123456789abcdef0", data.ErrorValueStringAsSourceIDNotValid("upid_0123456789abcdef0123456789abcdef0")),
			Entry("is third version with string length out of range (lower)", "0123456789abcdef0123456789abcde", data.ErrorValueStringAsSourceIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("is third version with string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("is third version with uppercase characters", "0123456789ABCDEF0123456789ABCDEF", data.ErrorValueStringAsSourceIDNotValid("0123456789ABCDEF0123456789ABCDEF")),
			Entry("is third version with string length out of range (upper)", "0123456789abcdef0123456789abcdef0", data.ErrorValueStringAsSourceIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has invalid prefix", "UPID_0123456789abcdef0123456789abcdef", data.ErrorValueStringAsSourceIDNotValid("UPID_0123456789abcdef0123456789abcdef")),
			Entry("has symbols", "0123456789!@#$%^0123456789!@#$%^", data.ErrorValueStringAsSourceIDNotValid("0123456789!@#$%^0123456789!@#$%^")),
			Entry("has whitespace", "0123456789      0123456789      ", data.ErrorValueStringAsSourceIDNotValid("0123456789      0123456789      ")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsSourceIDNotValid with empty string", data.ErrorValueStringAsSourceIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as data source id`),
			Entry("is ErrorValueStringAsSourceIDNotValid with non-empty string", data.ErrorValueStringAsSourceIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as data source id`),
		)
	})
})
