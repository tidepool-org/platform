package request_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	structureTest "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("ValuesParser", func() {
	var logger *logTest.Logger

	BeforeEach(func() {
		logger = logTest.NewLogger()
	})

	Context("NewValues", func() {
		It("returns successfully", func() {
			values := map[string][]string{"one": {"two"}}
			Expect(request.NewValues(logger, &values)).ToNot(BeNil())
		})

		It("returns successfully with nil values", func() {
			Expect(request.NewValues(logger, nil)).ToNot(BeNil())
		})
	})

	Context("with nil values", func() {
		var parser *request.Values

		BeforeEach(func() {
			parser = request.NewValues(logger, nil)
			Expect(parser).ToNot(BeNil())
		})

		It("IgnoreNotParsed only returns existing errors", func() {
			err := errorsTest.RandomError()
			parser.ReportError(err)
			parser.IgnoreNotParsed()
			Expect(parser.Error()).To(Equal(errors.Normalize(err)))
		})

		It("ReportNotParsed only returns existing errors", func() {
			err := errorsTest.RandomError()
			parser.ReportError(err)
			parser.ReportNotParsed()
			Expect(parser.Error()).To(Equal(errors.Normalize(err)))
		})
	})

	Context("IgnoreNotParsed", func() {
		var parser *request.Values

		BeforeEach(func() {
			values := map[string][]string{
				"zero": {"1"},
				"one":  {"two"},
				"two":  {"3", "4"},
			}
			parser = request.NewValues(logger, &values)
			Expect(parser).ToNot(BeNil())
		})

		It("without anything parsed marks every value parsed", func() {
			parser.IgnoreNotParsed()
			Expect(parser.NotParsed()).To(BeNil())
			parser.ReportNotParsed()
			Expect(parser.Error()).ToNot(HaveOccurred())
		})

		It("with some values parsed marks the remaining values parsed", func() {
			parser.String("one")
			parser.IgnoreNotParsed()
			Expect(parser.NotParsed()).To(BeNil())
			parser.ReportNotParsed()
			Expect(parser.Error()).ToNot(HaveOccurred())
		})

		It("marks the remainder of a partially consumed multi-value parsed", func() {
			parser.String("two")
			Expect(parser.NotParsed()).To(HaveKeyWithValue("two", []string{"4"}))
			parser.IgnoreNotParsed()
			Expect(parser.NotParsed()).To(BeNil())
			parser.ReportNotParsed()
			Expect(parser.Error()).ToNot(HaveOccurred())
		})

		It("applies to a parser derived with WithSource, which shares the parsed values", func() {
			derived := parser.WithSource(structureTest.NewSource())
			parser.IgnoreNotParsed()
			Expect(derived.NotParsed()).To(BeNil())
			derived.ReportNotParsed()
			Expect(parser.Error()).ToNot(HaveOccurred())
		})
	})
})
