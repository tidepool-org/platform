package validator_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Errors", func() {
	DescribeTable("all errors",
		func(err *service.Error, code string, title string, detail string) {
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(code))
			Expect(err.Title).To(Equal(title))
			Expect(err.Detail).To(Equal(detail))
		},
		Entry("ErrorValueDoesNotExist", validator.ErrorValueDoesNotExist(), "value-does-not-exist", "value does not exist", "Value does not exist"),
		Entry("ErrorValueNotTrue", validator.ErrorValueNotTrue(), "value-not-true", "value is not true", "Value is not true"),
		Entry("ErrorValueNotFalse", validator.ErrorValueNotFalse(), "value-not-false", "value is not false", "Value is not false"),
		Entry("ErrorValueNotEqualTo with int", validator.ErrorValueNotEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not equal to 2"),
		Entry("ErrorValueNotEqualTo with float", validator.ErrorValueNotEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not equal to 5.6"),
		Entry("ErrorValueNotEqualTo with string", validator.ErrorValueNotEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not equal to \"xyz\""),
		Entry("ErrorValueNotEqualTo with string with quotes", validator.ErrorValueNotEqualTo("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not equal to \"x\\\"y\\\"z\""),
		Entry("ErrorValueEqualTo with int", validator.ErrorValueEqualTo(2, 2), "value-out-of-range", "value is out of range", "Value 2 is equal to 2"),
		Entry("ErrorValueEqualTo with float", validator.ErrorValueEqualTo(5.6, 5.6), "value-out-of-range", "value is out of range", "Value 5.6 is equal to 5.6"),
		Entry("ErrorValueEqualTo with string", validator.ErrorValueEqualTo("xyz", "xyz"), "value-out-of-range", "value is out of range", "Value \"xyz\" is equal to \"xyz\""),
		Entry("ErrorValueEqualTo with string with quotes", validator.ErrorValueEqualTo("x\"y\"z", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is equal to \"x\\\"y\\\"z\""),
		Entry("ErrorValueNotLessThan with int", validator.ErrorValueNotLessThan(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than 1"),
		Entry("ErrorValueNotLessThan with float", validator.ErrorValueNotLessThan(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than 3.4"),
		Entry("ErrorValueNotLessThan with string", validator.ErrorValueNotLessThan("xyz", "abc"), "value-out-of-range", "value is out of range", "Value \"xyz\" is not less than \"abc\""),
		Entry("ErrorValueNotLessThan with string with quotes", validator.ErrorValueNotLessThan("x\"y\"z", "a\"b\"c"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is not less than \"a\\\"b\\\"c\""),
		Entry("ErrorValueNotLessThanOrEqualTo with int", validator.ErrorValueNotLessThanOrEqualTo(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than or equal to 1"),
		Entry("ErrorValueNotLessThanOrEqualTo with float", validator.ErrorValueNotLessThanOrEqualTo(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than or equal to 3.4"),
		Entry("ErrorValueNotLessThanOrEqualTo with string", validator.ErrorValueNotLessThanOrEqualTo("xyz", "abc"), "value-out-of-range", "value is out of range", "Value \"xyz\" is not less than or equal to \"abc\""),
		Entry("ErrorValueNotLessThanOrEqualTo with string with quotes", validator.ErrorValueNotLessThanOrEqualTo("x\"y\"z", "a\"b\"c"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is not less than or equal to \"a\\\"b\\\"c\""),
		Entry("ErrorValueNotGreaterThan with int", validator.ErrorValueNotGreaterThan(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than 2"),
		Entry("ErrorValueNotGreaterThan with float", validator.ErrorValueNotGreaterThan(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than 5.6"),
		Entry("ErrorValueNotGreaterThan with string", validator.ErrorValueNotGreaterThan("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not greater than \"xyz\""),
		Entry("ErrorValueNotGreaterThan with string with quotes", validator.ErrorValueNotGreaterThan("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not greater than \"x\\\"y\\\"z\""),
		Entry("ErrorValueNotGreaterThanOrEqualTo with int", validator.ErrorValueNotGreaterThanOrEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than or equal to 2"),
		Entry("ErrorValueNotGreaterThanOrEqualTo with float", validator.ErrorValueNotGreaterThanOrEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than or equal to 5.6"),
		Entry("ErrorValueNotGreaterThanOrEqualTo with string", validator.ErrorValueNotGreaterThanOrEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not greater than or equal to \"xyz\""),
		Entry("ErrorValueNotGreaterThanOrEqualTo with string with quotes", validator.ErrorValueNotGreaterThanOrEqualTo("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not greater than or equal to \"x\\\"y\\\"z\""),
		Entry("ErrorIntegerNotInRange", validator.ErrorIntegerNotInRange(1, 2, 3), "value-out-of-range", "value is out of range", "Value 1 is not between 2 and 3"),
		Entry("ErrorFloatNotInRange", validator.ErrorFloatNotInRange(1.4, 2.4, 3.4), "value-out-of-range", "value is out of range", "Value 1.4 is not between 2.4 and 3.4"),
		Entry("ErrorIntegerOneOf with nil array", validator.ErrorIntegerOneOf(2, nil), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []"),
		Entry("ErrorIntegerOneOf with empty array", validator.ErrorIntegerOneOf(2, []int{}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []"),
		Entry("ErrorIntegerOneOf with non-empty array", validator.ErrorIntegerOneOf(2, []int{2, 3, 4}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of [2, 3, 4]"),
		Entry("ErrorIntegerNotOneOf with nil array", validator.ErrorIntegerNotOneOf(1, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []"),
		Entry("ErrorIntegerNotOneOf with empty array", validator.ErrorIntegerNotOneOf(1, []int{}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []"),
		Entry("ErrorIntegerNotOneOf with non-empty array", validator.ErrorIntegerNotOneOf(1, []int{2, 3, 4}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of [2, 3, 4]"),
		Entry("ErrorFloatOneOf with nil array", validator.ErrorFloatOneOf(2.5, nil), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []"),
		Entry("ErrorFloatOneOf with empty array", validator.ErrorFloatOneOf(2.5, []float64{}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []"),
		Entry("ErrorFloatOneOf with non-empty array", validator.ErrorFloatOneOf(2.5, []float64{2.5, 3.5, 4.5}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of [2.5, 3.5, 4.5]"),
		Entry("ErrorFloatNotOneOf with nil array", validator.ErrorFloatNotOneOf(1.5, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []"),
		Entry("ErrorFloatNotOneOf with empty array", validator.ErrorFloatNotOneOf(1.5, []float64{}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []"),
		Entry("ErrorFloatNotOneOf with non-empty array", validator.ErrorFloatNotOneOf(1.5, []float64{2.5, 3.5, 4.5}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of [2.5, 3.5, 4.5]"),
		Entry("ErrorLengthNotEqualTo with int", validator.ErrorLengthNotEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not equal to 2"),
		Entry("ErrorLengthEqualTo with int", validator.ErrorLengthEqualTo(2, 2), "length-out-of-range", "length is out of range", "Length 2 is equal to 2"),
		Entry("ErrorLengthNotLessThan with int", validator.ErrorLengthNotLessThan(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than 1"),
		Entry("ErrorLengthNotLessThanOrEqualTo with int", validator.ErrorLengthNotLessThanOrEqualTo(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than or equal to 1"),
		Entry("ErrorLengthNotGreaterThan with int", validator.ErrorLengthNotGreaterThan(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than 2"),
		Entry("ErrorLengthNotGreaterThanOrEqualTo with int", validator.ErrorLengthNotGreaterThanOrEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than or equal to 2"),
		Entry("ErrorLengthNotInRange", validator.ErrorLengthNotInRange(1, 2, 3), "length-out-of-range", "length is out of range", "Length 1 is not between 2 and 3"),
		Entry("ErrorStringOneOf with nil array", validator.ErrorStringOneOf("abc", nil), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of []"),
		Entry("ErrorStringOneOf with empty array", validator.ErrorStringOneOf("abc", []string{}), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of []"),
		Entry("ErrorStringOneOf with non-empty array", validator.ErrorStringOneOf("abc", []string{"abc", "bcd", "cde"}), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of [\"abc\", \"bcd\", \"cde\"]"),
		Entry("ErrorStringNotOneOf with nil array", validator.ErrorStringNotOneOf("xyz", nil), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of []"),
		Entry("ErrorStringNotOneOf with empty array", validator.ErrorStringNotOneOf("xyz", []string{}), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of []"),
		Entry("ErrorStringNotOneOf with non-empty array", validator.ErrorStringNotOneOf("xyz", []string{"abc", "bcd", "cde"}), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of [\"abc\", \"bcd\", \"cde\"]"),
		Entry("ErrorTimeNotValid", validator.ErrorTimeNotValid("abc", "2006-01-02T15:04:05Z07:00"), "time-not-valid", "value is not a valid time", "Value \"abc\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""),
		Entry("ErrorTimeNotAfter", validator.ErrorTimeNotAfter(time.Unix(1451567655, 0).UTC(), time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-after", "value is not after the specified time", "Value \"2015-12-31T13:14:15Z\" is not after \"2025-01-01T13:14:15Z\""),
		Entry("ErrorTimeNotAfterNow", validator.ErrorTimeNotAfterNow(time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-after", "value is not after the specified time", "Value \"2015-12-31T13:14:15Z\" is not after now"),
		Entry("ErrorTimeNotBefore", validator.ErrorTimeNotBefore(time.Unix(1735737255, 0).UTC(), time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-before", "value is not before the specified time", "Value \"2025-01-01T13:14:15Z\" is not before \"2015-12-31T13:14:15Z\""),
		Entry("ErrorTimeNotBeforeNow", validator.ErrorTimeNotBeforeNow(time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-before", "value is not before the specified time", "Value \"2025-01-01T13:14:15Z\" is not before now"),
	)
})
