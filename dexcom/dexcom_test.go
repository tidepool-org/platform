package dexcom_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Dexcom", func() {
	It("TimeFormat is expected", func() {
		Expect(dexcom.TimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("SystemTimeNowThreshold is expected", func() {
		Expect(dexcom.SystemTimeNowThreshold).To(Equal(24 * time.Hour))
	})

	Context("IsValidTransmitterID, TransmitterIDValidator, and ValidateTransmitterID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dexcom.IsValidTransmitterID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dexcom.TransmitterIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dexcom.ValidateTransmitterID(value), expectedErrors...)
			},
			Entry("is an empty string", ""),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(5, 6, test.CharsetNumeric+test.CharsetUppercase)),
			Entry("has string length out of range (lower)", "0123", dexcom.ErrorValueStringAsTransmitterIDNotValid("0123")),
			Entry("has string length out of range (upper)", "0123456", dexcom.ErrorValueStringAsTransmitterIDNotValid("0123456")),
			Entry("has lowercase characters", "abcdef", dexcom.ErrorValueStringAsTransmitterIDNotValid("abcdef")),
			Entry("has symbols", "$%^&*(", dexcom.ErrorValueStringAsTransmitterIDNotValid("$%^&*(")),
			Entry("has whitespace", "a    b", dexcom.ErrorValueStringAsTransmitterIDNotValid("a    b")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsTransmitterIDNotValid with empty string", dexcom.ErrorValueStringAsTransmitterIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as transmitter id`),
			Entry("is ErrorValueStringAsTransmitterIDNotValid with non-empty string", dexcom.ErrorValueStringAsTransmitterIDNotValid("abcdef"), "value-not-valid", "value is not valid", `value "abcdef" is not valid as transmitter id`),
		)
	})
})
