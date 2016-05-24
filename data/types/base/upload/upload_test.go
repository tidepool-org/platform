package upload_test

import (
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Upload", func() {

	var rawObject = testing.RawBaseObject()

	rawObject["type"] = "upload"
	rawObject["byUser"] = "123456789x"
	rawObject["version"] = "123456"
	rawObject["computerTime"] = "2014-06-11T06:00:00.000"
	rawObject["deviceTags"] = []string{"cgm"}
	rawObject["deviceModel"] = "455"
	rawObject["deviceManufacturers"] = []string{"cgm-peeps"}
	rawObject["deviceSerialNumber"] = "InsOmn-111111111"
	rawObject["timeProcessing"] = "utc-bootstrapping"
	rawObject["dataState"] = "running"
	rawObject["deduplicator"] = "something"
	rawObject["timezone"] = "US/Central"

	Context("version", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "version", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(0, 5), "/version")},
			),
			Entry("less than 6 characters", rawObject, "version", "aaaaa",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(5, 5), "/version")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("6 characters", rawObject, "version", "aaaaaa"),
			Entry("more than 6 characters", rawObject, "version", "aaaaaabb"),
		)

	})

	Context("deviceTags", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty array", rawObject, "deviceTags", []string{},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceTags")},
			),
			Entry("not one of the allowed types", rawObject, "deviceTags", []string{"not-valid"},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/0")},
			),
			Entry("not one of the allowed types", rawObject, "deviceTags", []string{"bgm", "cgm", "not-valid"},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/2")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("insulin-pump", rawObject, "deviceTags", []string{"insulin-pump"}),
			Entry("cgm", rawObject, "deviceTags", []string{"cgm"}),
			Entry("bgm", rawObject, "deviceTags", []string{"bgm"}),
			Entry("multiple", rawObject, "deviceTags", []string{"bgm", "cgm", "insulin-pump"}),
		)

	})

	Context("deviceManufacturers", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty array", rawObject, "deviceManufacturers", []string{},
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceManufacturers")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("one item", rawObject, "deviceManufacturers", []string{"insulin-pump-people"}),
			Entry("multiple items", rawObject, "deviceManufacturers", []string{"bgm-people", "cgm-people"}),
		)

	})

	Context("computerTime", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "computerTime", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/computerTime")},
			),
			Entry("zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584Z",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("2013-05-04T03:58:44.584Z", "2006-01-02T15:04:05"), "/computerTime")},
			),
			Entry("offset time", rawObject, "computerTime", "2013-05-04T03:58:44-08:00",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("2013-05-04T03:58:44-08:00", "2006-01-02T15:04:05"), "/computerTime")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("non-zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584"),
		)

	})

	Context("deviceModel", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "deviceModel", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(0, 1), "/deviceModel")},
			),
			Entry("1 character", rawObject, "deviceModel", "x",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(1, 1), "/deviceModel")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("2 characters", rawObject, "deviceModel", "xx"),
			Entry("more than 2 characters", rawObject, "deviceModel", "model-x"),
		)

	})

	Context("deviceSerialNumber", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "deviceSerialNumber", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(0, 1), "/deviceSerialNumber")},
			),
			Entry("1 character", rawObject, "deviceSerialNumber", "x",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(1, 1), "/deviceSerialNumber")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("2 characters", rawObject, "deviceSerialNumber", "xx"),
			Entry("more than 2 characters", rawObject, "deviceSerialNumber", "model-x"),
		)

	})

	Context("timezone", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "timezone", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(0, 1), "/timezone")},
			),
			Entry("only one character", rawObject, "timezone", "a",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(1, 1), "/timezone")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("set", rawObject, "timezone", "US/Central"),
		)

	})

	Context("timeProcessing", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "timeProcessing", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing")},
			),
			Entry("not of predefinded type", rawObject, "timeProcessing", "invalid-time-processing",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("invalid-time-processing", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("none", rawObject, "timeProcessing", "none"),
			Entry("utc-bootstrapping", rawObject, "timeProcessing", "utc-bootstrapping"),
			Entry("across-the-board-timezone", rawObject, "timeProcessing", "across-the-board-timezone"),
		)

	})
})
