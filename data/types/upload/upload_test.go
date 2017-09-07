package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testData.RawBaseObject()
	rawObject["type"] = "upload"
	rawObject["version"] = "123456"
	rawObject["computerTime"] = "2014-06-11T06:00:00.000"
	rawObject["deviceTags"] = []string{"cgm"}
	rawObject["deviceModel"] = "455"
	rawObject["deviceManufacturers"] = []string{"cgm-peeps"}
	rawObject["deviceSerialNumber"] = "InsOmn-111111111"
	rawObject["timeProcessing"] = "utc-bootstrapping"
	rawObject["timezone"] = "US/Central"
	return rawObject
}

func NewMeta() interface{} {
	return &types.Meta{
		Type: "upload",
	}
}

var _ = Describe("Upload", func() {
	Context("version", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "version", "",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThanOrEqualTo(0, 5), "/version", NewMeta())},
			),
			Entry("is less than 5 characters", NewRawObject(), "version", "aaaa",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is 5 characters", NewRawObject(), "version", "aaaaa"),
			Entry("is more than 5 characters", NewRawObject(), "version", "aaaaabb"),
		)
	})

	Context("deviceTags", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty array", NewRawObject(), "deviceTags", []string{},
				[]*service.Error{testData.ComposeError(service.ErrorValueEmpty(), "/deviceTags", NewMeta())},
			),
			Entry("is not one of the allowed types", NewRawObject(), "deviceTags", []string{"not-valid"},
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/0", NewMeta())},
			),
			Entry("is not one of the allowed types", NewRawObject(), "deviceTags", []string{"bgm", "cgm", "not-valid"},
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("not-valid", []string{"insulin-pump", "cgm", "bgm"}), "/deviceTags/2", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is insulin-pump", NewRawObject(), "deviceTags", []string{"insulin-pump"}),
			Entry("is cgm", NewRawObject(), "deviceTags", []string{"cgm"}),
			Entry("is bgm", NewRawObject(), "deviceTags", []string{"bgm"}),
			Entry("is multiple", NewRawObject(), "deviceTags", []string{"bgm", "cgm", "insulin-pump"}),
		)
	})

	Context("deviceManufacturers", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty array", NewRawObject(), "deviceManufacturers", []string{},
				[]*service.Error{testData.ComposeError(service.ErrorValueEmpty(), "/deviceManufacturers", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is one item", NewRawObject(), "deviceManufacturers", []string{"insulin-pump-people"}),
			Entry("is multiple items", NewRawObject(), "deviceManufacturers", []string{"bgm-people", "cgm-people"}),
		)
	})

	Context("computerTime", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "computerTime", "",
				[]*service.Error{testData.ComposeError(service.ErrorValueTimeNotValid("", "2006-01-02T15:04:05"), "/computerTime", NewMeta())},
			),
			Entry("is zulu time", NewRawObject(), "computerTime", "2013-05-04T03:58:44.584Z",
				[]*service.Error{testData.ComposeError(service.ErrorValueTimeNotValid("2013-05-04T03:58:44.584Z", "2006-01-02T15:04:05"), "/computerTime", NewMeta())},
			),
			Entry("is offset time", NewRawObject(), "computerTime", "2013-05-04T03:58:44-08:00",
				[]*service.Error{testData.ComposeError(service.ErrorValueTimeNotValid("2013-05-04T03:58:44-08:00", "2006-01-02T15:04:05"), "/computerTime", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is non-zulu time", NewRawObject(), "computerTime", "2013-05-04T03:58:44.584"),
		)
	})

	Context("deviceModel", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "deviceModel", "",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(0, 1), "/deviceModel", NewMeta())},
			),
			Entry("is 1 character", NewRawObject(), "deviceModel", "x",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(1, 1), "/deviceModel", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is 2 characters", NewRawObject(), "deviceModel", "xx"),
			Entry("is more than 2 characters", NewRawObject(), "deviceModel", "model-x"),
		)
	})

	Context("deviceSerialNumber", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "deviceSerialNumber", "",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(0, 1), "/deviceSerialNumber", NewMeta())},
			),
			Entry("is 1 character", NewRawObject(), "deviceSerialNumber", "x",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(1, 1), "/deviceSerialNumber", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is 2 characters", NewRawObject(), "deviceSerialNumber", "xx"),
			Entry("is more than 2 characters", NewRawObject(), "deviceSerialNumber", "model-x"),
		)
	})

	Context("timezone", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "timezone", "",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(0, 1), "/timezone", NewMeta())},
			),
			Entry("is only one character", NewRawObject(), "timezone", "a",
				[]*service.Error{testData.ComposeError(service.ErrorLengthNotGreaterThan(1, 1), "/timezone", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is set", NewRawObject(), "timezone", "US/Central"),
		)
	})

	Context("timeProcessing", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "timeProcessing", "",
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing", NewMeta())},
			),
			Entry("is not of predefined type", NewRawObject(), "timeProcessing", "invalid-time-processing",
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("invalid-time-processing", []string{"across-the-board-timezone", "utc-bootstrapping", "none"}), "/timeProcessing", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is none", NewRawObject(), "timeProcessing", "none"),
			Entry("is utc-bootstrapping", NewRawObject(), "timeProcessing", "utc-bootstrapping"),
			Entry("is across-the-board-timezone", NewRawObject(), "timeProcessing", "across-the-board-timezone"),
		)
	})
})
