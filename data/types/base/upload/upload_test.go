package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Upload", func() {

	var rawObject = testing.RawBaseObject()
	var meta = &base.Meta{
		Type: "upload",
	}

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
			Entry("is empty", rawObject, "version", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(0, 5), "/version", meta)},
			),
			Entry("is less than 6 characters", rawObject, "version", "aaaaa",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(5, 5), "/version", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 6 characters", rawObject, "version", "aaaaaa"),
			Entry("is more than 6 characters", rawObject, "version", "aaaaaabb"),
		)
	})

	Context("deviceTags", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty array", rawObject, "deviceTags", []string{},
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceTags", meta)},
			),
			Entry("is not one of the allowed types", rawObject, "deviceTags", []string{"not-valid"},
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/0", meta)},
			),
			Entry("is not one of the allowed types", rawObject, "deviceTags", []string{"bgm", "cgm", "not-valid"},
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/2", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is insulin-pump", rawObject, "deviceTags", []string{"insulin-pump"}),
			Entry("is cgm", rawObject, "deviceTags", []string{"cgm"}),
			Entry("is bgm", rawObject, "deviceTags", []string{"bgm"}),
			Entry("is multiple", rawObject, "deviceTags", []string{"bgm", "cgm", "insulin-pump"}),
		)
	})

	Context("deviceManufacturers", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty array", rawObject, "deviceManufacturers", []string{},
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/deviceManufacturers", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is one item", rawObject, "deviceManufacturers", []string{"insulin-pump-people"}),
			Entry("is multiple items", rawObject, "deviceManufacturers", []string{"bgm-people", "cgm-people"}),
		)
	})

	Context("computerTime", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "computerTime", "",
				[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/computerTime", meta)},
			),
			Entry("is zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584Z",
				[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2013-05-04T03:58:44.584Z", "2006-01-02T15:04:05"), "/computerTime", meta)},
			),
			Entry("is offset time", rawObject, "computerTime", "2013-05-04T03:58:44-08:00",
				[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2013-05-04T03:58:44-08:00", "2006-01-02T15:04:05"), "/computerTime", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is non-zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584"),
		)
	})

	Context("deviceModel", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "deviceModel", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(0, 1), "/deviceModel", meta)},
			),
			Entry("is 1 character", rawObject, "deviceModel", "x",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(1, 1), "/deviceModel", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 2 characters", rawObject, "deviceModel", "xx"),
			Entry("is more than 2 characters", rawObject, "deviceModel", "model-x"),
		)
	})

	Context("deviceSerialNumber", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "deviceSerialNumber", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(0, 1), "/deviceSerialNumber", meta)},
			),
			Entry("is 1 character", rawObject, "deviceSerialNumber", "x",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(1, 1), "/deviceSerialNumber", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 2 characters", rawObject, "deviceSerialNumber", "xx"),
			Entry("is more than 2 characters", rawObject, "deviceSerialNumber", "model-x"),
		)
	})

	Context("timezone", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "timezone", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(0, 1), "/timezone", meta)},
			),
			Entry("is only one character", rawObject, "timezone", "a",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(1, 1), "/timezone", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is set", rawObject, "timezone", "US/Central"),
		)
	})

	Context("timeProcessing", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "timeProcessing", "",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing", meta)},
			),
			Entry("is not of predefined type", rawObject, "timeProcessing", "invalid-time-processing",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("invalid-time-processing", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is none", rawObject, "timeProcessing", "none"),
			Entry("is utc-bootstrapping", rawObject, "timeProcessing", "utc-bootstrapping"),
			Entry("is across-the-board-timezone", rawObject, "timeProcessing", "across-the-board-timezone"),
		)
	})
})
