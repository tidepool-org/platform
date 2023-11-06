package zone_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	timeZone "github.com/tidepool-org/platform/time/zone"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
)

var _ = Describe("Zone", func() {
	Context("Names", func() {
		It("returns a non-empty array", func() {
			Expect(timeZone.Names()).ToNot(BeEmpty())
		})
	})

	Context("IsValidName, NameValidator, and ValidateName", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(timeZone.IsValidName(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				timeZone.NameValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(timeZone.ValidateName(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is invalid", "invalid", timeZone.ErrorValueStringAsNameNotValid("invalid")),
			Entry("is valid", timeZoneTest.RandomName()),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsNameNotValid with empty string", timeZone.ErrorValueStringAsNameNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid time zone name`),
			Entry("is ErrorValueStringAsNameNotValid with non-empty string", timeZone.ErrorValueStringAsNameNotValid("invalid"), "value-not-valid", "value is not valid", `value "invalid" is not valid time zone name`),
		)
	})
})
