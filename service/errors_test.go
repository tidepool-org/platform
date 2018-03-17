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
		Entry("is ErrorJSONMalformed", service.ErrorJSONMalformed(), "json-malformed", "json is malformed", "JSON is malformed", 400),
		Entry("is ErrorTypeNotBoolean with nil parameter", service.ErrorTypeNotBoolean(nil), "type-not-boolean", "type is not boolean", "Type is not boolean, but <nil>", 0),
		Entry("is ErrorTypeNotBoolean with int parameter", service.ErrorTypeNotBoolean(-1), "type-not-boolean", "type is not boolean", "Type is not boolean, but int", 0),
		Entry("is ErrorTypeNotBoolean with string parameter", service.ErrorTypeNotBoolean("test"), "type-not-boolean", "type is not boolean", "Type is not boolean, but string", 0),
		Entry("is ErrorTypeNotBoolean with string array parameter", service.ErrorTypeNotBoolean([]string{}), "type-not-boolean", "type is not boolean", "Type is not boolean, but []string", 0),
		Entry("is ErrorTypeNotInteger with nil parameter", service.ErrorTypeNotInteger(nil), "type-not-integer", "type is not integer", "Type is not integer, but <nil>", 0),
		Entry("is ErrorTypeNotInteger with bool parameter", service.ErrorTypeNotInteger(true), "type-not-integer", "type is not integer", "Type is not integer, but bool", 0),
		Entry("is ErrorTypeNotInteger with string parameter", service.ErrorTypeNotInteger("test"), "type-not-integer", "type is not integer", "Type is not integer, but string", 0),
		Entry("is ErrorTypeNotInteger with string array parameter", service.ErrorTypeNotInteger([]string{}), "type-not-integer", "type is not integer", "Type is not integer, but []string", 0),
		Entry("is ErrorTypeNotFloat with nil parameter", service.ErrorTypeNotFloat(nil), "type-not-float", "type is not float", "Type is not float, but <nil>", 0),
		Entry("is ErrorTypeNotFloat with int parameter", service.ErrorTypeNotFloat(-1), "type-not-float", "type is not float", "Type is not float, but int", 0),
		Entry("is ErrorTypeNotFloat with string parameter", service.ErrorTypeNotFloat("test"), "type-not-float", "type is not float", "Type is not float, but string", 0),
		Entry("is ErrorTypeNotFloat with string array parameter", service.ErrorTypeNotFloat([]string{}), "type-not-float", "type is not float", "Type is not float, but []string", 0),
		Entry("is ErrorTypeNotString with nil parameter", service.ErrorTypeNotString(nil), "type-not-string", "type is not string", "Type is not string, but <nil>", 0),
		Entry("is ErrorTypeNotString with int parameter", service.ErrorTypeNotString(-1), "type-not-string", "type is not string", "Type is not string, but int", 0),
		Entry("is ErrorTypeNotString with string parameter", service.ErrorTypeNotString("test"), "type-not-string", "type is not string", "Type is not string, but string", 0),
		Entry("is ErrorTypeNotString with string array parameter", service.ErrorTypeNotString([]string{}), "type-not-string", "type is not string", "Type is not string, but []string", 0),
		Entry("is ErrorTypeNotTime with nil parameter", service.ErrorTypeNotTime(nil), "type-not-time", "type is not time", "Type is not time, but <nil>", 0),
		Entry("is ErrorTypeNotTime with int parameter", service.ErrorTypeNotTime(-1), "type-not-time", "type is not time", "Type is not time, but int", 0),
		Entry("is ErrorTypeNotTime with string parameter", service.ErrorTypeNotTime("test"), "type-not-time", "type is not time", "Type is not time, but string", 0),
		Entry("is ErrorTypeNotTime with string array parameter", service.ErrorTypeNotTime([]string{}), "type-not-time", "type is not time", "Type is not time, but []string", 0),
		Entry("is ErrorTypeNotObject with nil parameter", service.ErrorTypeNotObject(nil), "type-not-object", "type is not object", "Type is not object, but <nil>", 0),
		Entry("is ErrorTypeNotObject with int parameter", service.ErrorTypeNotObject(-1), "type-not-object", "type is not object", "Type is not object, but int", 0),
		Entry("is ErrorTypeNotObject with string parameter", service.ErrorTypeNotObject("test"), "type-not-object", "type is not object", "Type is not object, but string", 0),
		Entry("is ErrorTypeNotObject with string array parameter", service.ErrorTypeNotObject([]string{}), "type-not-object", "type is not object", "Type is not object, but []string", 0),
		Entry("is ErrorTypeNotArray with nil parameter", service.ErrorTypeNotArray(nil), "type-not-array", "type is not array", "Type is not array, but <nil>", 0),
		Entry("is ErrorTypeNotArray with int parameter", service.ErrorTypeNotArray(-1), "type-not-array", "type is not array", "Type is not array, but int", 0),
		Entry("is ErrorTypeNotArray with string parameter", service.ErrorTypeNotArray("test"), "type-not-array", "type is not array", "Type is not array, but string", 0),
		Entry("is ErrorTypeNotArray with string array parameter", service.ErrorTypeNotArray([]string{}), "type-not-array", "type is not array", "Type is not array, but []string", 0),
		Entry("is ErrorValueTimeNotParsable", service.ErrorValueTimeNotParsable("abc", time.RFC3339), "value-not-parsable", "value is not a parsable time", `value "abc" is not a parsable time of format "2006-01-02T15:04:05Z07:00"`, 0),
		Entry("is ErrorValueNotExists", service.ErrorValueNotExists(), "value-not-exists", "value does not exist", "Value does not exist", 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with int", service.ErrorValueNotGreaterThanOrEqualTo(1, 2), "value-out-of-range", "value is out of range", "Value 1 is not greater than or equal to 2", 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with float", service.ErrorValueNotGreaterThanOrEqualTo(3.4, 5.6), "value-out-of-range", "value is out of range", "Value 3.4 is not greater than or equal to 5.6", 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string", service.ErrorValueNotGreaterThanOrEqualTo("abc", "xyz"), "value-out-of-range", "value is out of range", `Value "abc" is not greater than or equal to "xyz"`, 0),
		Entry("is ErrorValueNotGreaterThanOrEqualTo with string with quotes", service.ErrorValueNotGreaterThanOrEqualTo(`a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "a\"b\"c" is not greater than or equal to "x\"y\"z"`, 0),
		Entry("is ErrorValueNotInRange with int", service.ErrorValueNotInRange(1, 2, 3), "value-out-of-range", "value is out of range", "Value 1 is not between 2 and 3", 0),
		Entry("is ErrorValueNotInRange with float", service.ErrorValueNotInRange(1.4, 2.4, 3.4), "value-out-of-range", "value is out of range", "Value 1.4 is not between 2.4 and 3.4", 0),
		Entry("is ErrorValueNotInRange with string", service.ErrorValueNotInRange("zzz", "abc", "xyz"), "value-out-of-range", "value is out of range", `Value "zzz" is not between "abc" and "xyz"`, 0),
		Entry("is ErrorValueNotInRange with string with quotes", service.ErrorValueNotInRange(`z"z"z`, `a"b"c`, `x"y"z`), "value-out-of-range", "value is out of range", `Value "z\"z\"z" is not between "a\"b\"c" and "x\"y\"z"`, 0),
		Entry("is ErrorValueStringNotOneOf with nil array", service.ErrorValueStringNotOneOf("xyz", nil), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of []`, 0),
		Entry("is ErrorValueStringNotOneOf with empty array", service.ErrorValueStringNotOneOf("xyz", []string{}), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of []`, 0),
		Entry("is ErrorValueStringNotOneOf with non-empty array", service.ErrorValueStringNotOneOf("xyz", []string{"abc", "bcd", "cde"}), "value-not-allowed", "value is not one of the allowed values", `Value "xyz" is not one of ["abc", "bcd", "cde"]`, 0),
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
