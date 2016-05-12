package base_test

import (
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/parser"
	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/pvn/data/validator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base", func() {

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
		//types from sample
		"subType":     "sub",
		"boolean":     true,
		"integer":     7,
		"float":       4.5,
		"string":      "aaa",
		"stringArray": []string{"bach", "blech"},
		"object": map[string]interface{}{
			"one": 1,
			"two": "two",
			"three": map[string]interface{}{
				"a": "apple",
			},
		},
		"objectArray": []map[string]interface{}{
			{
				"alpha": "a",
			},
			{
				"bravo": "b",
			},
		},
		"interface": "yes",
		"interfaceArray": []interface{}{
			"alpha", map[string]interface{}{"alpha": "a"},
			map[string]interface{}{"bravo": "b"},
			-999,
		},
	}

	var objectParser *parser.StandardObject
	var testContext *context.Standard
	var standardValidator *validator.Standard

	BeforeEach(func() {
		testContext = context.NewStandard()
		standardValidator = validator.NewStandard(testContext)
	})

	var isValid = func(feild string, val interface{}) {
		rawObject[feild] = val

		objectParser = parser.NewStandardObject(testContext, &rawObject)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected initialization errors")
		parsedObject := types.Parse(objectParser)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected parse errors")
		parsedObject.Validate(standardValidator)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected validate errors")
	}

	/*var notValid = func(feild string, val interface{}, expectedErrors *service.Errors) {
		rawObject[feild] = val
		objectParser = parser.NewStandardObject(testContext, &rawObject)
		Expect(testContext.HasErrors()).To(BeFalse(), "there should not be errors after creating the object")
		parsedObject := types.Parse(objectParser)
		Expect(testContext.HasErrors()).To(BeFalse(), "there should not be errors after doing a parse")
		parsedObject.Validate(standardValidator)
		Expect(testContext.HasErrors()).To(Equal(true))
		Expect(testContext.Errors).To(Equal(expectedErrors))
	}*/

	// "userId":           "b676436f60",
	// 		"_groupId":         "43099shgs55",
	// 		"uploadId":         "upid_b856b0e6e519",
	// 		"deviceTime":       "2014-06-11T06:00:00.000",
	// 		"time":             "2014-06-11T06:00:00.000Z",
	// 		"timezoneOffset":   720,
	// 		"conversionOffset": 0,
	// 		"clockDriftOffset": 0,
	// 		"deviceId":         "InsOmn-111111111",

	DescribeTable("userId valid", isValid,
		Entry("when given string id", "userId", "b676436f60"),
		Entry("when given string id", "userId", "dddddddddddddd"),
	)

	DescribeTable("uploadId valid", isValid,
		Entry("when given string id", "uploadId", "upid_b856b0e6e519"),
		Entry("when given string id", "uploadId", "dddddddddddddd"),
	)

	DescribeTable("deviceId valid", isValid,
		Entry("when given string id", "deviceId", "InsOmn-111111111"),
		Entry("when given string id", "deviceId", "dddddddddddddd"),
	)

	DescribeTable("deviceId valid", isValid,
		Entry("when given string id", "deviceId", "InsOmn-111111111"),
		Entry("when given string id", "deviceId", "dddddddddddddd"),
	)

	DescribeTable("timezoneOffset valid", isValid,
		Entry("when given string id", "timezoneOffset", 480),
		Entry("when given string id", "timezoneOffset", 0),
	)

	DescribeTable("conversionOffset valid", isValid,
		Entry("when given string id", "conversionOffset", 45),
		Entry("when given string id", "conversionOffset", 0),
	)

	DescribeTable("clockDriftOffset valid", isValid,
		Entry("when given string id", "clockDriftOffset", 45),
		Entry("when given string id", "clockDriftOffset", 0),
	)

	DescribeTable("deviceTime valid", isValid,
		Entry("when given string id", "deviceTime", "2014-06-11T06:00:00.000"),
		Entry("when given string id", "deviceTime", "2013-01-01T06:00:00.000"),
	)

})
