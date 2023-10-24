package blob_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/tidepool-org/platform/blob"
	errorsTest "github.com/tidepool-org/platform/errors/test"
)

var _ = Describe("Errors", func() {
	DescribeTable("have expected details when error",
		errorsTest.ExpectErrorDetails,
		Entry("is ErrorValueStringAsIDNotValid", blob.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as blob id`),
	)
})
