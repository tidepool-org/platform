package user_test

import (
	. "github.com/onsi/ginkgo/v2"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/user"
)

var _ = Describe("Errors", func() {
	DescribeTable("have expected details when error",
		errorsTest.ExpectErrorDetails,
		Entry("is ErrorValueStringAsIDNotValid", user.ErrorValueStringAsIDNotValid("01234abcde"), "value-not-valid", "value is not valid", `value "01234abcde" is not valid as user id`),
	)
})
