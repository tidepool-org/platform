package testing

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/service"
)

func RawBaseObject() map[string]interface{} {
	return map[string]interface{}{
		"deviceTime":       "2014-06-11T06:00:00.000",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   720,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

func ComposeError(expectedError *service.Error, source string, meta interface{}) *service.Error {
	expectedError.Source = &service.Source{Parameter: "", Pointer: source}
	expectedError.Meta = meta
	return expectedError
}

var ExpectFieldIsValid = func(object map[string]interface{}, field string, value interface{}) {
	checkErrorsFromParseValidateNormalize(object, field, value, []*service.Error{})
}

var ExpectFieldNotValid = func(object map[string]interface{}, field string, value interface{}, expectedErrors []*service.Error) {
	checkErrorsFromParseValidateNormalize(object, field, value, expectedErrors)
}

var ParseAndNormalize = func(object map[string]interface{}, field string, value interface{}) data.Datum {
	return checkErrorsFromParseValidateNormalize(object, field, value, []*service.Error{})
}

func checkErrorsFromParseValidateNormalize(object map[string]interface{}, field string, value interface{}, errors []*service.Error) data.Datum {
	if value != nil {
		object[field] = value
	} else {
		delete(object, field)
	}

	standardContext, err := context.NewStandard(test.NewLogger())
	gomega.Expect(standardContext).ToNot(gomega.BeNil())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	standardValidator, err := validator.NewStandard(standardContext)
	gomega.Expect(standardValidator).ToNot(gomega.BeNil())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	standardObjectParser, err := parser.NewStandardObject(standardContext, &object, parser.AppendErrorNotParsed)
	gomega.Expect(standardObjectParser).ToNot(gomega.BeNil())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	standardNormalizer, err := normalizer.NewStandard(standardContext)
	gomega.Expect(standardNormalizer).ToNot(gomega.BeNil())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	parsedObject, err := types.Parse(standardObjectParser)
	gomega.Expect(parsedObject).ToNot(gomega.BeNil())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	standardObjectParser.ProcessNotParsed()
	parsedObject.Validate(standardValidator)
	parsedObject.Normalize(standardNormalizer)

	gomega.Expect(standardContext.Errors()).To(gomega.ConsistOf(errors))

	return parsedObject
}
