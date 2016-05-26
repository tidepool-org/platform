package base_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Base", func() {

	// TODO_DATA: Need to find another way to test base not using sample/sub

	// var rawObject = testing.RawBaseObject()
	// var meta = &base.Meta{
	// 	Type: "sample",
	// }

	// rawObject["type"] = "sample"
	// rawObject["subType"] = "sub"
	// rawObject["boolean"] = true
	// rawObject["integer"] = 7
	// rawObject["float"] = 4.5
	// rawObject["string"] = "aaa"
	// rawObject["stringArray"] = []string{"bach", "blech"}
	// rawObject["object"] = map[string]interface{}{
	// 	"one": 1,
	// 	"two": "two",
	// 	"three": map[string]interface{}{
	// 		"a": "apple",
	// 	},
	// }
	// rawObject["objectArray"] = []map[string]interface{}{
	// 	{
	// 		"alpha": "a",
	// 	},
	// 	{
	// 		"bravo": "b",
	// 	},
	// }
	// rawObject["interface"] = "yes"
	// rawObject["interfaceArray"] = []interface{}{
	// 	"alpha", map[string]interface{}{"alpha": "a"},
	// 	map[string]interface{}{"bravo": "b"},
	// 	-999,
	// }
	// rawObject["timeString"] = "2013-05-04T03:58:44-08:00"

	// Context("userId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", rawObject, "userId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 10), "/userId", meta)},
	// 		),
	// 		Entry("is less than 10 characters", rawObject, "userId", "123456789",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(9, 10), "/userId", meta)},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("has id 10 characters in length", rawObject, "userId", "b676436f60"),
	// 		Entry("has id more 10 characters in length", rawObject, "userId", "b676436f60-b676436f60"),
	// 	)
	// })

	// Context("uploadId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", rawObject, "uploadId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/uploadId", meta)},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("has string id", rawObject, "uploadId", "upid_b856b0e6e519"),
	// 		Entry("has string id 1 or more characters", rawObject, "uploadId", "d"),
	// 	)
	// })

	// Context("deviceId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", rawObject, "deviceId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceId", meta)},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is given string id", rawObject, "deviceId", "InsOmn-111111111"),
	// 		Entry("has string id 1 or more characters", rawObject, "deviceId", "d"),
	// 	)
	// })

	// Context("timezoneOffset", func() {
	// 	DescribeTable("timezoneOffset valid", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", rawObject, "timezoneOffset", 480),
	// 		Entry("is zero", rawObject, "timezoneOffset", 0),
	// 		Entry("is negative", rawObject, "timezoneOffset", -100),
	// 	)
	// })

	// Context("conversionOffset", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is negative", rawObject, "conversionOffset", -1,
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/conversionOffset", meta)},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", rawObject, "conversionOffset", 45),
	// 		Entry("is zero", rawObject, "conversionOffset", 0),
	// 	)
	// })

	// Context("clockDriftOffset", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is negative", rawObject, "clockDriftOffset", -1,
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/clockDriftOffset", meta)},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", rawObject, "clockDriftOffset", 45),
	// 		Entry("is zero", rawObject, "clockDriftOffset", 0),
	// 	)
	// })

	// Context("time", func() {
	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is zulu time", rawObject, "time", "2013-05-04T03:58:44.584Z"),
	// 	)

	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is non zulu time", rawObject, "time", "2013-05-04T03:58:44.584",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2013-05-04T03:58:44.584", "2006-01-02T15:04:05Z"), "/time", meta)},
	// 		),
	// 	)
	// })
})
