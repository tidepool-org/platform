package unstructured_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	storeUnstructuredTest "github.com/tidepool-org/platform/store/unstructured/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Unstructured", func() {
	Context("Options", func() {
		Context("NewOptions", func() {
			It("returns successfully with default values", func() {
				Expect(storeUnstructured.NewOptions()).To(Equal(&storeUnstructured.Options{}))
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *storeUnstructured.Options), expectedErrors ...error) {
					datum := storeUnstructuredTest.RandomOptions()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *storeUnstructured.Options) {},
				),
				Entry("media type missing",
					func(datum *storeUnstructured.Options) { datum.MediaType = nil },
				),
				Entry("media type empty",
					func(datum *storeUnstructured.Options) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *storeUnstructured.Options) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *storeUnstructured.Options) {
						datum.MediaType = pointer.FromString(netTest.RandomMediaType())
					},
				),
				Entry("multiple errors",
					func(datum *storeUnstructured.Options) {
						datum.MediaType = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
			)
		})
	})

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
			Entry("is valid", "abc_123/D=EF-456.g7"),
			Entry("starts with a slash", "/abc_123/D=EF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("/abc_123/D=EF-456.g7")),
			Entry("starts with a period", ".abc_123/D=EF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid(".abc_123/D=EF-456.g7")),
			Entry("starts with an underscore", "_abc_123/D=EF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("_abc_123/D=EF-456.g7")),
			Entry("starts with a dash", "-abc_123/D=EF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("-abc_123/D=EF-456.g7")),
			Entry("contains a non-ASCII character", "abc😁123/D=EF-456.g7", storeUnstructured.ErrorValueStringAsKeyNotValid("abc😁123/D=EF-456.g7")),
			Entry("contains no slashes", "D=EF-456.g7"),
			Entry("contains multiple slashes", "abc_123/abc_123/abc_123/D=EF-456.g7"),
			Entry("has length in range (upper)", test.RandomStringFromRangeAndCharset(2047, 2047, test.CharsetAlphaNumeric)),
			Entry("has length out of range (upper)", test.RandomStringFromRangeAndCharset(2048, 2048, test.CharsetAlphaNumeric), structureValidator.ErrorLengthNotLessThanOrEqualTo(2048, 2047)),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsKeyNotValid with empty string", storeUnstructured.ErrorValueStringAsKeyNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as unstructured key`),
			Entry("is ErrorValueStringAsKeyNotValid with non-empty string", storeUnstructured.ErrorValueStringAsKeyNotValid("abc_123/D=EF-456.g7"), "value-not-valid", "value is not valid", `value "abc_123/D=EF-456.g7" is not valid as unstructured key`),
		)
	})
})
