package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataContext "github.com/tidepool-org/platform/data/context"
	dataParser "github.com/tidepool-org/platform/data/parser"
	logNull "github.com/tidepool-org/platform/log/null"
)

func ExpectSerializedArray(expected interface{}, array []interface{}, parserFunc func(parser data.ArrayParser) interface{}) {
	ctx, err := dataContext.NewStandard(logNull.NewLogger())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ctx).ToNot(gomega.BeNil())
	parser, err := dataParser.NewStandardArray(ctx, &array, dataParser.AppendErrorNotParsed)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(parser).ToNot(gomega.BeNil())
	gomega.Expect(parserFunc(parser)).To(gomega.Equal(expected))
	gomega.Expect(ctx.Errors()).To(gomega.BeEmpty())
}

func ExpectSerializedObject(expected interface{}, object map[string]interface{}, parserFunc func(parser data.ObjectParser) interface{}) {
	ctx, err := dataContext.NewStandard(logNull.NewLogger())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ctx).ToNot(gomega.BeNil())
	parser, err := dataParser.NewStandardObject(ctx, &object, dataParser.AppendErrorNotParsed)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(parser).ToNot(gomega.BeNil())
	gomega.Expect(parserFunc(parser)).To(gomega.Equal(expected))
	gomega.Expect(ctx.Errors()).To(gomega.BeEmpty())
}
