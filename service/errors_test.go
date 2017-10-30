package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Errors", func() {
	DescribeTable("all errors",
		func(err *service.Error, code string, title string, detail string, status int) {
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(code))
			Expect(err.Title).To(Equal(title))
			Expect(err.Detail).To(Equal(detail))
			Expect(err.Status).To(Equal(status))
			Expect(err.Source).To(BeNil())
			Expect(err.Meta).To(BeNil())
		},
		Entry("is ErrorInternalServerFailure", service.ErrorInternalServerFailure(), "internal-server-failure", "internal server failure", "Internal server failure", 500),
		Entry("is ErrorUnauthenticated", service.ErrorUnauthenticated(), "unauthenticated", "auth token is invalid", "Auth token is invalid", 401),
		Entry("is ErrorUnauthorized", service.ErrorUnauthorized(), "unauthorized", "auth token is not authorized for requested action", "Auth token is not authorized for requested action", 403),
		Entry("is ErrorResourceNotFound", service.ErrorResourceNotFound(), "resource-not-found", "resource not found", "Resource not found", 404),
		Entry("is ErrorResourceNotFoundWithID", service.ErrorResourceNotFoundWithID("test-id"), "resource-not-found", "resource not found", `Resource with id "test-id" not found`, 404),
		Entry("is ErrorParameterMissing", service.ErrorParameterMissing("test_parameter"), "parameter-missing", "parameter is missing", `parameter "test_parameter" is missing`, 403),
		Entry("is ErrorJSONMalformed", service.ErrorJSONMalformed(), "json-malformed", "json is malformed", "JSON is malformed", 400),
		Entry("is ErrorTypeNotBoolean with nil parameter", service.ErrorTypeNotBoolean(nil), "type-not-boolean", "type is not boolean", "Type is not boolean, but <nil>", 0),
		Entry("is ErrorTypeNotBoolean with int parameter", service.ErrorTypeNotBoolean(-1), "type-not-boolean", "type is not boolean", "Type is not boolean, but int", 0),
		Entry("is ErrorTypeNotBoolean with string parameter", service.ErrorTypeNotBoolean("test"), "type-not-boolean", "type is not boolean", "Type is not boolean, but string", 0),
		Entry("is ErrorTypeNotBoolean with string array parameter", service.ErrorTypeNotBoolean([]string{}), "type-not-boolean", "type is not boolean", "Type is not boolean, but []string", 0),
		Entry("is ErrorTypeNotInteger with nil parameter", service.ErrorTypeNotInteger(nil), "type-not-integer", "type is not integer", "Type is not integer, but <nil>", 0),
		Entry("is ErrorTypeNotInteger with bool parameter", service.ErrorTypeNotInteger(true), "type-not-integer", "type is not integer", "Type is not integer, but bool", 0),
		Entry("is ErrorTypeNotInteger with string parameter", service.ErrorTypeNotInteger("test"), "type-not-integer", "type is not integer", "Type is not integer, but string", 0),
		Entry("is ErrorTypeNotInteger with string array parameter", service.ErrorTypeNotInteger([]string{}), "type-not-integer", "type is not integer", "Type is not integer, but []string", 0),
		Entry("is ErrorTypeNotUnsignedInteger with nil parameter", service.ErrorTypeNotUnsignedInteger(nil), "type-not-unsigned-integer", "type is not unsigned integer", "Type is not unsigned integer, but <nil>", 0),
		Entry("is ErrorTypeNotUnsignedInteger with int parameter", service.ErrorTypeNotUnsignedInteger(-1), "type-not-unsigned-integer", "type is not unsigned integer", "Type is not unsigned integer, but int", 0),
		Entry("is ErrorTypeNotUnsignedInteger with string parameter", service.ErrorTypeNotUnsignedInteger("test"), "type-not-unsigned-integer", "type is not unsigned integer", "Type is not unsigned integer, but string", 0),
		Entry("is ErrorTypeNotUnsignedInteger with string array parameter", service.ErrorTypeNotUnsignedInteger([]string{}), "type-not-unsigned-integer", "type is not unsigned integer", "Type is not unsigned integer, but []string", 0),
		Entry("is ErrorTypeNotFloat with nil parameter", service.ErrorTypeNotFloat(nil), "type-not-float", "type is not float", "Type is not float, but <nil>", 0),
		Entry("is ErrorTypeNotFloat with int parameter", service.ErrorTypeNotFloat(-1), "type-not-float", "type is not float", "Type is not float, but int", 0),
		Entry("is ErrorTypeNotFloat with string parameter", service.ErrorTypeNotFloat("test"), "type-not-float", "type is not float", "Type is not float, but string", 0),
		Entry("is ErrorTypeNotFloat with string array parameter", service.ErrorTypeNotFloat([]string{}), "type-not-float", "type is not float", "Type is not float, but []string", 0),
		Entry("is ErrorTypeNotString with nil parameter", service.ErrorTypeNotString(nil), "type-not-string", "type is not string", "Type is not string, but <nil>", 0),
		Entry("is ErrorTypeNotString with int parameter", service.ErrorTypeNotString(-1), "type-not-string", "type is not string", "Type is not string, but int", 0),
		Entry("is ErrorTypeNotString with string parameter", service.ErrorTypeNotString("test"), "type-not-string", "type is not string", "Type is not string, but string", 0),
		Entry("is ErrorTypeNotString with string array parameter", service.ErrorTypeNotString([]string{}), "type-not-string", "type is not string", "Type is not string, but []string", 0),
		Entry("is ErrorTypeNotObject with nil parameter", service.ErrorTypeNotObject(nil), "type-not-object", "type is not object", "Type is not object, but <nil>", 0),
		Entry("is ErrorTypeNotObject with int parameter", service.ErrorTypeNotObject(-1), "type-not-object", "type is not object", "Type is not object, but int", 0),
		Entry("is ErrorTypeNotObject with string parameter", service.ErrorTypeNotObject("test"), "type-not-object", "type is not object", "Type is not object, but string", 0),
		Entry("is ErrorTypeNotObject with string array parameter", service.ErrorTypeNotObject([]string{}), "type-not-object", "type is not object", "Type is not object, but []string", 0),
		Entry("is ErrorTypeNotArray with nil parameter", service.ErrorTypeNotArray(nil), "type-not-array", "type is not array", "Type is not array, but <nil>", 0),
		Entry("is ErrorTypeNotArray with int parameter", service.ErrorTypeNotArray(-1), "type-not-array", "type is not array", "Type is not array, but int", 0),
		Entry("is ErrorTypeNotArray with string parameter", service.ErrorTypeNotArray("test"), "type-not-array", "type is not array", "Type is not array, but string", 0),
		Entry("is ErrorTypeNotArray with string array parameter", service.ErrorTypeNotArray([]string{}), "type-not-array", "type is not array", "Type is not array, but []string", 0),
		Entry("is ErrorValueNotExists", service.ErrorValueNotExists(), "value-not-exists", "value does not exist", "Value does not exist", 0),
		Entry("is ErrorValueExists", service.ErrorValueExists(), "value-exists", "value exists", "Value exists", 0),
		Entry("is ErrorValueNotEmpty", service.ErrorValueNotEmpty(), "value-not-empty", "value is not empty", "Value is not empty", 0),
		Entry("is ErrorValueEmpty", service.ErrorValueEmpty(), "value-empty", "value is empty", "Value is empty", 0),
		Entry("is ErrorValueNotTrue", service.ErrorValueNotTrue(), "value-not-true", "value is not true", "Value is not true", 0),
		Entry("is ErrorValueNotFalse", service.ErrorValueNotFalse(), "value-not-false", "value is not false", "Value is not false", 0),
		Entry("is ErrorValueNotEqualTo with int", service.ErrorValueNotEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not equal to 2", 0),
		Entry("is ErrorValueNotEqualTo with float", service.ErrorValueNotEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not equal to 5.6", 0),
		Entry("is ErrorValueNotEqualTo with string", service.ErrorValueNotEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", `Value "abc" is not equal to "xyz"`, 0),
		Entry("is ErrorValueNotEqualTo with string with quotes", service.ErrorValueNotEqualTo(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "a\"b\"c" is not equal to "x\"y\"z"`, 0),
		Entry("is ErrorValueEqualTo with int", service.ErrorValueEqualTo(2, 2), "value-out-of-range", "value is out of range", "Value 2 is equal to 2", 0),
		Entry("is ErrorValueEqualTo with float", service.ErrorValueEqualTo(5.6, 5.6), "value-out-of-range", "value is out of range", "Value 5.6 is equal to 5.6", 0),
		Entry("is ErrorValueEqualTo with string", service.ErrorValueEqualTo("xyz", "xyz"), "value-out-of-range", "value is out of range", `Value "xyz" is equal to "xyz"`, 0),
		Entry("is ErrorValueEqualTo with string with quotes", service.ErrorValueEqualTo(`x"y"z`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "x\"y\"z" is equal to "x\"y\"z"`, 0),
		Entry("is ErrorValueNotLessThan with int", service.ErrorValueNotLessThan(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than 1", 0),
		Entry("is ErrorValueNotLessThan with float", service.ErrorValueNotLessThan(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than 3.4", 0),
		Entry("is ErrorValueNotLessThan with string", service.ErrorValueNotLessThan("xyz", "abc"), "value-out-of-range", "value is out of range", `Value "xyz" is not less than "abc"`, 0),
		Entry("is ErrorValueNotLessThan with string with quotes", service.ErrorValueNotLessThan(`x"y"z`, `a"b"c`), "value-out-of-range", "value is out of range", `Value "x\"y\"z" is not less than "a\"b\"c"`, 0),
		Entry("is ErrorValueNotLessThanOrEqualTo with int", service.ErrorValueNotLessThanOrEqualTo(2, 1), "value-out-of-range", "value is out of range", "Value 2 is not less than or equal to 1", 0),
		Entry("is ErrorValueNotLessThanOrEqualTo with float", service.ErrorValueNotLessThanOrEqualTo(5.6, 3.4), "value-out-of-range", "value is out of range", "Value 5.6 is not less than or equal to 3.4", 0),
		Entry("is ErrorValueNotLessThanOrEqualTo with string", service.ErrorValueNotLessThanOrEqualTo("xyz", "abc"), "value-out-of-range", "value is out of range", `Value "xyz" is not less than or equal to "abc"`, 0),
		Entry("is ErrorValueNotLessThanOrEqualTo with string with quotes", service.ErrorValueNotLessThanOrEqualTo(`x"y"z`, `a"b"c`), "value-out-of-range", "value is out of range", `Value "x\"y\"z" is not less than or equal to "a\"b\"c"`, 0),
		Entry("is ErrorValueNotGreaterThan with int", service.ErrorValueNotGreaterThan(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than 2", 0),
		Entry("is ErrorValueNotGreaterThan with float", service.ErrorValueNotGreaterThan(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than 5.6", 0),
		Entry("is ErrorValueNotGreaterThan with string", service.ErrorValueNotGreaterThan("abc", "xyz"), "value-out-of-range", "value is out of range", `Value "abc" is not greater than "xyz"`, 0),
		Entry("is ErrorValueNotGreaterThan with string with quotes", service.ErrorValueNotGreaterThan(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "a\"b\"c" is not greater than "x\"y\"z"`, 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with int", service.ErrorValueNotGreaterThanOrEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than or equal to 2", 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with float", service.ErrorValueNotGreaterThanOrEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than or equal to 5.6", 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string", service.ErrorValueNotGreaterThanOrEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", `Value "abc" is not greater than or equal to "xyz"`, 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string with quotes", service.ErrorValueNotGreaterThanOrEqualTo(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "a\"b\"c" is not greater than or equal to "x\"y\"z"`, 0),
		Entry("is ErrorValueNotInRange with int", service.ErrorValueNotInRange(1, 2, 3), "value-out-of-range", "value is out of range", "Value 1 is not between 2 and 3", 0),
		Entry("is ErrorValueNotInRange with float", service.ErrorValueNotInRange(1.4, 2.4, 3.4), "value-out-of-range", "value is out of range", "Value 1.4 is not between 2.4 and 3.4", 0),
		Entry("is ErrorValueNotInRange with string", service.ErrorValueNotInRange("zzz", "abc", "xyz"), "value-out-of-range", "value is out of range", `Value "zzz" is not between "abc" and "xyz"`, 0),
		Entry("is ErrorValueNotInRange with string with quotes", service.ErrorValueNotInRange(`z"z"z`, `a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "z\"z\"z" is not between "a\"b\"c" and "x\"y\"z"`, 0),
		Entry("is ErrorValueIntegerOneOf with nil array", service.ErrorValueIntegerOneOf(2, nil), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []", 0),
		Entry("is ErrorValueIntegerOneOf with empty array", service.ErrorValueIntegerOneOf(2, []int{}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of []", 0),
		Entry("is ErrorValueIntegerOneOf with non-empty array", service.ErrorValueIntegerOneOf(2, []int{2, 3, 4}), "value-disallowed", "value is one of the disallowed values", "Value 2 is one of [2, 3, 4]", 0),
		Entry("is ErrorValueIntegerNotOneOf with nil array", service.ErrorValueIntegerNotOneOf(1, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []", 0),
		Entry("is ErrorValueIntegerNotOneOf with empty array", service.ErrorValueIntegerNotOneOf(1, []int{}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of []", 0),
		Entry("is ErrorValueIntegerNotOneOf with non-empty array", service.ErrorValueIntegerNotOneOf(1, []int{2, 3, 4}), "value-not-allowed", "value is not one of the allowed values", "Value 1 is not one of [2, 3, 4]", 0),
		Entry("is ErrorValueFloatOneOf with nil array", service.ErrorValueFloatOneOf(2.5, nil), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []", 0),
		Entry("is ErrorValueFloatOneOf with empty array", service.ErrorValueFloatOneOf(2.5, []float64{}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of []", 0),
		Entry("is ErrorValueFloatOneOf with non-empty array", service.ErrorValueFloatOneOf(2.5, []float64{2.5, 3.5, 4.5}), "value-disallowed", "value is one of the disallowed values", "Value 2.5 is one of [2.5, 3.5, 4.5]", 0),
		Entry("is ErrorValueFloatNotOneOf with nil array", service.ErrorValueFloatNotOneOf(1.5, nil), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []", 0),
		Entry("is ErrorValueFloatNotOneOf with empty array", service.ErrorValueFloatNotOneOf(1.5, []float64{}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of []", 0),
		Entry("is ErrorValueFloatNotOneOf with non-empty array", service.ErrorValueFloatNotOneOf(1.5, []float64{2.5, 3.5, 4.5}), "value-not-allowed", "value is not one of the allowed values", "Value 1.5 is not one of [2.5, 3.5, 4.5]", 0),
		Entry("is ErrorValueStringOneOf with nil array", service.ErrorValueStringOneOf("abc", nil), "value-disallowed", "value is one of the disallowed values", `Value "abc" is one of []`, 0),
		Entry("is ErrorValueStringOneOf with empty array", service.ErrorValueStringOneOf("abc", []string{}), "value-disallowed", "value is one of the disallowed values", `Value "abc" is one of []`, 0),
		Entry("is ErrorValueStringOneOf with non-empty array", service.ErrorValueStringOneOf("abc", []string{"abc", "bcd", "cde"}), "value-disallowed", "value is one of the disallowed values", `Value "abc" is one of ["abc", "bcd", "cde"]`, 0),
		Entry("is ErrorValueStringNotOneOf with nil array", service.ErrorValueStringNotOneOf("xyz", nil), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of []`, 0),
		Entry("is ErrorValueStringNotOneOf with empty array", service.ErrorValueStringNotOneOf("xyz", []string{}), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of []`, 0),
		Entry("is ErrorValueStringNotOneOf with non-empty array", service.ErrorValueStringNotOneOf("xyz", []string{"abc", "bcd", "cde"}), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of ["abc", "bcd", "cde"]`, 0),
		Entry("is ErrorValueTimeNotValid", service.ErrorValueTimeNotValid("abc", "2006-01-02T15:04:05Z07:00"), "value-not-valid", "value is not a valid time", `Value "abc" is not a valid time of format "2006-01-02T15:04:05Z07:00"`, 0),
		Entry("is ErrorValueTimeNotAfter", service.ErrorValueTimeNotAfter(time.Unix(1451567655, 0).UTC(), time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "value-not-after", "value is not after the specified time", `Value "2015-12-31T13:14:15Z" is not after "2025-01-01T13:14:15Z"`, 0),
		Entry("is ErrorValueTimeNotAfterNow", service.ErrorValueTimeNotAfterNow(time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "value-not-after", "value is not after the specified time", `Value "2015-12-31T13:14:15Z" is not after now`, 0),
		Entry("is ErrorValueTimeNotBefore", service.ErrorValueTimeNotBefore(time.Unix(1735737255, 0).UTC(), time.Unix(1451567655, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "value-not-before", "value is not before the specified time", `Value "2025-01-01T13:14:15Z" is not before "2015-12-31T13:14:15Z"`, 0),
		Entry("is ErrorValueTimeNotBeforeNow", service.ErrorValueTimeNotBeforeNow(time.Unix(1735737255, 0).UTC(), "2006-01-02T15:04:05Z07:00"), "value-not-before", "value is not before the specified time", `Value "2025-01-01T13:14:15Z" is not before now`, 0),
		Entry("is ErrorLengthNotEqualTo with int", service.ErrorLengthNotEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not equal to 2", 0),
		Entry("is ErrorLengthEqualTo with int", service.ErrorLengthEqualTo(2, 2), "length-out-of-range", "length is out of range", "Length 2 is equal to 2", 0),
		Entry("is ErrorLengthNotLessThan with int", service.ErrorLengthNotLessThan(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than 1", 0),
		Entry("is ErrorLengthNotLessThanOrEqualTo with int", service.ErrorLengthNotLessThanOrEqualTo(2, 1), "length-out-of-range", "length is out of range", "Length 2 is not less than or equal to 1", 0),
		Entry("is ErrorLengthNotGreaterThan with int", service.ErrorLengthNotGreaterThan(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than 2", 0),
		Entry("is ErrorLengthNotGreaterThanOrEqualTo with int", service.ErrorLengthNotGreaterThanOrEqualTo(1, 2), "length-out-of-range", "length is out of range", "Length 1 is not greater than or equal to 2", 0),
		Entry("is ErrorLengthNotInRange", service.ErrorLengthNotInRange(1, 2, 3), "length-out-of-range", "length is out of range", "Length 1 is not between 2 and 3", 0),
	)

	Context("QuoteIfString", func() {
		It("returns nil when the interface value is nil", func() {
			Expect(service.QuoteIfString(nil)).To(BeNil())
		})

		DescribeTable("returns expected value when",
			func(interfaceValue interface{}, expectedValue interface{}) {
				Expect(service.QuoteIfString(interfaceValue)).To(Equal(expectedValue))
			},
			Entry("is a string", "a string", `"a string"`),
			Entry("is an empty string", "", `""`),
			Entry("is an error", errors.New("error"), errors.New("error")),
			Entry("is an integer", 1, 1),
			Entry("is a float", 1.23, 1.23),
			Entry("is an array", []string{"a"}, []string{"a"}),
			Entry("is a map", map[string]string{"a": "b"}, map[string]string{"a": "b"}),
		)
	})
})
