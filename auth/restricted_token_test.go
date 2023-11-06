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

var _ = Describe("RestrictedToken", func() {
	Context("NewRestrictedTokenID", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(auth.NewRestrictedTokenID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(auth.NewRestrictedTokenID()).ToNot(Equal(auth.NewRestrictedTokenID()))
		})
	})

	Context("IsValidRestrictedTokenID, RestrictedTokenIDValidator, and ValidateRestrictedTokenID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(auth.IsValidRestrictedTokenID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				auth.RestrictedTokenIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(auth.ValidateRestrictedTokenID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdef0123456789abcde", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("has string length out of range (upper)", "0123456789abcdef0123456789abcdef0", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has uppercase characters", "0123456789ABCDEF0123456789abcdef", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789ABCDEF0123456789abcdef")),
			Entry("has symbols", "0123456789$%^&*(0123456789abcdef", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789$%^&*(0123456789abcdef")),
			Entry("has whitespace", "0123456789      0123456789abcdef", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789      0123456789abcdef")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsRestrictedTokenIDNotValid with empty string", auth.ErrorValueStringAsRestrictedTokenIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as restricted token id`),
			Entry("is ErrorValueStringAsRestrictedTokenIDNotValid with non-empty string", auth.ErrorValueStringAsRestrictedTokenIDNotValid("0123456789abcdef0123456789abcdef"), "value-not-valid", "value is not valid", `value "0123456789abcdef0123456789abcdef" is not valid as restricted token id`),
		)
	})
})
