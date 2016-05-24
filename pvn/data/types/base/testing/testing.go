package testing

import (
	"log"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/normalizer"
	"github.com/tidepool-org/platform/pvn/data/parser"
	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"
)

func RawBaseObject() map[string]interface{} {
	return map[string]interface{}{
		"userId":           "b676436f60",
		"_groupId":         "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   720,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

func reportAndFailOnErrors(testContext *context.Standard, step string) {
	gomega.Expect(testContext.Errors()).To(gomega.BeEmpty(), step)
	for _, err := range testContext.Errors() {
		log.Println(err.Source, err.Detail)
	}
}

var ExpectFieldIsValid = func(object map[string]interface{}, field string, val interface{}) {
	object[field] = val
	testContext := context.NewStandard()
	standardValidator, err := validator.NewStandard(testContext)
	gomega.Expect(err).To(gomega.BeNil())
	objectParser, err := parser.NewStandardObject(testContext, &object)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Initialization:")
	parsedObject, err := types.Parse(testContext, objectParser)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Parsing:")
	parsedObject.Validate(standardValidator)
	reportAndFailOnErrors(testContext, "Validation:")
}

func SetExpectedErrorSource(expectedError *service.Error, source string) *service.Error {
	expectedError.Source = &service.Source{Parameter: "", Pointer: source}
	return expectedError
}

var ExpectFieldNotValid = func(object map[string]interface{}, field string, val interface{}, expectedErrors []*service.Error) {
	object[field] = val
	testContext := context.NewStandard()
	standardValidator, err := validator.NewStandard(testContext)
	gomega.Expect(err).To(gomega.BeNil())
	objectParser, err := parser.NewStandardObject(testContext, &object)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Initialization:")
	parsedObject, err := types.Parse(testContext, objectParser)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Parsing:")
	parsedObject.Validate(standardValidator)
	gomega.Expect(testContext.Errors()).ToNot(gomega.BeEmpty())
	gomega.Expect(testContext.Errors()).To(gomega.HaveLen(len(expectedErrors)))
	for i := range expectedErrors {
		gomega.Expect(testContext.Errors()).To(gomega.ContainElement(expectedErrors[i]))
	}
}

var ParseAndNormalize = func(object map[string]interface{}, field string, val interface{}) data.Datum {
	object[field] = val
	testContext := context.NewStandard()
	standardValidator, err := validator.NewStandard(testContext)
	gomega.Expect(err).To(gomega.BeNil())
	objectParser, err := parser.NewStandardObject(testContext, &object)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Initialization:")
	parsedObject, err := types.Parse(testContext, objectParser)
	gomega.Expect(err).To(gomega.BeNil())
	reportAndFailOnErrors(testContext, "Parsing:")
	parsedObject.Validate(standardValidator)
	reportAndFailOnErrors(testContext, "Validate:")
	standardNormalizer, err := normalizer.NewStandard(testContext)
	gomega.Expect(err).To(gomega.BeNil())
	parsedObject.Normalize(standardNormalizer)
	return parsedObject
}
