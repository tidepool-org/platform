package testing

import (
	"log"

	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pvn/data/context"
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
	if testContext.HasErrors() {
		for i := range testContext.GetErrors() {
			log.Println(testContext.GetErrors()[i].Source, testContext.GetErrors()[i].Detail)
		}
		Expect(testContext.HasErrors()).To(BeFalse(), step)
	}
}

var ExpectFieldIsValid = func(object map[string]interface{}, field string, val interface{}) {
	object[field] = val
	testContext := context.NewStandard()
	standardValidator := validator.NewStandard(testContext)
	objectParser := parser.NewStandardObject(testContext, &object)
	reportAndFailOnErrors(testContext, "Initialization:")
	parsedObject := types.Parse(objectParser)
	reportAndFailOnErrors(testContext, "Parsing:")
	parsedObject.Validate(standardValidator)
	reportAndFailOnErrors(testContext, "Validation:")
}

var ExpectFieldNotValid = func(object map[string]interface{}, field string, val interface{}, expectedErrors []*service.Error) {
	object[field] = val
	testContext := context.NewStandard()
	standardValidator := validator.NewStandard(testContext)
	objectParser := parser.NewStandardObject(testContext, &object)
	reportAndFailOnErrors(testContext, "Initialization:")
	parsedObject := types.Parse(objectParser)
	reportAndFailOnErrors(testContext, "Parsing:")
	parsedObject.Validate(standardValidator)
	Expect(testContext.HasErrors()).To(BeTrue())
	Expect(len(testContext.GetErrors())).To(Equal(len(expectedErrors)))
}
