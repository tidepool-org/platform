package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("User", func() {
	Context("NewUserID", func() {
		It("returns a string of 10 lowercase hexidecimal characters", func() {
			Expect(auth.NewUserID()).To(MatchRegexp("^[0-9a-f]{10}$"))
		})

		It("returns different UserIDs for each invocation", func() {
			Expect(auth.NewUserID()).ToNot(Equal(auth.NewUserID()))
		})
	})

	Context("IsValidUserID, UserIDValidator, and ValidateUserID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(auth.IsValidUserID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				auth.UserIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(auth.ValidateUserID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "01234abcd", auth.ErrorValueStringAsUserIDNotValid("01234abcd")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)),
			Entry("has string length out of range (upper)", "01234abcdef01234abcdef01234abcdef", auth.ErrorValueStringAsUserIDNotValid("01234abcdef01234abcdef01234abcdef")),
			Entry("has uppercase characters", "01234ABCDE", auth.ErrorValueStringAsUserIDNotValid("01234ABCDE")),
			Entry("has symbols", "0123$%BCDE", auth.ErrorValueStringAsUserIDNotValid("0123$%BCDE")),
			Entry("has whitespace", "0123  BCDE", auth.ErrorValueStringAsUserIDNotValid("0123  BCDE")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsUserIDNotValid with empty string", auth.ErrorValueStringAsUserIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as user id`),
			Entry("is ErrorValueStringAsUserIDNotValid with non-empty string", auth.ErrorValueStringAsUserIDNotValid("01234abcde"), "value-not-valid", "value is not valid", `value "01234abcde" is not valid as user id`),
		)
	})
})
