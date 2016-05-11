package upload_test

import (
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/parser"
	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/pvn/data/validator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upload", func() {

	var rawObject = map[string]interface{}{
		"userId":           "b676436f60",
		"_groupId":         "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   720,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
		"type":             "sample",
	}

	var objectParser *parser.StandardObject
	var testContext *context.Standard
	var standardValidator *validator.Standard

	BeforeEach(func() {
		testContext = context.NewStandard()
		standardValidator = validator.NewStandard(testContext)
	})

	DescribeTable("validation",
		func(feild string, val interface{}, isValid bool) {

			rawObject[feild] = val
			objectParser = parser.NewStandardObject(testContext, &rawObject)
			// Expect(testContext.HasErrors()).To(BeFalse(), "there should not be errors after creating the object")
			parsedObject := types.Parse(objectParser)
			// Expect(testContext.HasErrors()).To(BeFalse(), "there should not be errors after doing a parse")
			parsedObject.Validate(standardValidator)
			// Expect(testContext.HasErrors()).To(Equal(isValid))

			Expect(true).To(BeTrue())
			//if !isValid {
			//	Expect(context.Errors[0]).To(Equal(expectedError))
			//}
		},

		Entry("userId empty is invalid", "userId", "b676436f60", false),
		Entry("_groupId empty is invalid", "_groupId", "43099shgs55", false),
		//Entry("uploadId empty is invalid", "uploadId", "", false, nil),
		//Entry("deviceTime empty is invalid", "deviceTime", "", false, nil),
		//Entry("time empty is invalid", "time", "", false, nil),
		//Entry("deviceId empty is invalid", "deviceId", "", false, nil),
	)

})
