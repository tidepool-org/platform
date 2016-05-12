package base_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/parser"
	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"
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
		"timeString": "2013-05-04T03:58:44-08:00",
	}

	var isValid = func(feild string, val interface{}) {
		rawObject[feild] = val

		testContext := context.NewStandard()
		standardValidator := validator.NewStandard(testContext)
		objectParser := parser.NewStandardObject(testContext, &rawObject)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected initialization errors")
		parsedObject := types.Parse(objectParser)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected parse errors")
		parsedObject.Validate(standardValidator)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected validate errors")
	}

	var notValid = func(feild string, val interface{}, expectedErrors []*service.Error) {
		rawObject[feild] = val
		testContext := context.NewStandard()
		standardValidator := validator.NewStandard(testContext)
		objectParser := parser.NewStandardObject(testContext, &rawObject)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected initialization errors")
		parsedObject := types.Parse(objectParser)
		Expect(testContext.HasErrors()).To(BeFalse(), "Unexpected parse errors")
		parsedObject.Validate(standardValidator)
		Expect(testContext.HasErrors()).To(BeTrue())
		Expect(len(testContext.GetErrors())).To(Equal(len(expectedErrors)))
	}

	Context("userId", func() {

		DescribeTable("invalid when", notValid,
			Entry("empty", "userId", "", []*service.Error{&service.Error{}}),
			Entry("less than 10 characters", "userId", "123456789", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", isValid,
			Entry("id 10 characters in length", "userId", "b676436f60"),
			Entry("id more 10 characters in length", "userId", "b676436f60-b676436f60"),
		)

	})

	Context("uploadId", func() {

		DescribeTable("invalid when", notValid,
			Entry("empty", "uploadId", "", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", isValid,
			Entry("string id", "uploadId", "upid_b856b0e6e519"),
			Entry("string id 1 or more characters", "uploadId", "d"),
		)

	})

	Context("deviceId", func() {

		DescribeTable("invalid when", notValid,
			Entry("empty", "deviceId", "", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", isValid,
			Entry("when given string id", "deviceId", "InsOmn-111111111"),
			Entry("string id 1 or more characters", "deviceId", "d"),
		)

	})

	Context("timezoneOffset", func() {

		DescribeTable("timezoneOffset valid", isValid,
			Entry("greater than zero", "timezoneOffset", 480),
			Entry("zero", "timezoneOffset", 0),
			Entry("negative", "timezoneOffset", -100),
		)

	})

	Context("conversionOffset", func() {

		DescribeTable("invalid when", notValid,
			Entry("negative", "conversionOffset", -1, []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", isValid,
			Entry("greater than zero", "conversionOffset", 45),
			Entry("zero", "conversionOffset", 0),
		)

	})

	Context("clockDriftOffset", func() {

		DescribeTable("invalid when", notValid,
			Entry("negative", "clockDriftOffset", -1, []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", isValid,
			Entry("greater than zero", "clockDriftOffset", 45),
			Entry("zero", "clockDriftOffset", 0),
		)

	})

	Context("time", func() {

		DescribeTable("valid when", isValid,
			Entry("zulu time", "time", "2013-05-04T03:58:44.584Z"),
		)

		DescribeTable("invalid when", notValid,
			Entry("non zulu time", "time", "2013-05-04T03:58:44.584", []*service.Error{&service.Error{}}),
		)

	})

})
