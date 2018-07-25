package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	structureParser "github.com/tidepool-org/platform/structure/parser"
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
		Entry("is ErrorTypeNotBool with nil parameter", structureParser.ErrorTypeNotBool(nil), "type-not-bool", "type is not bool", "type is not bool, but <nil>"),
		Entry("is ErrorTypeNotBool with int parameter", structureParser.ErrorTypeNotBool(-1), "type-not-bool", "type is not bool", "type is not bool, but int"),
		Entry("is ErrorTypeNotBool with string parameter", structureParser.ErrorTypeNotBool("test"), "type-not-bool", "type is not bool", "type is not bool, but string"),
		Entry("is ErrorTypeNotBool with string array parameter", structureParser.ErrorTypeNotBool([]string{}), "type-not-bool", "type is not bool", "type is not bool, but []string"),
		Entry("is ErrorTypeNotFloat64 with nil parameter", structureParser.ErrorTypeNotFloat64(nil), "type-not-float64", "type is not float64", "type is not float64, but <nil>"),
		Entry("is ErrorTypeNotFloat64 with int parameter", structureParser.ErrorTypeNotFloat64(-1), "type-not-float64", "type is not float64", "type is not float64, but int"),
		Entry("is ErrorTypeNotFloat64 with string parameter", structureParser.ErrorTypeNotFloat64("test"), "type-not-float64", "type is not float64", "type is not float64, but string"),
		Entry("is ErrorTypeNotFloat64 with string array parameter", structureParser.ErrorTypeNotFloat64([]string{}), "type-not-float64", "type is not float64", "type is not float64, but []string"),
		Entry("is ErrorTypeNotInt with nil parameter", structureParser.ErrorTypeNotInt(nil), "type-not-int", "type is not int", "type is not int, but <nil>"),
		Entry("is ErrorTypeNotInt with bool parameter", structureParser.ErrorTypeNotInt(true), "type-not-int", "type is not int", "type is not int, but bool"),
		Entry("is ErrorTypeNotInt with string parameter", structureParser.ErrorTypeNotInt("test"), "type-not-int", "type is not int", "type is not int, but string"),
		Entry("is ErrorTypeNotInt with string array parameter", structureParser.ErrorTypeNotInt([]string{}), "type-not-int", "type is not int", "type is not int, but []string"),
		Entry("is ErrorTypeNotString with nil parameter", structureParser.ErrorTypeNotString(nil), "type-not-string", "type is not string", "type is not string, but <nil>"),
		Entry("is ErrorTypeNotString with int parameter", structureParser.ErrorTypeNotString(-1), "type-not-string", "type is not string", "type is not string, but int"),
		Entry("is ErrorTypeNotString with string parameter", structureParser.ErrorTypeNotString("test"), "type-not-string", "type is not string", "type is not string, but string"),
		Entry("is ErrorTypeNotString with string array parameter", structureParser.ErrorTypeNotString([]string{}), "type-not-string", "type is not string", "type is not string, but []string"),
		Entry("is ErrorTypeNotTime with nil parameter", structureParser.ErrorTypeNotTime(nil), "type-not-time", "type is not time", "type is not time, but <nil>"),
		Entry("is ErrorTypeNotTime with int parameter", structureParser.ErrorTypeNotTime(-1), "type-not-time", "type is not time", "type is not time, but int"),
		Entry("is ErrorTypeNotTime with string parameter", structureParser.ErrorTypeNotTime("test"), "type-not-time", "type is not time", "type is not time, but string"),
		Entry("is ErrorTypeNotTime with string array parameter", structureParser.ErrorTypeNotTime([]string{}), "type-not-time", "type is not time", "type is not time, but []string"),
		Entry("is ErrorTypeNotObject with nil parameter", structureParser.ErrorTypeNotObject(nil), "type-not-object", "type is not object", "type is not object, but <nil>"),
		Entry("is ErrorTypeNotObject with int parameter", structureParser.ErrorTypeNotObject(-1), "type-not-object", "type is not object", "type is not object, but int"),
		Entry("is ErrorTypeNotObject with string parameter", structureParser.ErrorTypeNotObject("test"), "type-not-object", "type is not object", "type is not object, but string"),
		Entry("is ErrorTypeNotObject with string array parameter", structureParser.ErrorTypeNotObject([]string{}), "type-not-object", "type is not object", "type is not object, but []string"),
		Entry("is ErrorTypeNotArray with nil parameter", structureParser.ErrorTypeNotArray(nil), "type-not-array", "type is not array", "type is not array, but <nil>"),
		Entry("is ErrorTypeNotArray with int parameter", structureParser.ErrorTypeNotArray(-1), "type-not-array", "type is not array", "type is not array, but int"),
		Entry("is ErrorTypeNotArray with string parameter", structureParser.ErrorTypeNotArray("test"), "type-not-array", "type is not array", "type is not array, but string"),
		Entry("is ErrorTypeNotArray with string array parameter", structureParser.ErrorTypeNotArray([]string{}), "type-not-array", "type is not array", "type is not array, but []string"),
		Entry("is ErrorValueTimeNotParsable", structureParser.ErrorValueTimeNotParsable("abc", time.RFC3339), "value-not-parsable", "value is not a parsable time", `value "abc" is not a parsable time of format "2006-01-02T15:04:05Z07:00"`),
		Entry("is ErrorNotParsed", structureParser.ErrorNotParsed(), "not-parsed", "not parsed", "not parsed"),
	)
})
