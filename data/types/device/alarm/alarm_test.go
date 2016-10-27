package alarm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testData.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "alarm"
	rawObject["alarmType"] = "other"
	rawObject["status"] = NewEmbeddedStatus("deviceEvent", "status")
	return rawObject
}

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "alarm",
	}
}

func NewEmbeddedStatus(datumType interface{}, subType interface{}) map[string]interface{} {
	var rawStatus = testData.RawBaseObject()

	if datumType != nil {
		rawStatus["type"] = datumType
	}
	if subType != nil {
		rawStatus["subType"] = subType
	}
	rawStatus["status"] = "suspended"
	rawStatus["duration"] = 360000
	rawStatus["reason"] = map[string]interface{}{"suspended": "automatic", "resumed": "automatic"}

	return rawStatus
}

var _ = Describe("Alarm", func() {
	Context("alarmType", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "alarmType", "",
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", NewMeta())},
			),
			Entry("is not one of the predefined types", NewRawObject(), "alarmType", "bad-robot",
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("bad-robot", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", NewMeta())},
			),
		)

		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is low_insulin type", NewRawObject(), "alarmType", "low_insulin"),
			Entry("is low_power type", NewRawObject(), "alarmType", "low_power"),
			Entry("is no_power type", NewRawObject(), "alarmType", "no_power"),
			Entry("is occlusion type", NewRawObject(), "alarmType", "occlusion"),
			Entry("is no_delivery type", NewRawObject(), "alarmType", "no_delivery"),
			Entry("is auto_off type", NewRawObject(), "alarmType", "auto_off"),
			Entry("is over_limit type", NewRawObject(), "alarmType", "over_limit"),
			Entry("is other type", NewRawObject(), "alarmType", "other"),
		)
	})

	Context("status", func() {
		DescribeTable("invalid when", testData.ExpectFieldNotValid,
			Entry("status is not an object", NewRawObject(), "status", "string",
				[]*service.Error{testData.ComposeError(service.ErrorTypeNotObject("string"), "/status", NewMeta())},
			),
			Entry("type is missing", NewRawObject(), "status", NewEmbeddedStatus(nil, "status"),
				[]*service.Error{testData.ComposeError(service.ErrorValueNotExists(), "/status/type", NewMeta())},
			),
			Entry("type is not valid", NewRawObject(), "status", NewEmbeddedStatus("invalid", "status"),
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"deviceEvent"}), "/status/type", NewMeta())},
			),
			Entry("subType is missing", NewRawObject(), "status", NewEmbeddedStatus("deviceEvent", nil),
				[]*service.Error{testData.ComposeError(service.ErrorValueNotExists(), "/status/subType", NewMeta())},
			),
			Entry("subType is not valid", NewRawObject(), "status", NewEmbeddedStatus("deviceEvent", "invalid"),
				[]*service.Error{testData.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"status"}), "/status/subType", NewMeta())},
			),
		)
	})
})
