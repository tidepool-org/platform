package auth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("ProviderSession", func() {
	Context("NewProviderSessionID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(auth.NewProviderSessionID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(auth.NewProviderSessionID()).ToNot(Equal(auth.NewProviderSessionID()))
		})
	})

	Context("IsValidProviderSessionID, ProviderSessionIDValidator, and ValidateProviderSessionID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(auth.IsValidProviderSessionID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				auth.ProviderSessionIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(auth.ValidateProviderSessionID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdef0123456789abcde", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("has string length out of range (upper)", "0123456789abcdef0123456789abcdef0", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has uppercase characters", "0123456789ABCDEF0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789ABCDEF0123456789abcdef")),
			Entry("has symbols", "0123456789$%^&*(0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789$%^&*(0123456789abcdef")),
			Entry("has whitespace", "0123456789      0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789      0123456789abcdef")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsProviderSessionIDNotValid with empty string", auth.ErrorValueStringAsProviderSessionIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as provider session id`),
			Entry("is ErrorValueStringAsProviderSessionIDNotValid with non-empty string", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcdef"), "value-not-valid", "value is not valid", `value "0123456789abcdef0123456789abcdef" is not valid as provider session id`),
		)
	})
})
