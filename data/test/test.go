package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

	standardContext, err := context.NewStandard(null.NewLogger())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(standardContext).ToNot(gomega.BeNil())

	standardObjectParser, err := parser.NewStandardObject(standardContext, &object, parser.AppendErrorNotParsed)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(standardObjectParser).ToNot(gomega.BeNil())

	validator := structureValidator.New()
	gomega.Expect(validator).ToNot(gomega.BeNil())

	normalizer := dataNormalizer.New()
	gomega.Expect(normalizer).ToNot(gomega.BeNil())

	parsedObject := dataTypesFactory.ParseDatum(standardObjectParser)
	gomega.Expect(parsedObject).ToNot(gomega.BeNil())
	gomega.Expect(*parsedObject).ToNot(gomega.BeNil())
	standardObjectParser.ProcessNotParsed()
	(*parsedObject).Validate(validator)
	(*parsedObject).Normalize(normalizer)

	gomega.Expect(standardContext.Errors()).To(gomega.ConsistOf(errors))
	gomega.Expect(normalizer.Error()).ToNot(gomega.HaveOccurred())

	return (*parsedObject)
}
