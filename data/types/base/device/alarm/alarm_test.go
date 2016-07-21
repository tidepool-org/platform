package alarm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
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
	var rawStatus = testing.RawBaseObject()

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
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", NewRawObject(), "alarmType", "",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", NewMeta())},
			),
			Entry("is not one of the predefined types", NewRawObject(), "alarmType", "bad-robot",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("bad-robot", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
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
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("status is not an object", NewRawObject(), "status", "string",
				[]*service.Error{testing.ComposeError(parser.ErrorTypeNotObject("string"), "/status", NewMeta())},
			),
			Entry("type is missing", NewRawObject(), "status", NewEmbeddedStatus(nil, "status"),
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/status/type", NewMeta())},
			),
			Entry("type is not valid", NewRawObject(), "status", NewEmbeddedStatus("invalid", "status"),
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("invalid", []string{"deviceEvent"}), "/status/type", NewMeta())},
			),
			Entry("subType is missing", NewRawObject(), "status", NewEmbeddedStatus("deviceEvent", nil),
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/status/subType", NewMeta())},
			),
			Entry("subType is not valid", NewRawObject(), "status", NewEmbeddedStatus("deviceEvent", "invalid"),
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("invalid", []string{"status"}), "/status/subType", NewMeta())},
			),
		)
	})
})
