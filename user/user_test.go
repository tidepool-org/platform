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
	DescribeTable("HasRole",
		func(roles []string, role string, expectedResult bool) {
			testUser := &user.User{
				Roles: roles,
			}
			Expect(testUser.HasRole(role)).To(Equal(expectedResult))
		},
		Entry("roles is nil, role is empty", nil, "", false),
		Entry("roles is nil, role is specified", nil, user.ClinicRole, false),
		Entry("roles is empty, role is empty", []string{}, "", false),
		Entry("roles is empty, role is specified", []string{}, user.ClinicRole, false),
		Entry("roles has one, role is empty", []string{user.ClinicRole}, "", false),
		Entry("roles has one, role is specified, not in roles", []string{user.ClinicRole}, "unknown", false),
		Entry("roles has one, role is specified, in roles", []string{user.ClinicRole}, user.ClinicRole, true),
		Entry("roles has many, role is empty", []string{"administrator", user.ClinicRole, "manager"}, "", false),
		Entry("roles has many, role is specified, not in roles", []string{"administrator", user.ClinicRole, "manager"}, "unknown", false),
		Entry("roles has many, role is specified, in roles", []string{"administrator", user.ClinicRole, "manager"}, user.ClinicRole, true),
	)

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
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "01234abcd", user.ErrorValueStringAsIDNotValid("01234abcd")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)),
			Entry("has string length out of range (upper)", "01234abcdef", user.ErrorValueStringAsIDNotValid("01234abcdef")),
			Entry("has uppercase characters", "01234ABCDE", user.ErrorValueStringAsIDNotValid("01234ABCDE")),
			Entry("has symbols", "0123$%BCDE", user.ErrorValueStringAsIDNotValid("0123$%BCDE")),
			Entry("has whitespace", "0123  BCDE", user.ErrorValueStringAsIDNotValid("0123  BCDE")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsIDNotValid with empty string", user.ErrorValueStringAsIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as user id`),
			Entry("is ErrorValueStringAsIDNotValid with non-empty string", user.ErrorValueStringAsIDNotValid("01234abcde"), "value-not-valid", "value is not valid", `value "01234abcde" is not valid as user id`),
		)
	})
})
