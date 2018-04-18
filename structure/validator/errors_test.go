package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Errors", func() {
	DescribeTable("all errors",
		func(err error, code string, title string, detail string) {
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal(code))
			Expect(errors.Cause(err)).To(Equal(err))
			bytes, bytesErr := json.Marshal(errors.Sanitize(err))
			Expect(bytesErr).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
		},
		Entry("is ErrorValueNotExists", structureValidator.ErrorValueNotExists(), "value-not-exists", "value does not exist", "value does not exist"),
		Entry("is ErrorValueExists", structureValidator.ErrorValueExists(), "value-exists", "value exists", "value exists"),
		Entry("is ErrorValueNotEmpty", structureValidator.ErrorValueNotEmpty(), "value-not-empty", "value is not empty", "value is not empty"),
		Entry("is ErrorValueEmpty", structureValidator.ErrorValueEmpty(), "value-empty", "value is empty", "value is empty"),
		Entry("is ErrorValueDuplicate", structureValidator.ErrorValueDuplicate(), "value-duplicate", "value is a duplicate", "value is a duplicate"),
		Entry("is ErrorValueBooleanNotTrue", structureValidator.ErrorValueBooleanNotTrue(), "value-not-true", "value is not true", "value is not true"),
		Entry("is ErrorValueBooleanNotFalse", structureValidator.ErrorValueBooleanNotFalse(), "value-not-false", "value is not false", "value is not false"),
		Entry("is ErrorValueNotEqualTo with int", structureValidator.ErrorValueNotEqualTo(1, 2), "value-out-of-range", "value is out of range", "value 1 is not equal to 2"),
		Entry("is ErrorValueNotEqualTo with float", structureValidator.ErrorValueNotEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "value 3.4 is not equal to 5.6"),
		Entry("is ErrorValueNotEqualTo with string", structureValidator.ErrorValueNotEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", `value "abc" is not equal to "xyz"`),
		Entry("is ErrorValueNotEqualTo with string with quotes", structureValidator.ErrorValueNotEqualTo(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `value "a\"b\"c" is not equal to "x\"y\"z"`),
		Entry("is ErrorValueEqualTo with int", structureValidator.ErrorValueEqualTo(2, 2), "value-out-of-range", "value is out of range", "value 2 is equal to 2"),
		Entry("is ErrorValueEqualTo with float", structureValidator.ErrorValueEqualTo(5.6, 5.6), "value-out-of-range", "value is out of range", "value 5.6 is equal to 5.6"),
		Entry("is ErrorValueEqualTo with string", structureValidator.ErrorValueEqualTo("xyz", "xyz"), "value-out-of-range", "value is out of range", `value "xyz" is equal to "xyz"`),
		Entry("is ErrorValueEqualTo with string with quotes", structureValidator.ErrorValueEqualTo(`x"y"z`, `x"y"z`), "value-out-of-range", "value is out of range", `value "x\"y\"z" is equal to "x\"y\"z"`),
		Entry("is ErrorValueNotLessThan with int", structureValidator.ErrorValueNotLessThan(2, 1), "value-out-of-range", "value is out of range", "value 2 is not less than 1"),
		Entry("is ErrorValueNotLessThan with float", structureValidator.ErrorValueNotLessThan(5.6, 3.4), "value-out-of-range", "value is out of range", "value 5.6 is not less than 3.4"),
		Entry("is ErrorValueNotLessThan with string", structureValidator.ErrorValueNotLessThan("xyz", "abc"), "value-out-of-range", "value is out of range", `value "xyz" is not less than "abc"`),
		Entry("is ErrorValueNotLessThan with string with quotes", structureValidator.ErrorValueNotLessThan(`x"y"z`, `a"b"c`), "value-out-of-range", "value is out of range", `value "x\"y\"z" is not less than "a\"b\"c"`),
		Entry("is ErrorValueNotLessThanOrEqualTo with int", structureValidator.ErrorValueNotLessThanOrEqualTo(2, 1), "value-out-of-range", "value is out of range", "value 2 is not less than or equal to 1"),
		Entry("is ErrorValueNotLessThanOrEqualTo with float", structureValidator.ErrorValueNotLessThanOrEqualTo(5.6, 3.4), "value-out-of-range", "value is out of range", "value 5.6 is not less than or equal to 3.4"),
		Entry("is ErrorValueNotLessThanOrEqualTo with string", structureValidator.ErrorValueNotLessThanOrEqualTo("xyz", "abc"), "value-out-of-range", "value is out of range", `value "xyz" is not less than or equal to "abc"`),
		Entry("is ErrorValueNotLessThanOrEqualTo with string with quotes", structureValidator.ErrorValueNotLessThanOrEqualTo(`x"y"z`, `a"b"c`), "value-out-of-range", "value is out of range", `value "x\"y\"z" is not less than or equal to "a\"b\"c"`),
		Entry("is ErrorValueNotGreaterThan with int", structureValidator.ErrorValueNotGreaterThan(1, 2), "value-out-of-range", "value is out of range", "value 1 is not greater than 2"),
		Entry("is ErrorValueNotGreaterThan with float", structureValidator.ErrorValueNotGreaterThan(3.4, 5.6), "value-out-of-range", "value is out of range", "value 3.4 is not greater than 5.6"),
		Entry("is ErrorValueNotGreaterThan with string", structureValidator.ErrorValueNotGreaterThan("abc", "xyz"), "value-out-of-range", "value is out of range", `value "abc" is not greater than "xyz"`),
		Entry("is ErrorValueNotGreaterThan with string with quotes", structureValidator.ErrorValueNotGreaterThan(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `value "a\"b\"c" is not greater than "x\"y\"z"`),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with int", structureValidator.ErrorValueNotGreaterThanOrEqualTo(1, 2), "value-out-of-range", "value is out of range", "value 1 is not greater than or equal to 2"),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with float", structureValidator.ErrorValueNotGreaterThanOrEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "value 3.4 is not greater than or equal to 5.6"),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string", structureValidator.ErrorValueNotGreaterThanOrEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", `value "abc" is not greater than or equal to "xyz"`),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string with quotes", structureValidator.ErrorValueNotGreaterThanOrEqualTo(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `value "a\"b\"c" is not greater than or equal to "x\"y\"z"`),
		Entry("is ErrorValueNotInRange with int", structureValidator.ErrorValueNotInRange(1, 2, 3), "value-out-of-range", "value is out of range", "value 1 is not between 2 and 3"),
		Entry("is ErrorValueNotInRange with float", structureValidator.ErrorValueNotInRange(1.4, 2.4, 3.4), "value-out-of-range", "value is out of range", "value 1.4 is not between 2.4 and 3.4"),
		Entry("is ErrorValueNotInRange with string", structureValidator.ErrorValueNotInRange("zzz", "abc", "xyz"), "value-out-of-range", "value is out of range", `value "zzz" is not between "abc" and "xyz"`),
		Entry("is ErrorValueNotInRange with string with quotes", structureValidator.ErrorValueNotInRange(`z"z"z`, `a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `value "z\"z\"z" is not between "a\"b\"c" and "x\"y\"z"`),
		Entry("is ErrorValueFloat64OneOf with nil array", structureValidator.ErrorValueFloat64OneOf(2.5, nil), "value-disallowed", "value is one of the disallowed values", "value 2.5 is one of []"),
		Entry("is ErrorValueFloat64OneOf with empty array", structureValidator.ErrorValueFloat64OneOf(2.5, []float64{}), "value-disallowed", "value is one of the disallowed values", "value 2.5 is one of []"),
		Entry("is ErrorValueFloat64OneOf with non-empty array", structureValidator.ErrorValueFloat64OneOf(2.5, []float64{2.5, 3.5, 4.5}), "value-disallowed", "value is one of the disallowed values", "value 2.5 is one of [2.5, 3.5, 4.5]"),
		Entry("is ErrorValueFloat64NotOneOf with nil array", structureValidator.ErrorValueFloat64NotOneOf(1.5, nil), "value-not-allowed", "value is not one of the allowed values", "value 1.5 is not one of []"),
		Entry("is ErrorValueFloat64NotOneOf with empty array", structureValidator.ErrorValueFloat64NotOneOf(1.5, []float64{}), "value-not-allowed", "value is not one of the allowed values", "value 1.5 is not one of []"),
		Entry("is ErrorValueFloat64NotOneOf with non-empty array", structureValidator.ErrorValueFloat64NotOneOf(1.5, []float64{2.5, 3.5, 4.5}), "value-not-allowed", "value is not one of the allowed values", "value 1.5 is not one of [2.5, 3.5, 4.5]"),
		Entry("is ErrorValueIntOneOf with nil array", structureValidator.ErrorValueIntOneOf(2, nil), "value-disallowed", "value is one of the disallowed values", "value 2 is one of []"),
		Entry("is ErrorValueIntOneOf with empty array", structureValidator.ErrorValueIntOneOf(2, []int{}), "value-disallowed", "value is one of the disallowed values", "value 2 is one of []"),
		Entry("is ErrorValueIntOneOf with non-empty array", structureValidator.ErrorValueIntOneOf(2, []int{2, 3, 4}), "value-disallowed", "value is one of the disallowed values", "value 2 is one of [2, 3, 4]"),
		Entry("is ErrorValueIntNotOneOf with nil array", structureValidator.ErrorValueIntNotOneOf(1, nil), "value-not-allowed", "value is not one of the allowed values", "value 1 is not one of []"),
		Entry("is ErrorValueIntNotOneOf with empty array", structureValidator.ErrorValueIntNotOneOf(1, []int{}), "value-not-allowed", "value is not one of the allowed values", "value 1 is not one of []"),
		Entry("is ErrorValueIntNotOneOf with non-empty array", structureValidator.ErrorValueIntNotOneOf(1, []int{2, 3, 4}), "value-not-allowed", "value is not one of the allowed values", "value 1 is not one of [2, 3, 4]"),
		Entry("is ErrorValueStringOneOf with nil array", structureValidator.ErrorValueStringOneOf("abc", nil), "value-disallowed", "value is one of the disallowed values", `value "abc" is one of []`),
		Entry("is ErrorValueStringOneOf with empty array", structureValidator.ErrorValueStringOneOf("abc", []string{}), "value-disallowed", "value is one of the disallowed values", `value "abc" is one of []`),
		Entry("is ErrorValueStringOneOf with non-empty array", structureValidator.ErrorValueStringOneOf("abc", []string{"abc", "bcd", "cde"}), "value-disallowed", "value is one of the disallowed values", `value "abc" is one of ["abc", "bcd", "cde"]`),
		Entry("is ErrorValueStringNotOneOf with nil array", structureValidator.ErrorValueStringNotOneOf("xyz", nil), "value-not-allowed", "value is not one of the allowed values", `value "xyz" is not one of []`),
		Entry("is ErrorValueStringNotOneOf with empty array", structureValidator.ErrorValueStringNotOneOf("xyz", []string{}), "value-not-allowed", "value is not one of the allowed values", `value "xyz" is not one of []`),
		Entry("is ErrorValueStringNotOneOf with non-empty array", structureValidator.ErrorValueStringNotOneOf("xyz", []string{"abc", "bcd", "cde"}), "value-not-allowed", "value is not one of the allowed values", `value "xyz" is not one of ["abc", "bcd", "cde"]`),
		Entry("is ErrorValueStringMatches with nil expression", structureValidator.ErrorValueStringMatches("abc", nil), "value-matches", "value matches expression", `value "abc" matches expression "<MISSING>"`),
		Entry("is ErrorValueStringMatches with empty expression", structureValidator.ErrorValueStringMatches("abc", regexp.MustCompile("")), "value-matches", "value matches expression", `value "abc" matches expression ""`),
		Entry("is ErrorValueStringMatches with non-empty expression", structureValidator.ErrorValueStringMatches("abc", regexp.MustCompile("[a-z]*")), "value-matches", "value matches expression", `value "abc" matches expression "[a-z]*"`),
		Entry("is ErrorValueStringNotMatches with nil expression", structureValidator.ErrorValueStringNotMatches("abc", nil), "value-not-matches", "value does not match expression", `value "abc" does not match expression "<MISSING>"`),
		Entry("is ErrorValueStringNotMatches with empty expression", structureValidator.ErrorValueStringNotMatches("abc", regexp.MustCompile("")), "value-not-matches", "value does not match expression", `value "abc" does not match expression ""`),
		Entry("is ErrorValueStringNotMatches with non-empty expression", structureValidator.ErrorValueStringNotMatches("abc", regexp.MustCompile("[a-z]*")), "value-not-matches", "value does not match expression", `value "abc" does not match expression "[a-z]*"`),
		Entry("is ErrorValueStringAsTimeNotValid", structureValidator.ErrorValueStringAsTimeNotValid("abc", time.RFC3339), "value-not-valid", "value is not valid", `value "abc" is not valid as time with layout "2006-01-02T15:04:05Z07:00"`),
		Entry("is ErrorValueTimeNotAfter", structureValidator.ErrorValueTimeNotAfter(time.Unix(1451567655, 0).UTC(), time.Unix(1735737255, 0).UTC()), "value-not-after", "value is not after the specified time", `value "2015-12-31T13:14:15Z" is not after "2025-01-01T13:14:15Z"`),
		Entry("is ErrorValueTimeNotAfterNow", structureValidator.ErrorValueTimeNotAfterNow(time.Unix(1451567655, 0).UTC()), "value-not-after", "value is not after the specified time", `value "2015-12-31T13:14:15Z" is not after now`),
		Entry("is ErrorValueTimeNotBefore", structureValidator.ErrorValueTimeNotBefore(time.Unix(1735737255, 0).UTC(), time.Unix(1451567655, 0).UTC()), "value-not-before", "value is not before the specified time", `value "2025-01-01T13:14:15Z" is not before "2015-12-31T13:14:15Z"`),
		Entry("is ErrorValueTimeNotBeforeNow", structureValidator.ErrorValueTimeNotBeforeNow(time.Unix(1735737255, 0).UTC()), "value-not-before", "value is not before the specified time", `value "2025-01-01T13:14:15Z" is not before now`),
		Entry("is ErrorLengthNotEqualTo with int", structureValidator.ErrorLengthNotEqualTo(1, 2), "length-out-of-range", "length is out of range", "length 1 is not equal to 2"),
		Entry("is ErrorLengthEqualTo with int", structureValidator.ErrorLengthEqualTo(2, 2), "length-out-of-range", "length is out of range", "length 2 is equal to 2"),
		Entry("is ErrorLengthNotLessThan with int", structureValidator.ErrorLengthNotLessThan(2, 1), "length-out-of-range", "length is out of range", "length 2 is not less than 1"),
		Entry("is ErrorLengthNotLessThanOrEqualTo with int", structureValidator.ErrorLengthNotLessThanOrEqualTo(2, 1), "length-out-of-range", "length is out of range", "length 2 is not less than or equal to 1"),
		Entry("is ErrorLengthNotGreaterThan with int", structureValidator.ErrorLengthNotGreaterThan(1, 2), "length-out-of-range", "length is out of range", "length 1 is not greater than 2"),
		Entry("is ErrorLengthNotGreaterThanOrEqualTo with int", structureValidator.ErrorLengthNotGreaterThanOrEqualTo(1, 2), "length-out-of-range", "length is out of range", "length 1 is not greater than or equal to 2"),
		Entry("is ErrorLengthNotInRange", structureValidator.ErrorLengthNotInRange(1, 2, 3), "length-out-of-range", "length is out of range", "length 1 is not between 2 and 3"),
	)

	Context("QuoteIfString", func() {
		It("returns nil when the interface value is nil", func() {
			Expect(structureValidator.QuoteIfString(nil)).To(BeNil())
		})

		DescribeTable("returns expected value when",
			func(interfaceValue interface{}, expectedValue interface{}) {
				Expect(structureValidator.QuoteIfString(interfaceValue)).To(Equal(expectedValue))
			},
			Entry("is a string", "a string", `"a string"`),
			Entry("is an empty string", "", `""`),
			Entry("is an integer", 1, 1),
			Entry("is a float", 1.23, 1.23),
			Entry("is an array", []string{"a"}, []string{"a"}),
			Entry("is a map", map[string]string{"a": "b"}, map[string]string{"a": "b"}),
		)
	})

	Context("ExpressionAsString", func() {
		It("returns as expected if the expression is nil", func() {
			Expect(structureValidator.ExpressionAsString(nil)).To(Equal("<MISSING>"))
		})

		DescribeTable("returns expected value when",
			func(expressionString string, expectedString string) {
				Expect(structureValidator.ExpressionAsString(regexp.MustCompile(expressionString))).To(Equal(expectedString))
			},
			Entry("is an empty string", "", ""),
			Entry("is a simple string", "abc", "abc"),
			Entry("is a complex string", "[a-z]*.?.{32}", "[a-z]*.?.{32}"),
			Entry("contains single and double quotes", "'\"", "'\""),
		)
	})
})
