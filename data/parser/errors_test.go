package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/parser"
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
		Entry("is ErrorTypeNotBoolean with nil parameter", parser.ErrorTypeNotBoolean(nil), "type-not-boolean", "type is not boolean", "Type is not boolean, but <nil>"),
		Entry("is ErrorTypeNotBoolean with int parameter", parser.ErrorTypeNotBoolean(0), "type-not-boolean", "type is not boolean", "Type is not boolean, but int"),
		Entry("is ErrorTypeNotBoolean with string parameter", parser.ErrorTypeNotBoolean("test"), "type-not-boolean", "type is not boolean", "Type is not boolean, but string"),
		Entry("is ErrorTypeNotBoolean with string array parameter", parser.ErrorTypeNotBoolean([]string{}), "type-not-boolean", "type is not boolean", "Type is not boolean, but []string"),
		Entry("is ErrorTypeNotInteger with nil parameter", parser.ErrorTypeNotInteger(nil), "type-not-integer", "type is not integer", "Type is not integer, but <nil>"),
		Entry("is ErrorTypeNotInteger with int parameter", parser.ErrorTypeNotInteger(0), "type-not-integer", "type is not integer", "Type is not integer, but int"),
		Entry("is ErrorTypeNotInteger with string parameter", parser.ErrorTypeNotInteger("test"), "type-not-integer", "type is not integer", "Type is not integer, but string"),
		Entry("is ErrorTypeNotInteger with string array parameter", parser.ErrorTypeNotInteger([]string{}), "type-not-integer", "type is not integer", "Type is not integer, but []string"),
		Entry("is ErrorTypeNotFloat with nil parameter", parser.ErrorTypeNotFloat(nil), "type-not-float", "type is not float", "Type is not float, but <nil>"),
		Entry("is ErrorTypeNotFloat with int parameter", parser.ErrorTypeNotFloat(0), "type-not-float", "type is not float", "Type is not float, but int"),
		Entry("is ErrorTypeNotFloat with string parameter", parser.ErrorTypeNotFloat("test"), "type-not-float", "type is not float", "Type is not float, but string"),
		Entry("is ErrorTypeNotFloat with string array parameter", parser.ErrorTypeNotFloat([]string{}), "type-not-float", "type is not float", "Type is not float, but []string"),
		Entry("is ErrorTypeNotString with nil parameter", parser.ErrorTypeNotString(nil), "type-not-string", "type is not string", "Type is not string, but <nil>"),
		Entry("is ErrorTypeNotString with int parameter", parser.ErrorTypeNotString(0), "type-not-string", "type is not string", "Type is not string, but int"),
		Entry("is ErrorTypeNotString with string parameter", parser.ErrorTypeNotString("test"), "type-not-string", "type is not string", "Type is not string, but string"),
		Entry("is ErrorTypeNotString with string array parameter", parser.ErrorTypeNotString([]string{}), "type-not-string", "type is not string", "Type is not string, but []string"),
		Entry("is ErrorTypeNotObject with nil parameter", parser.ErrorTypeNotObject(nil), "type-not-object", "type is not object", "Type is not object, but <nil>"),
		Entry("is ErrorTypeNotObject with int parameter", parser.ErrorTypeNotObject(0), "type-not-object", "type is not object", "Type is not object, but int"),
		Entry("is ErrorTypeNotObject with string parameter", parser.ErrorTypeNotObject("test"), "type-not-object", "type is not object", "Type is not object, but string"),
		Entry("is ErrorTypeNotObject with string array parameter", parser.ErrorTypeNotObject([]string{}), "type-not-object", "type is not object", "Type is not object, but []string"),
		Entry("is ErrorTypeNotArray with nil parameter", parser.ErrorTypeNotArray(nil), "type-not-array", "type is not array", "Type is not array, but <nil>"),
		Entry("is ErrorTypeNotArray with int parameter", parser.ErrorTypeNotArray(0), "type-not-array", "type is not array", "Type is not array, but int"),
		Entry("is ErrorTypeNotArray with string parameter", parser.ErrorTypeNotArray("test"), "type-not-array", "type is not array", "Type is not array, but string"),
		Entry("is ErrorTypeNotArray with string array parameter", parser.ErrorTypeNotArray([]string{}), "type-not-array", "type is not array", "Type is not array, but []string"),
	)
})
