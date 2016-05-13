package upload_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
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

	Context("byUser", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "byUser", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("less than 10 characters", rawObject, "byUser", "123456789", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("10 characters", rawObject, "byUser", "b676436f60"),
			Entry("more than 10 characters", rawObject, "byUser", "b676436f60-b676436f60"),
		)

	})

	Context("version", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "version", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("less than 6 characters", rawObject, "version", "aaaaa", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("6 characters", rawObject, "version", "aaaaaa"),
			Entry("more than 6 characters", rawObject, "version", "aaaaaabb"),
		)

	})

	Context("deviceTags", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty array", rawObject, "deviceTags", []string{}, []*service.Error{validator.ErrorValueNotTrue()}),
			//Entry("empty item", rawObject, "deviceTags", []string{""}, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("not one of the allowed types", rawObject, "deviceTags", []string{"not-valid"}, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("not one of the allowed types", rawObject, "deviceTags", []string{"bgm", "cgm", "not-valid"}, []*service.Error{validator.ErrorValueNotTrue()}),
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
			Entry("empty array", rawObject, "deviceManufacturers", []string{}, []*service.Error{validator.ErrorValueNotTrue()}),
			//Entry("empty item", rawObject, "deviceManufacturers", []string{""}, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("one item", rawObject, "deviceManufacturers", []string{"insulin-pump-people"}),
			Entry("multiple items", rawObject, "deviceManufacturers", []string{"bgm-people", "cgm-people"}),
		)

	})

	Context("computerTime", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "computerTime", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584Z", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("offset time", rawObject, "computerTime", "2013-05-04T03:58:44-08:00", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("non-zulu time", rawObject, "computerTime", "2013-05-04T03:58:44.584"),
		)

	})

	Context("deviceModel", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "deviceModel", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("1 character", rawObject, "deviceModel", "x", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("2 characters", rawObject, "deviceModel", "xx"),
			Entry("more than 2 characters", rawObject, "deviceModel", "model-x"),
		)

	})

	Context("deviceSerialNumber", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "deviceSerialNumber", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("1 character", rawObject, "deviceSerialNumber", "x", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("2 characters", rawObject, "deviceSerialNumber", "xx"),
			Entry("more than 2 characters", rawObject, "deviceSerialNumber", "model-x"),
		)

	})

	Context("timeProcessing", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "timeProcessing", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("not of predefinded type", rawObject, "timeProcessing", "invalid-time-processing", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("none", rawObject, "timeProcessing", "none"),
			Entry("utc-bootstrapping", rawObject, "timeProcessing", "utc-bootstrapping"),
			Entry("across-the-board-timezone", rawObject, "timeProcessing", "across-the-board-timezone"),
		)

	})

	Context("dataState", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "dataState", "", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("set", rawObject, "dataState", "running"),
		)

	})

	Context("deduplicator", func() {

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("set", rawObject, "deduplicator", map[string]interface{}{"todod": 7}),
		)

	})

})
