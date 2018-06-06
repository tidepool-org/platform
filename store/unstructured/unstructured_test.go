package unstructured_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Unstructured", func() {
	Context("IsValidKey, KeyValidator, and ValidateKey", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(storeUnstructured.IsValidKey(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				storeUnstructured.KeyValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(storeUnstructured.ValidateKey(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is valid", "abc_123/DEF-456.g7"),
			Entry("starts with a slash", "/abc_123/DEF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("/abc_123/DEF-456.g7")),
			Entry("starts with a period", ".abc_123/DEF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid(".abc_123/DEF-456.g7")),
			Entry("starts with an underscore", "_abc_123/DEF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("_abc_123/DEF-456.g7")),
			Entry("starts with a dash", "-abc_123/DEF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("-abc_123/DEF-456.g7")),
			Entry("contains a non-ASCII character", "abc😁123/DEF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("abc😁123/DEF-456.g7")),
			Entry("contains no slashes", "DEF-456.g7"),
			Entry("contains multiple slashes", "abc_123/abc_123/abc_123/DEF-456.g7"),
			Entry("has length in range (upper)", test.NewString(2047, test.CharsetAlphaNumeric)),
			Entry("has length out of range (upper)", test.NewString(2048, test.CharsetAlphaNumeric), structureValidator.ErrorLengthNotLessThanOrEqualTo(2048, 2047)),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsKeyNotValid with empty string", storeUnstructured.ErrorValueStringAsKeyNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as unstructured key`),
			Entry("is ErrorValueStringAsKeyNotValid with non-empty string", storeUnstructured.ErrorValueStringAsKeyNotValid("abc_123/DEF-456.g7"), "value-not-valid", "value is not valid", `value "abc_123/DEF-456.g7" is not valid as unstructured key`),
		)
	})
})
