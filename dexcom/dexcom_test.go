package dexcom_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Dexcom", func() {
	It("TimeFormatMilli is expected", func() {
		Expect(dexcom.TimeFormatMilli).To(Equal("2006-01-02T15:04:05.999"))
	})

	It("TimeFormatMilliZ is expected", func() {
		Expect(dexcom.TimeFormatMilliZ).To(Equal("2006-01-02T15:04:05.999-07:00"))
	})

	It("TimeFormatMilliUTC is expected", func() {
		Expect(dexcom.TimeFormatMilliUTC).To(Equal("2006-01-02T15:04:05.999Z"))
	})

	It("TimeFormat is expected", func() {
		Expect(dexcom.TimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("TimeMinimalFormat is expected", func() {
		Expect(dexcom.TimeMinimalFormat).To(Equal("2006-01-02T15:04"))
	})

	It("SystemTimeNowThreshold is expected", func() {
		Expect(dexcom.SystemTimeNowThreshold).To(Equal(24 * time.Hour))
	})

	Context("IsValidTransmitterID, TransmitterIDValidator, and ValidateTransmitterID", func() {

		const validTransmitterId = "6f1c584eb070e0e7ec3f8a9af313c34028374eee50928be47d807f333891369f"

		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dexcom.IsValidTransmitterID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dexcom.TransmitterIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dexcom.ValidateTransmitterID(value), expectedErrors...)
			},
			Entry("is an empty string", ""),
			Entry("has string length in range", validTransmitterId),
			Entry("has string length out of range (lower)", validTransmitterId[:40], dexcom.ErrorValueStringAsTransmitterIDNotValid(validTransmitterId[:40])),
			Entry("has string length out of range (upper)", validTransmitterId+"a", dexcom.ErrorValueStringAsTransmitterIDNotValid(validTransmitterId+"a")),
			Entry("has uppercase characters", strings.ToUpper(validTransmitterId), dexcom.ErrorValueStringAsTransmitterIDNotValid(strings.ToUpper(validTransmitterId))),
			Entry("has symbols", strings.ReplaceAll(validTransmitterId, "a", "$"), dexcom.ErrorValueStringAsTransmitterIDNotValid(strings.ReplaceAll(validTransmitterId, "a", "$"))),
			Entry("has whitespace", strings.ReplaceAll(validTransmitterId, "a", " "), dexcom.ErrorValueStringAsTransmitterIDNotValid(strings.ReplaceAll(validTransmitterId, "a", " "))),
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
