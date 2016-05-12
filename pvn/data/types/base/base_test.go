package base_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Base", func() {

	var rawObject = testing.RawBaseObject()

	rawObject["type"] = "sample"
	rawObject["subType"] = "sub"
	rawObject["boolean"] = true
	rawObject["integer"] = 7
	rawObject["float"] = 4.5
	rawObject["string"] = "aaa"
	rawObject["stringArray"] = []string{"bach", "blech"}
	rawObject["object"] = map[string]interface{}{
		"one": 1,
		"two": "two",
		"three": map[string]interface{}{
			"a": "apple",
		},
	}
	rawObject["objectArray"] = []map[string]interface{}{
		{
			"alpha": "a",
		},
		{
			"bravo": "b",
		},
	}
	rawObject["interface"] = "yes"
	rawObject["interfaceArray"] = []interface{}{
		"alpha", map[string]interface{}{"alpha": "a"},
		map[string]interface{}{"bravo": "b"},
		-999,
	}
	rawObject["timeString"] = "2013-05-04T03:58:44-08:00"

	Context("userId", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "userId", "", []*service.Error{&service.Error{}}),
			Entry("less than 10 characters", rawObject, "userId", "123456789", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("id 10 characters in length", rawObject, "userId", "b676436f60"),
			Entry("id more 10 characters in length", rawObject, "userId", "b676436f60-b676436f60"),
		)

	})

	Context("uploadId", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "uploadId", "", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("string id", rawObject, "uploadId", "upid_b856b0e6e519"),
			Entry("string id 1 or more characters", rawObject, "uploadId", "d"),
		)

	})

	Context("deviceId", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "deviceId", "", []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("when given string id", rawObject, "deviceId", "InsOmn-111111111"),
			Entry("string id 1 or more characters", rawObject, "deviceId", "d"),
		)

	})

	Context("timezoneOffset", func() {

		DescribeTable("timezoneOffset valid", testing.ExpectFieldIsValid,
			Entry("greater than zero", rawObject, "timezoneOffset", 480),
			Entry("zero", rawObject, "timezoneOffset", 0),
			Entry("negative", rawObject, "timezoneOffset", -100),
		)

	})

	Context("conversionOffset", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "conversionOffset", -1, []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("greater than zero", rawObject, "conversionOffset", 45),
			Entry("zero", rawObject, "conversionOffset", 0),
		)

	})

	Context("clockDriftOffset", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "clockDriftOffset", -1, []*service.Error{&service.Error{}}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("greater than zero", rawObject, "clockDriftOffset", 45),
			Entry("zero", rawObject, "clockDriftOffset", 0),
		)

	})

	Context("time", func() {

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("zulu time", rawObject, "time", "2013-05-04T03:58:44.584Z"),
		)

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("non zulu time", rawObject, "time", "2013-05-04T03:58:44.584", []*service.Error{&service.Error{}}),
		)

	})

})
