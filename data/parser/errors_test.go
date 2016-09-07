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
		func(err *service.Error, code string, title string, detail string, status int) {
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(code))
			Expect(err.Title).To(Equal(title))
			Expect(err.Detail).To(Equal(detail))
			Expect(err.Status).To(Equal(status))
			Expect(err.Source).To(BeNil())
			Expect(err.Meta).To(BeNil())
		},
		Entry("is ErrorNotParsed", parser.ErrorNotParsed(), "not-parsed", "not parsed", "Not parsed", 0),
	)
})
