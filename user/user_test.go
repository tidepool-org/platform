package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

var _ = Describe("User", func() {
	Context("ID", func() {
		Context("NewID", func() {
			It("returns a string of 10 lowercase hexidecimal characters", func() {
				Expect(user.NewID()).To(MatchRegexp("^[0-9a-f]{10}$"))
			})

			It("returns different IDs for each invocation", func() {
				Expect(user.NewID()).ToNot(Equal(user.NewID()))
			})
		})

		Context("IsValidID, IDValidator, and ValidateID", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(user.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					user.IDValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(user.ValidateID(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("has string length out of range (lower)", "01234abcd", user.ErrorValueStringAsIDNotValid("01234abcd")),
				Entry("has string length in range", test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)),
				Entry("has string length out of range (upper)", "01234abcdef01234abcdef01234abcdef", user.ErrorValueStringAsIDNotValid("01234abcdef01234abcdef01234abcdef")),
				Entry("has uppercase characters", "01234ABCDE", user.ErrorValueStringAsIDNotValid("01234ABCDE")),
				Entry("has symbols", "012$%^&cde", user.ErrorValueStringAsIDNotValid("012$%^&cde")),
				Entry("has whitespace", "012    cde", user.ErrorValueStringAsIDNotValid("012    cde")),
			)
		})
	})
})
