package base_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/ginkgo/extensions/table"
)

// func NewRawObject() map[string]interface{} {
// 	rawObject := testing.RawBaseObject()
// 	rawObject["type"] = "sample"
// 	rawObject["subType"] = "sub"
// 	rawObject["boolean"] = true
// 	rawObject["integer"] = 7
// 	rawObject["float"] = 4.5
// 	rawObject["string"] = "aaa"
// 	rawObject["stringArray"] = []string{"bach", "blech"}
// 	rawObject["object"] = map[string]interface{}{
// 		"one": 1,
// 		"two": "two",
// 		"three": map[string]interface{}{
// 			"a": "apple",
// 		},
// 	}
// 	rawObject["objectArray"] = []map[string]interface{}{
// 		{
// 			"alpha": "a",
// 		},
// 		{
// 			"bravo": "b",
// 		},
// 	}
// 	rawObject["interface"] = "yes"
// 	rawObject["interfaceArray"] = []interface{}{
// 		"alpha", map[string]interface{}{"alpha": "a"},
// 		map[string]interface{}{"bravo": "b"},
// 		-999,
// 	}
// 	rawObject["timeString"] = "2013-05-04T03:58:44-08:00"
// 	return rawObject
// }

// func NewMeta() interface{} {
// 	return &base.Meta{
// 		Type: "sample",
// 	}
// }

// TODO_DATA: Need to find another way to test base not using sample/sub

var _ = PDescribe("Base", func() {
	// Context("_userId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", NewRawObject(), "_userId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 10), "/userId", NewMeta())},
	// 		),
	// 		Entry("is less than 10 characters", NewRawObject(), "_userId", "123456789",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(9, 10), "/userId", NewMeta())},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("has id 10 characters in length", NewRawObject(), "_userId", "b676436f60"),
	// 		Entry("has id more 10 characters in length", NewRawObject(), "_userId", "b676436f60-b676436f60"),
	// 	)
	// })

	// Context("uploadId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", NewRawObject(), "uploadId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueEmpty(), "/uploadId", NewMeta())},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("has string id", NewRawObject(), "uploadId", "upid_b856b0e6e519"),
	// 		Entry("has string id 1 or more characters", NewRawObject(), "uploadId", "d"),
	// 	)
	// })

	// Context("deviceId", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is empty", NewRawObject(), "deviceId", "",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueEmpty(), "/deviceId", NewMeta())},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is given string id", NewRawObject(), "deviceId", "InsOmn-111111111"),
	// 		Entry("has string id 1 or more characters", NewRawObject(), "deviceId", "d"),
	// 	)
	// })

	// Context("timezoneOffset", func() {
	// 	DescribeTable("timezoneOffset valid", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", NewRawObject(), "timezoneOffset", 480),
	// 		Entry("is zero", NewRawObject(), "timezoneOffset", 0),
	// 		Entry("is negative", NewRawObject(), "timezoneOffset", -100),
	// 	)
	// })

	// Context("conversionOffset", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is negative", NewRawObject(), "conversionOffset", -1,
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/conversionOffset", NewMeta())},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", NewRawObject(), "conversionOffset", 45),
	// 		Entry("is zero", NewRawObject(), "conversionOffset", 0),
	// 	)
	// })

	// Context("clockDriftOffset", func() {
	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is negative", NewRawObject(), "clockDriftOffset", -1,
	// 			[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/clockDriftOffset", NewMeta())},
	// 		),
	// 	)

	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is greater than zero", NewRawObject(), "clockDriftOffset", 45),
	// 		Entry("is zero", NewRawObject(), "clockDriftOffset", 0),
	// 	)
	// })

	// Context("time", func() {
	// 	DescribeTable("valid when", testing.ExpectFieldIsValid,
	// 		Entry("is zulu time", NewRawObject(), "time", "2013-05-04T03:58:44.584Z"),
	// 	)

	// 	DescribeTable("invalid when", testing.ExpectFieldNotValid,
	// 		Entry("is non zulu time", NewRawObject(), "time", "2013-05-04T03:58:44.584",
	// 			[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2013-05-04T03:58:44.584", "2006-01-02T15:04:05Z"), "/time", NewMeta())},
	// 		),
	// 	)
	// })
})
