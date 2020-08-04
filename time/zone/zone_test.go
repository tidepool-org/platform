package zone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
	Context("GetEtcZone", func() {
		It("returns an Etc/Gmt-{offset} when passing positive offset", func() {
			Expect(timeZone.GetEtcZone(120)).To(Equal("Etc/GMT-2"))
		})
		It("returns an Etc/Gmt+{offset} when passing negative offset", func() {
			Expect(timeZone.GetEtcZone(-240)).To(Equal("Etc/GMT+4"))
		})
		It("returns an Etc/Gmt when passing 0 offset", func() {
			Expect(timeZone.GetEtcZone(0)).To(Equal("Etc/GMT"))
		})
	})
})
