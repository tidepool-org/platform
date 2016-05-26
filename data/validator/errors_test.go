package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

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
		Entry("is ErrorValueDoesNotExist", validator.ErrorValueDoesNotExist(), "value-does-not-exist", "value does not exist", "Value does not exist"),
		Entry("is ErrorValueNotTrue", validator.ErrorValueNotTrue(), "value-not-true", "value is not true", "Value is not true"),
		Entry("is ErrorValueNotFalse", validator.ErrorValueNotFalse(), "value-not-false", "value is not false", "Value is not false"),
		Entry("is ErrorValueNotEqualTo with int", validator.ErrorValueNotEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not equal to 2"),
		Entry("is ErrorValueNotEqualTo with float", validator.ErrorValueNotEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not equal to 5.6"),
		Entry("is ErrorValueNotEqualTo with string", validator.ErrorValueNotEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not equal to \"xyz\""),
		Entry("is ErrorValueNotEqualTo with string with quotes", validator.ErrorValueNotEqualTo("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not equal to \"x\\\"y\\\"z\""),
		Entry("is ErrorValueEqualTo with int", validator.ErrorValueEqualTo(2, 2), "value-out-of-range", "value is out of range", "Value 2 is equal to 2"),
		Entry("is ErrorValueEqualTo with float", validator.ErrorValueEqualTo(5.6, 5.6), "value-out-of-range", "value is out of range", "Value 5.6 is equal to 5.6"),
		Entry("is ErrorValueEqualTo with string", validator.ErrorValueEqualTo("xyz", "xyz"), "value-out-of-range", "value is out of range", "Value \"xyz\" is equal to \"xyz\""),
		Entry("is ErrorValueEqualTo with string with quotes", validator.ErrorValueEqualTo("x\"y\"z", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is equal to \"x\\\"y\\\"z\""),
		Entry("is ErrorValueNotLessThan with int", validator.ErrorValueNotLessThan(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than 1"),
		Entry("is ErrorValueNotLessThan with float", validator.ErrorValueNotLessThan(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than 3.4"),
		Entry("is ErrorValueNotLessThan with string", validator.ErrorValueNotLessThan("xyz", "abc"), "value-out-of-range", "value is out of range", "Value \"xyz\" is not less than \"abc\""),
		Entry("is ErrorValueNotLessThan with string with quotes", validator.ErrorValueNotLessThan("x\"y\"z", "a\"b\"c"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is not less than \"a\\\"b\\\"c\""),
		Entry("is ErrorValueNotLessThanOrEqualTo with int", validator.ErrorValueNotLessThanOrEqualTo(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than or equal to 1"),
		Entry("is ErrorValueNotLessThanOrEqualTo with float", validator.ErrorValueNotLessThanOrEqualTo(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than or equal to 3.4"),
		Entry("is ErrorValueNotLessThanOrEqualTo with string", validator.ErrorValueNotLessThanOrEqualTo("xyz", "abc"), "value-out-of-range", "value is out of range", "Value \"xyz\" is not less than or equal to \"abc\""),
		Entry("is ErrorValueNotLessThanOrEqualTo with string with quotes", validator.ErrorValueNotLessThanOrEqualTo("x\"y\"z", "a\"b\"c"), "value-out-of-range", "value is out of range", "Value \"x\\\"y\\\"z\" is not less than or equal to \"a\\\"b\\\"c\""),
		Entry("is ErrorValueNotGreaterThan with int", validator.ErrorValueNotGreaterThan(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than 2"),
		Entry("is ErrorValueNotGreaterThan with float", validator.ErrorValueNotGreaterThan(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than 5.6"),
		Entry("is ErrorValueNotGreaterThan with string", validator.ErrorValueNotGreaterThan("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not greater than \"xyz\""),
		Entry("is ErrorValueNotGreaterThan with string with quotes", validator.ErrorValueNotGreaterThan("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not greater than \"x\\\"y\\\"z\""),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with int", validator.ErrorValueNotGreaterThanOrEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than or equal to 2"),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with float", validator.ErrorValueNotGreaterThanOrEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than or equal to 5.6"),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string", validator.ErrorValueNotGreaterThanOrEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", "Value \"abc\" is not greater than or equal to \"xyz\""),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string with quotes", validator.ErrorValueNotGreaterThanOrEqualTo("a\"b\"c", "x\"y\"z"), "value-out-of-range", "value is out of range", "Value \"a\\\"b\\\"c\" is not greater than or equal to \"x\\\"y\\\"z\""),
		Entry("is ErrorIntegerNotInRange", validator.ErrorIntegerNotInRange(1, 2, 3), "value-out-of-range", "value is out of range", "Value 1 is not between 2 and 3"),
		Entry("is ErrorFloatNotInRange", validator.ErrorFloatNotInRange(1.4, 2.4, 3.4), "value-out-of-range", "value is out of range", "Value 1.4 is not between 2.4 and 3.4"),
		Entry("is ErrorIntegerOneOf with nil array", validator.ErrorIntegerOneOf(2, nil), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []"),
		Entry("is ErrorIntegerOneOf with empty array", validator.ErrorIntegerOneOf(2, []int{}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []"),
		Entry("is ErrorIntegerOneOf with non-empty array", validator.ErrorIntegerOneOf(2, []int{2, 3, 4}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of [2, 3, 4]"),
		Entry("is ErrorIntegerNotOneOf with nil array", validator.ErrorIntegerNotOneOf(1, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []"),
		Entry("is ErrorIntegerNotOneOf with empty array", validator.ErrorIntegerNotOneOf(1, []int{}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []"),
		Entry("is ErrorIntegerNotOneOf with non-empty array", validator.ErrorIntegerNotOneOf(1, []int{2, 3, 4}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of [2, 3, 4]"),
		Entry("is ErrorFloatOneOf with nil array", validator.ErrorFloatOneOf(2.5, nil), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []"),
		Entry("is ErrorFloatOneOf with empty array", validator.ErrorFloatOneOf(2.5, []float64{}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []"),
		Entry("is ErrorFloatOneOf with non-empty array", validator.ErrorFloatOneOf(2.5, []float64{2.5, 3.5, 4.5}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of [2.5, 3.5, 4.5]"),
		Entry("is ErrorFloatNotOneOf with nil array", validator.ErrorFloatNotOneOf(1.5, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []"),
		Entry("is ErrorFloatNotOneOf with empty array", validator.ErrorFloatNotOneOf(1.5, []float64{}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []"),
		Entry("is ErrorFloatNotOneOf with non-empty array", validator.ErrorFloatNotOneOf(1.5, []float64{2.5, 3.5, 4.5}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of [2.5, 3.5, 4.5]"),
		Entry("is ErrorLengthNotEqualTo with int", validator.ErrorLengthNotEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not equal to 2"),
		Entry("is ErrorLengthEqualTo with int", validator.ErrorLengthEqualTo(2, 2), "length-out-of-range", "length is out of range", "Length 2 is equal to 2"),
		Entry("is ErrorLengthNotLessThan with int", validator.ErrorLengthNotLessThan(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than 1"),
		Entry("is ErrorLengthNotLessThanOrEqualTo with int", validator.ErrorLengthNotLessThanOrEqualTo(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than or equal to 1"),
		Entry("is ErrorLengthNotGreaterThan with int", validator.ErrorLengthNotGreaterThan(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than 2"),
		Entry("is ErrorLengthNotGreaterThanOrEqualTo with int", validator.ErrorLengthNotGreaterThanOrEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than or equal to 2"),
		Entry("is ErrorLengthNotInRange", validator.ErrorLengthNotInRange(1, 2, 3), "length-out-of-range", "length is out of range", "Length 1 is not between 2 and 3"),
		Entry("is ErrorStringOneOf with nil array", validator.ErrorStringOneOf("abc", nil), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of []"),
		Entry("is ErrorStringOneOf with empty array", validator.ErrorStringOneOf("abc", []string{}), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of []"),
		Entry("is ErrorStringOneOf with non-empty array", validator.ErrorStringOneOf("abc", []string{"abc", "bcd", "cde"}), "value-disallowed", "value is one of the disallowed values", "Value \"abc\" is one of [\"abc\", \"bcd\", \"cde\"]"),
		Entry("is ErrorStringNotOneOf with nil array", validator.ErrorStringNotOneOf("xyz", nil), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of []"),
		Entry("is ErrorStringNotOneOf with empty array", validator.ErrorStringNotOneOf("xyz", []string{}), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of []"),
		Entry("is ErrorStringNotOneOf with non-empty array", validator.ErrorStringNotOneOf("xyz", []string{"abc", "bcd", "cde"}), "value-not-allowed", "value is not one of the allowed values", "Value \"xyz\" is not one of [\"abc\", \"bcd\", \"cde\"]"),
		Entry("is ErrorTimeNotValid", validator.ErrorTimeNotValid("abc", "2006-01-02T15:04:05Z07:00"), "time-not-valid", "value is not a valid time", "Value \"abc\" is not a valid time of format \"2006-01-02T15:04:05Z07:00\""),
		Entry("is ErrorTimeNotAfter", validator.ErrorTimeNotAfter(time.Unix(1451567655, 0).UTC(), time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-after", "value is not after the specified time", "Value \"2015-12-31T13:14:15Z\" is not after \"2025-01-01T13:14:15Z\""),
		Entry("is ErrorTimeNotAfterNow", validator.ErrorTimeNotAfterNow(time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-after", "value is not after the specified time", "Value \"2015-12-31T13:14:15Z\" is not after now"),
		Entry("is ErrorTimeNotBefore", validator.ErrorTimeNotBefore(time.Unix(1735737255, 0).UTC(), time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-before", "value is not before the specified time", "Value \"2025-01-01T13:14:15Z\" is not before \"2015-12-31T13:14:15Z\""),
		Entry("is ErrorTimeNotBeforeNow", validator.ErrorTimeNotBeforeNow(time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "time-not-before", "value is not before the specified time", "Value \"2025-01-01T13:14:15Z\" is not before now"),
	)
})
